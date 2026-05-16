package actions

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/TimothyYe/skm/internal/models"
	"github.com/TimothyYe/skm/internal/utils"
	"github.com/fatih/color"
	"gopkg.in/urfave/cli.v1"
)

// Import pulls an existing SSH key pair from anywhere on disk into the SKM
// store under a chosen alias. The user supplies the path to either the
// private or public key; the matching half is inferred. With --move, the
// original files are deleted after a successful copy.
func Import(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)

	if c.NArg() == 0 {
		color.Red("%sUsage: skm import [--alias <name>] [--move] <path-to-key-or-bundle>", utils.CrossSymbol)
		return nil
	}
	// urfave/cli v1 stops parsing flags at the first positional arg, so flags
	// like `--alias foo` placed *after* the path get captured as positionals
	// and silently ignored. Detect that and fail loudly with a hint.
	if stray := findStrayFlag(c.Args()); stray != "" {
		color.Red("%sFlag %q appears after the path; put all flags before the path", utils.CrossSymbol, stray)
		return nil
	}
	src := c.Args().Get(0)

	if isBundlePath(src) {
		return importBundle(c, env, src)
	}

	privatePath, publicPath, err := resolveKeyPair(src)
	if err != nil {
		color.Red("%s%s", utils.CrossSymbol, err.Error())
		return nil
	}

	kt, err := detectKeyType(publicPath)
	if err != nil {
		color.Red("%s%s", utils.CrossSymbol, err.Error())
		return nil
	}

	alias := strings.TrimSpace(c.String("alias"))
	if alias == "" && c.NArg() >= 2 {
		alias = c.Args().Get(1)
	}
	if alias == "" {
		// Default to the source file's base name (minus extension) so the user
		// gets a reasonable alias without forcing the flag.
		base := filepath.Base(privatePath)
		alias = strings.TrimSuffix(base, filepath.Ext(base))
		if alias == "" || alias == "id_rsa" || alias == "id_ed25519" {
			color.Red("%sCannot infer an alias from %s; pass --alias <name>", utils.CrossSymbol, privatePath)
			return nil
		}
	}
	if err := validateAlias(alias); err != nil {
		color.Red("%s%s", utils.CrossSymbol, err.Error())
		return nil
	}

	keys := utils.LoadSSHKeys(env)
	if _, exists := keys[alias]; exists {
		color.Red("%sSSH key alias [%s] already exists", utils.CrossSymbol, alias)
		return nil
	}

	dstDir := filepath.Join(env.StorePath, alias)
	if err := os.Mkdir(dstDir, 0700); err != nil {
		color.Red("%sFailed to create alias directory: %s", utils.CrossSymbol, err.Error())
		return nil
	}

	dstPrivate := filepath.Join(dstDir, kt.PrivateKey())
	dstPublic := filepath.Join(dstDir, kt.PublicKey())

	if err := copyFile(privatePath, dstPrivate, 0600); err != nil {
		_ = os.RemoveAll(dstDir)
		color.Red("%sFailed to copy private key: %s", utils.CrossSymbol, err.Error())
		return nil
	}
	if err := copyFile(publicPath, dstPublic, 0644); err != nil {
		_ = os.RemoveAll(dstDir)
		color.Red("%sFailed to copy public key: %s", utils.CrossSymbol, err.Error())
		return nil
	}

	if c.Bool("move") {
		// Only remove originals once both copies have succeeded, so a failure
		// halfway through doesn't lose the source.
		_ = os.Remove(privatePath)
		_ = os.Remove(publicPath)
	}

	color.Green("%sImported %s key as [%s]", utils.CheckSymbol, kt.Name, alias)
	return nil
}

// resolveKeyPair takes the path the user supplied and returns the private and
// public key paths. The user may point at either half; the other is inferred
// by adding or removing the .pub suffix.
func resolveKeyPair(src string) (string, string, error) {
	abs, err := filepath.Abs(src)
	if err != nil {
		return "", "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", "", fmt.Errorf("cannot read %s: %w", src, err)
	}
	if info.IsDir() {
		return "", "", fmt.Errorf("%s is a directory; pass the path to a key file", src)
	}

	var priv, pub string
	if strings.HasSuffix(abs, ".pub") {
		pub = abs
		priv = strings.TrimSuffix(abs, ".pub")
	} else {
		priv = abs
		pub = abs + ".pub"
	}
	if _, err := os.Stat(priv); err != nil {
		return "", "", fmt.Errorf("private key %s not found: %w", priv, err)
	}
	if _, err := os.Stat(pub); err != nil {
		return "", "", fmt.Errorf("public key %s not found: %w", pub, err)
	}
	return priv, pub, nil
}

// detectKeyType reads the first token of the public key file (e.g. "ssh-rsa",
// "ssh-ed25519") and maps it to one of the SupportedKeyTypes. Returns an
// error for key types SKM doesn't manage yet, like ecdsa or FIDO2 keys.
func detectKeyType(pubPath string) (*models.KeyType, error) {
	data, err := os.ReadFile(pubPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read public key: %w", err)
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return nil, fmt.Errorf("public key %s is empty or malformed", pubPath)
	}
	switch fields[0] {
	case "ssh-rsa":
		kt := models.SupportedKeyTypes["rsa"]
		return &kt, nil
	case "ssh-ed25519":
		kt := models.SupportedKeyTypes["ed25519"]
		return &kt, nil
	default:
		return nil, fmt.Errorf("unsupported key type %q (skm currently manages rsa and ed25519)", fields[0])
	}
}

// isBundlePath returns true when the source path looks like an archive
// produced by `skm export` — either a gzipped tarball or its encrypted
// .enc counterpart. Detection is by extension; the caller validates the
// contents when unpacking.
func isBundlePath(p string) bool {
	lower := strings.ToLower(p)
	return strings.HasSuffix(lower, ".tar.gz") ||
		strings.HasSuffix(lower, ".tgz") ||
		strings.HasSuffix(lower, ".tar.gz.enc") ||
		strings.HasSuffix(lower, ".tgz.enc")
}

// importBundle restores a single-alias archive produced by `skm export`.
// When the path ends in .enc, openssl is invoked to decrypt it first; the
// resulting plaintext lives in a temp file that's cleaned up before return.
// The archive's top-level directory becomes the alias, overridable with
// --alias. Refuses to overwrite an existing alias.
func importBundle(c *cli.Context, env *models.Environment, src string) error {
	abs, err := filepath.Abs(src)
	if err != nil {
		color.Red("%s%s", utils.CrossSymbol, err.Error())
		return nil
	}
	if _, err := os.Stat(abs); err != nil {
		color.Red("%scannot read %s: %s", utils.CrossSymbol, src, err.Error())
		return nil
	}

	archivePath := abs
	encrypted := strings.HasSuffix(strings.ToLower(abs), ".enc")
	if encrypted {
		if _, err := exec.LookPath("openssl"); err != nil {
			color.Red("%sopenssl not found in PATH; required to decrypt %s", utils.CrossSymbol, src)
			return nil
		}
		tmp, err := os.CreateTemp("", "skm-import-*.tar.gz")
		if err != nil {
			color.Red("%s%s", utils.CrossSymbol, err.Error())
			return nil
		}
		tmp.Close()
		defer os.Remove(tmp.Name())

		if ok := utils.Execute("", "openssl", "enc", "-d", "-aes-256-cbc", "-pbkdf2", "-in", abs, "-out", tmp.Name()); !ok {
			color.Red("%sFailed to decrypt %s", utils.CrossSymbol, src)
			return nil
		}
		archivePath = tmp.Name()
	}

	// For bundles we deliberately do NOT fall back to the second positional
	// here: urfave/cli v1 doesn't parse flags placed after positional args, so
	// `skm import bundle.tar.gz --alias foo` leaves --alias unparsed and the
	// trailing "--alias" string would land in Args().Get(1). The archive's
	// own top-level directory is a sane default — keep the override explicit
	// to --alias and reject anything that looks like a stray flag.
	overrideAlias := strings.TrimSpace(c.String("alias"))
	if overrideAlias != "" {
		if err := validateAlias(overrideAlias); err != nil {
			color.Red("%s%s", utils.CrossSymbol, err.Error())
			return nil
		}
	}

	alias, err := extractAliasArchive(archivePath, env.StorePath, overrideAlias, utils.LoadSSHKeys(env))
	if err != nil {
		color.Red("%s%s", utils.CrossSymbol, err.Error())
		return nil
	}

	if c.Bool("move") {
		_ = os.Remove(abs)
	}

	color.Green("%sImported bundle as [%s]", utils.CheckSymbol, alias)
	return nil
}

// extractAliasArchive reads a tar.gz produced by `skm export` and writes its
// contents under storePath. The archive's top-level directory is renamed to
// overrideAlias when supplied. Returns the alias actually written. Validates
// that the archive contains exactly one top-level directory, that the alias
// does not collide with an existing key, and rewrites file permissions to
// 600 for private keys and 644 for public keys regardless of how they were
// stored on the sending side.
func extractAliasArchive(archivePath, storePath, overrideAlias string, existing map[string]*models.SSHKey) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("not a valid gzip archive: %w", err)
	}
	defer gz.Close()

	// First pass: read headers to determine the top-level directory and
	// validate the archive shape before touching the store.
	headers, err := readTarHeaders(tar.NewReader(gz))
	if err != nil {
		return "", err
	}
	srcAlias, err := singleTopLevelDir(headers)
	if err != nil {
		return "", err
	}
	alias := srcAlias
	if overrideAlias != "" {
		alias = overrideAlias
	}
	if _, exists := existing[alias]; exists {
		return "", fmt.Errorf("SSH key alias [%s] already exists; pass --alias to choose another name", alias)
	}

	// Reopen and stream entries into place under the chosen alias.
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	gz2, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gz2.Close()
	tr := tar.NewReader(gz2)

	dstDir := filepath.Join(storePath, alias)
	if err := os.Mkdir(dstDir, 0700); err != nil {
		return "", fmt.Errorf("cannot create alias directory: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(dstDir) }

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			cleanup()
			return "", err
		}
		rel, ok := stripTopDir(hdr.Name, srcAlias)
		if !ok {
			cleanup()
			return "", fmt.Errorf("archive contains entry outside alias directory: %s", hdr.Name)
		}
		if rel == "" {
			continue
		}
		// Guard against absolute paths or traversal attempts in malicious archives.
		clean := filepath.Clean(rel)
		if strings.HasPrefix(clean, "..") || filepath.IsAbs(clean) {
			cleanup()
			return "", fmt.Errorf("archive contains unsafe path: %s", hdr.Name)
		}
		target := filepath.Join(dstDir, clean)
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0700); err != nil {
				cleanup()
				return "", err
			}
		case tar.TypeReg:
			mode := perms(clean, hdr.Mode)
			if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
				cleanup()
				return "", err
			}
			w, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
			if err != nil {
				cleanup()
				return "", err
			}
			if _, err := io.Copy(w, tr); err != nil {
				w.Close()
				cleanup()
				return "", err
			}
			w.Close()
			if err := os.Chmod(target, mode); err != nil {
				cleanup()
				return "", err
			}
		default:
			// Skip symlinks, devices, etc. — SKM bundles only ever contain
			// regular files and directories.
		}
	}

	// Validate the unpacked alias has at least one supported key pair.
	if _, err := os.Stat(filepath.Join(dstDir)); err != nil {
		cleanup()
		return "", err
	}
	loaded := utils.LoadSSHKeys(&models.Environment{StorePath: storePath, SSHPath: ""})
	if _, ok := loaded[alias]; !ok {
		cleanup()
		return "", fmt.Errorf("archive did not contain a recognised SSH key pair")
	}
	return alias, nil
}

func readTarHeaders(tr *tar.Reader) ([]*tar.Header, error) {
	var headers []*tar.Header
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		headers = append(headers, hdr)
	}
	return headers, nil
}

// singleTopLevelDir returns the single top-level directory name shared by
// every header in the archive. Errors if the archive is empty or contains
// multiple top-level entries, which would indicate it wasn't built by `skm
// export`.
func singleTopLevelDir(headers []*tar.Header) (string, error) {
	if len(headers) == 0 {
		return "", fmt.Errorf("archive is empty")
	}
	top := ""
	for _, h := range headers {
		name := strings.TrimPrefix(h.Name, "./")
		if name == "" {
			continue
		}
		first, _, _ := strings.Cut(name, "/")
		if top == "" {
			top = first
		} else if top != first {
			return "", fmt.Errorf("archive contains multiple top-level entries (%q and %q); not an skm export bundle", top, first)
		}
	}
	if top == "" {
		return "", fmt.Errorf("could not determine alias from archive")
	}
	return top, nil
}

// stripTopDir removes the leading "<top>/" segment from name. Returns false
// when the entry isn't under that directory (and so doesn't belong in the
// alias we're extracting).
func stripTopDir(name, top string) (string, bool) {
	name = strings.TrimPrefix(name, "./")
	if name == top {
		return "", true
	}
	prefix := top + "/"
	if !strings.HasPrefix(name, prefix) {
		return "", false
	}
	return name[len(prefix):], true
}

// perms returns the file mode to use when writing an extracted entry. Private
// keys always land at 0600 and public keys at 0644, regardless of how the
// sending machine had them — this normalises any drift from manual editing.
func perms(rel string, archiveMode int64) os.FileMode {
	if strings.HasSuffix(rel, ".pub") {
		return 0644
	}
	if _, ok := models.SupportedKeyTypes.GetByFilename(filepath.Base(rel)); ok {
		return 0600
	}
	// Hook scripts and anything else: respect the archive's bits but mask off
	// world/group write to avoid surprises.
	m := os.FileMode(archiveMode) & 0755
	if m == 0 {
		m = 0644
	}
	return m
}

// findStrayFlag scans positional args (skipping the first, which is the path)
// and returns the first one that begins with "-". urfave/cli v1 doesn't
// re-enter flag parsing after the first positional, so anything looking like
// a flag here is something the user expected to be parsed but wasn't.
func findStrayFlag(args cli.Args) string {
	for i := 1; i < len(args); i++ {
		a := args.Get(i)
		if strings.HasPrefix(a, "-") {
			return a
		}
	}
	return ""
}

// validateAlias rejects alias names that are empty, contain path separators
// or whitespace, or start with "-". The leading-dash check catches a common
// pitfall: urfave/cli v1 does not parse flags placed after positional args,
// so a stray `--alias` or `--move` after the path is captured as a positional
// and would otherwise be silently used as the alias name.
func validateAlias(alias string) error {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return fmt.Errorf("alias cannot be empty")
	}
	if strings.HasPrefix(alias, "-") {
		return fmt.Errorf("alias %q looks like a flag; put --alias/--move before the path argument", alias)
	}
	if strings.ContainsAny(alias, "/\\ \t") {
		return fmt.Errorf("alias cannot contain spaces or path separators")
	}
	return nil
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Chmod(mode)
}
