package actions

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/TimothyYe/skm/internal/models"
	"github.com/TimothyYe/skm/internal/utils"
	"github.com/fatih/color"
	cli "gopkg.in/urfave/cli.v1"
)

type checkLevel string

const (
	levelOK   checkLevel = "ok"
	levelWarn checkLevel = "warn"
	levelFail checkLevel = "fail"
)

type checkResult struct {
	Name    string     `json:"name"`
	Level   checkLevel `json:"level"`
	Message string     `json:"message"`
	Hint    string     `json:"hint,omitempty"`
}

// Doctor runs diagnostic checks against the SKM environment, the SSH agent,
// and the keys in the store. Exits non-zero when any check fails.
func Doctor(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)
	keyMap := utils.LoadSSHKeys(env)

	results := []checkResult{}
	results = append(results, checkBinaries()...)
	results = append(results, checkAgent())
	results = append(results, checkStorePath(env))
	results = append(results, checkSSHPath(env))
	results = append(results, checkDefaultKey(keyMap))
	results = append(results, checkKeys(env, keyMap)...)
	results = append(results, checkHooks(env, keyMap)...)

	if c.Bool("json") {
		return printDoctorJSON(results)
	}

	printDoctorText(results)

	for _, r := range results {
		if r.Level == levelFail {
			return cli.NewExitError("", 1)
		}
	}
	return nil
}

func checkBinaries() []checkResult {
	bins := []string{"ssh-keygen", "ssh-add", "ssh-copy-id"}
	out := make([]checkResult, 0, len(bins))
	for _, b := range bins {
		if _, err := exec.LookPath(b); err != nil {
			out = append(out, checkResult{
				Name:    "binary:" + b,
				Level:   levelFail,
				Message: fmt.Sprintf("%s not found in PATH", b),
				Hint:    "Install OpenSSH client tools",
			})
			continue
		}
		out = append(out, checkResult{
			Name:    "binary:" + b,
			Level:   levelOK,
			Message: b + " available",
		})
	}
	return out
}

func checkAgent() checkResult {
	if os.Getenv("SSH_AUTH_SOCK") == "" {
		return checkResult{
			Name:    "ssh-agent",
			Level:   levelWarn,
			Message: "SSH_AUTH_SOCK is not set; ssh-agent is likely not running",
			Hint:    "Start ssh-agent (e.g. `eval $(ssh-agent)`) or rely on your OS keychain",
		}
	}
	// `ssh-add -l` exits 0 with identities, 1 with none, 2 if agent unreachable.
	cmd := exec.Command("ssh-add", "-l")
	cmd.Env = append(os.Environ(), "LC_ALL=C")
	err := cmd.Run()
	if err == nil {
		return checkResult{Name: "ssh-agent", Level: levelOK, Message: "ssh-agent reachable with identities loaded"}
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		switch exitErr.ExitCode() {
		case 1:
			return checkResult{Name: "ssh-agent", Level: levelOK, Message: "ssh-agent reachable (no identities loaded)"}
		case 2:
			return checkResult{
				Name:    "ssh-agent",
				Level:   levelFail,
				Message: "ssh-agent is not reachable via SSH_AUTH_SOCK",
				Hint:    "Start ssh-agent (e.g. `eval $(ssh-agent)`)",
			}
		}
	}
	return checkResult{Name: "ssh-agent", Level: levelWarn, Message: "ssh-add returned unexpected error: " + err.Error()}
}

func checkStorePath(env *models.Environment) checkResult {
	return checkDirWritable("store-path", env.StorePath)
}

func checkSSHPath(env *models.Environment) checkResult {
	return checkDirWritable("ssh-path", env.SSHPath)
}

func checkDirWritable(name, path string) checkResult {
	info, err := os.Stat(path)
	if err != nil {
		return checkResult{Name: name, Level: levelFail, Message: fmt.Sprintf("cannot stat %s: %v", path, err)}
	}
	if !info.IsDir() {
		return checkResult{Name: name, Level: levelFail, Message: fmt.Sprintf("%s is not a directory", path)}
	}
	probe, err := os.CreateTemp(path, ".skm-doctor-*")
	if err != nil {
		return checkResult{
			Name:    name,
			Level:   levelFail,
			Message: fmt.Sprintf("%s is not writable: %v", path, err),
			Hint:    fmt.Sprintf("Check permissions on %s", path),
		}
	}
	probe.Close()
	os.Remove(probe.Name())
	return checkResult{Name: name, Level: levelOK, Message: fmt.Sprintf("%s writable (mode %s)", path, info.Mode().Perm())}
}

func checkDefaultKey(keyMap map[string]*models.SSHKey) checkResult {
	if len(keyMap) == 0 {
		return checkResult{
			Name:    "default-key",
			Level:   levelWarn,
			Message: "no SSH keys in store",
			Hint:    "Run `skm create <alias>` to add one",
		}
	}
	for alias, key := range keyMap {
		if !key.IsDefault {
			continue
		}
		if _, err := os.Stat(key.PrivateKey); err != nil {
			return checkResult{Name: "default-key", Level: levelFail, Message: fmt.Sprintf("default key [%s] private file missing: %v", alias, err)}
		}
		if _, err := os.Stat(key.PublicKey); err != nil {
			return checkResult{Name: "default-key", Level: levelFail, Message: fmt.Sprintf("default key [%s] public file missing: %v", alias, err)}
		}
		return checkResult{Name: "default-key", Level: levelOK, Message: fmt.Sprintf("default key [%s] resolves to %s", alias, key.PrivateKey)}
	}
	return checkResult{
		Name:    "default-key",
		Level:   levelWarn,
		Message: "no default key selected",
		Hint:    "Run `skm use <alias>` to set one",
	}
}

func checkKeys(env *models.Environment, keyMap map[string]*models.SSHKey) []checkResult {
	if len(keyMap) == 0 {
		return nil
	}
	results := []checkResult{}
	for alias, key := range keyMap {
		results = append(results, checkKeyPermissions(alias, key)...)
		results = append(results, checkKeyStrength(alias, key))
	}
	_ = env
	return results
}

func checkKeyPermissions(alias string, key *models.SSHKey) []checkResult {
	results := []checkResult{}
	if info, err := os.Stat(key.PrivateKey); err == nil {
		mode := info.Mode().Perm()
		// Private keys must not be group/world readable. 0600 is canonical; 0400
		// is also fine (read-only owner).
		if mode&0077 != 0 {
			results = append(results, checkResult{
				Name:    "perm:" + alias + ":private",
				Level:   levelFail,
				Message: fmt.Sprintf("private key %s has loose permissions %s (want 600)", key.PrivateKey, mode),
				Hint:    fmt.Sprintf("chmod 600 %s", key.PrivateKey),
			})
		} else {
			results = append(results, checkResult{
				Name:    "perm:" + alias + ":private",
				Level:   levelOK,
				Message: fmt.Sprintf("private key %s mode %s", filepath.Base(key.PrivateKey), mode),
			})
		}
	}
	if info, err := os.Stat(key.PublicKey); err == nil {
		mode := info.Mode().Perm()
		// Public keys must not be world-writable. 0644 is canonical.
		if mode&0022 != 0 {
			results = append(results, checkResult{
				Name:    "perm:" + alias + ":public",
				Level:   levelWarn,
				Message: fmt.Sprintf("public key %s is group/world writable (mode %s)", key.PublicKey, mode),
				Hint:    fmt.Sprintf("chmod 644 %s", key.PublicKey),
			})
		}
	}
	return results
}

func checkKeyStrength(alias string, key *models.SSHKey) checkResult {
	details := inspectKey(key.PublicKey)
	if details.Fingerprint == "" {
		return checkResult{Name: "strength:" + alias, Level: levelWarn, Message: fmt.Sprintf("could not inspect key [%s]", alias)}
	}
	keyType := ""
	if key.Type != nil {
		keyType = key.Type.Name
	}
	if keyType == "rsa" {
		bits, _ := strconv.Atoi(details.Bits)
		if bits > 0 && bits < 3072 {
			return checkResult{
				Name:    "strength:" + alias,
				Level:   levelWarn,
				Message: fmt.Sprintf("key [%s] is RSA-%d (weak; want >= 3072)", alias, bits),
				Hint:    "Replace with `skm create <alias> -t ed25519` or RSA >= 3072",
			}
		}
	}
	return checkResult{Name: "strength:" + alias, Level: levelOK, Message: fmt.Sprintf("key [%s] %s/%s", alias, keyType, details.Bits)}
}

func checkHooks(env *models.Environment, keyMap map[string]*models.SSHKey) []checkResult {
	results := []checkResult{}
	for alias := range keyMap {
		hookPath := filepath.Join(env.StorePath, alias, utils.HookName)
		info, err := os.Stat(hookPath)
		if err != nil {
			continue
		}
		if info.Mode()&0111 == 0 {
			results = append(results, checkResult{
				Name:    "hook:" + alias,
				Level:   levelWarn,
				Message: fmt.Sprintf("hook for [%s] is not executable (mode %s)", alias, info.Mode().Perm()),
				Hint:    fmt.Sprintf("chmod +x %s", hookPath),
			})
			continue
		}
		results = append(results, checkResult{Name: "hook:" + alias, Level: levelOK, Message: fmt.Sprintf("hook for [%s] is executable", alias)})
	}
	return results
}

func printDoctorText(results []checkResult) {
	fmt.Println()
	var okN, warnN, failN int
	for _, r := range results {
		switch r.Level {
		case levelOK:
			color.Green("  %s %s", utils.CheckSymbol, r.Message)
			okN++
		case levelWarn:
			color.Yellow("  ! %s", r.Message)
			if r.Hint != "" {
				color.Yellow("      hint: %s", r.Hint)
			}
			warnN++
		case levelFail:
			color.Red("  %s%s", utils.CrossSymbol, r.Message)
			if r.Hint != "" {
				color.Red("      hint: %s", r.Hint)
			}
			failN++
		}
	}
	fmt.Println()
	summary := fmt.Sprintf("%d passed, %d warning(s), %d failure(s)", okN, warnN, failN)
	switch {
	case failN > 0:
		color.Red(summary)
	case warnN > 0:
		color.Yellow(summary)
	default:
		color.Green(summary)
	}
}

func printDoctorJSON(results []checkResult) error {
	out, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	for _, r := range results {
		if r.Level == levelFail {
			return cli.NewExitError("", 1)
		}
	}
	return nil
}

