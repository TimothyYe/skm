package actions

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/TimothyYe/skm/internal/models"
	"github.com/TimothyYe/skm/internal/utils"
	"github.com/fatih/color"
	"gopkg.in/urfave/cli.v1"
)

const (
	defaultKeyType = "ed25519"
	minRSABits     = 3072
)

type createParams struct {
	Alias   string
	Type    string
	Bits    int // 0 when not applicable
	Comment string
}

func Create(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)

	if c.NArg() == 0 {
		color.Red("%sPlease input key alias name!", utils.CrossSymbol)
		return nil
	}

	params, err := parseCreateParams(c)
	if err != nil {
		color.Red("%s%s", utils.CrossSymbol, err.Error())
		return nil
	}

	// All filesystem checks before any mutation, so an invalid request never
	// leaves an orphan alias directory behind.
	keyMap := utils.LoadSSHKeys(env)
	if _, ok := keyMap[params.Alias]; ok {
		color.Red("%sSSH key alias [%s] already exists!", utils.CrossSymbol, params.Alias)
		return nil
	}
	aliasDir := filepath.Join(env.StorePath, params.Alias)
	createdDir, err := prepareAliasDir(aliasDir)
	if err != nil {
		color.Red("%s%s", utils.CrossSymbol, err.Error())
		return nil
	}

	args := buildKeygenArgs(params, env.StorePath)
	if !utils.Execute("", "ssh-keygen", args...) {
		color.Red("%sCreate SSH key failed!", utils.CrossSymbol)
		// Only clean up the directory if WE created it. Don't wipe an
		// existing empty dir the user might have set up themselves.
		if createdDir {
			_ = os.Remove(aliasDir)
		}
		return nil
	}

	color.Green("%sSSH key [%s] created!", utils.CheckSymbol, params.Alias)
	_ = utils.RunHook(utils.EventPostCreate, params.Alias, env)
	return nil
}

// parseCreateParams reads and validates CLI inputs without touching the
// filesystem. Returned errors carry messages safe to print to the user.
func parseCreateParams(c *cli.Context) (createParams, error) {
	alias := strings.TrimSpace(c.Args().Get(0))
	if err := validateAlias(alias); err != nil {
		return createParams{}, err
	}

	keyType := strings.ToLower(strings.TrimSpace(c.String("t")))
	if keyType == "" {
		keyType = defaultKeyType
	}
	kt, ok := models.SupportedKeyTypes[keyType]
	if !ok {
		return createParams{}, fmt.Errorf("unsupported key type %q (supported: %s)", keyType, supportedKeyTypeList())
	}

	bits := 0
	if kt.SupportsVariableBitsize {
		raw := strings.TrimSpace(c.String("b"))
		if raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil {
				return createParams{}, fmt.Errorf("invalid -b value %q: must be an integer", raw)
			}
			bits = parsed
		}
		if keyType == "rsa" {
			if bits == 0 {
				bits = minRSABits
			}
			if bits < minRSABits {
				return createParams{}, fmt.Errorf("RSA keys below %d bits are not allowed (got %d); `skm audit` flags these as weak", minRSABits, bits)
			}
		}
	}

	return createParams{
		Alias:   alias,
		Type:    keyType,
		Bits:    bits,
		Comment: strings.TrimSpace(c.String("C")),
	}, nil
}

// buildKeygenArgs returns the ssh-keygen argv for the given (already
// validated) parameters. Pure function — no filesystem access — so it can
// be unit-tested without invoking ssh-keygen.
func buildKeygenArgs(p createParams, storePath string) []string {
	kt := models.SupportedKeyTypes[p.Type]
	args := []string{"-t", p.Type, "-f", filepath.Join(storePath, p.Alias, kt.KeyBaseName)}
	if p.Bits > 0 {
		args = append(args, "-b", strconv.Itoa(p.Bits))
	}
	if p.Comment != "" {
		args = append(args, "-C", p.Comment)
	}
	return args
}

// prepareAliasDir ensures the alias directory exists and is empty, ready to
// receive the new key files. Returns true if the directory was created (so
// the caller can clean it up on failure), false if it already existed empty.
func prepareAliasDir(dir string) (bool, error) {
	info, err := os.Stat(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			return false, fmt.Errorf("stat %s: %w", dir, err)
		}
		if err := os.Mkdir(dir, 0700); err != nil {
			return false, fmt.Errorf("create alias directory: %w", err)
		}
		return true, nil
	}
	if !info.IsDir() {
		return false, fmt.Errorf("%s already exists and is not a directory", dir)
	}
	empty, err := utils.IsEmpty(dir)
	if err != nil {
		return false, fmt.Errorf("check %s: %w", dir, err)
	}
	if !empty {
		return false, fmt.Errorf("%s already exists and is not empty", dir)
	}
	return false, nil
}

func supportedKeyTypeList() string {
	names := make([]string, 0, len(models.SupportedKeyTypes))
	for name := range models.SupportedKeyTypes {
		names = append(names, name)
	}
	// Sort for stable error messages.
	for i := 1; i < len(names); i++ {
		for j := i; j > 0 && names[j] < names[j-1]; j-- {
			names[j], names[j-1] = names[j-1], names[j]
		}
	}
	return strings.Join(names, ", ")
}
