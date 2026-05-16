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
	"time"

	"github.com/TimothyYe/skm/internal/utils"
	"github.com/fatih/color"
	"gopkg.in/urfave/cli.v1"
)

// Export bundles a single alias directory (private key, public key, and any
// hook file) into a gzipped tar archive. With --encrypt, the archive is piped
// through `openssl enc -aes-256-cbc -pbkdf2 -salt`, which prompts the user
// for a passphrase. The resulting bundle can be moved to another machine and
// expanded under that machine's SKM store path.
func Export(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)

	if c.NArg() == 0 {
		color.Red("%sUsage: skm export [-o file] [--encrypt] <alias>", utils.CrossSymbol)
		return nil
	}
	alias := c.Args().Get(0)

	keys := utils.LoadSSHKeys(env)
	if _, ok := keys[alias]; !ok {
		color.Red("%sSSH key alias [%s] not found", utils.CrossSymbol, alias)
		return nil
	}

	encrypt := c.Bool("encrypt")
	out := strings.TrimSpace(c.String("output"))
	if out == "" {
		ts := time.Now().Format("20060102150405")
		name := fmt.Sprintf("skm-%s-%s.tar.gz", alias, ts)
		if encrypt {
			name += ".enc"
		}
		out = filepath.Join(os.Getenv("HOME"), name)
	}
	if _, err := os.Stat(out); err == nil {
		color.Red("%sOutput file %s already exists", utils.CrossSymbol, out)
		return nil
	}

	tarPath := out
	if encrypt {
		// Write the plaintext archive to a temp file first, then encrypt it
		// into the final destination. This keeps openssl's stdin/stdout free
		// for the passphrase prompt.
		tmp, err := os.CreateTemp("", "skm-export-*.tar.gz")
		if err != nil {
			color.Red("%s%s", utils.CrossSymbol, err.Error())
			return nil
		}
		tmp.Close()
		tarPath = tmp.Name()
		defer os.Remove(tarPath)
	}

	if err := writeAliasArchive(env.StorePath, alias, tarPath); err != nil {
		color.Red("%sFailed to build archive: %s", utils.CrossSymbol, err.Error())
		return nil
	}

	if encrypt {
		if _, err := exec.LookPath("openssl"); err != nil {
			color.Red("%sopenssl not found in PATH; install it or omit --encrypt", utils.CrossSymbol)
			return nil
		}
		if ok := utils.Execute("", "openssl", "enc", "-aes-256-cbc", "-pbkdf2", "-salt", "-in", tarPath, "-out", out); !ok {
			// Clean up the partial output so the user can re-run.
			_ = os.Remove(out)
			return nil
		}
	}

	color.Green("%sExported [%s] to %s", utils.CheckSymbol, alias, out)
	if encrypt {
		color.Yellow("Decrypt with: openssl enc -d -aes-256-cbc -pbkdf2 -in %s -out %s.tar.gz", out, alias)
	}
	return nil
}

func writeAliasArchive(storePath, alias, dst string) error {
	srcDir := filepath.Join(storePath, alias)
	info, err := os.Stat(srcDir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", srcDir)
	}

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer out.Close()
	gz := gzip.NewWriter(out)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()

	return filepath.Walk(srcDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(filepath.Dir(srcDir), path)
		if err != nil {
			return err
		}
		hdr, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}
		// filepath.Rel uses the OS separator; tar headers want forward slashes.
		hdr.Name = filepath.ToSlash(rel)
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(tw, f)
		return err
	})
}
