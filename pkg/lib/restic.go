package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/TimothyYe/skm/internal/models"
	"github.com/TimothyYe/skm/internal/utils"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	cli "gopkg.in/urfave/cli.v1"
)

// ResticConfig is the persisted shape of ~/.skm/restic.json.
type ResticConfig struct {
	Repository   string `json:"repository"`
	PasswordFile string `json:"password_file"`
}

func MustHaveRestic(env *models.Environment) {
	if env.ResticPath == "" {
		utils.Fatalf("Restic not available. See https://restic.net/ for installation instructions.\n")
	}
}

func resticConfigPath(env *models.Environment) string {
	return filepath.Join(env.StorePath, "restic.json")
}

// LoadResticSettings reads the restic config from disk. It does NOT create
// the file on miss — that's the job of InitResticRepository, which only
// writes the config after `restic init` succeeds.
func LoadResticSettings(env *models.Environment) (*ResticConfig, error) {
	path := resticConfigPath(env)
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	cfg := &ResticConfig{}
	if err := json.NewDecoder(fp).Decode(cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if strings.TrimSpace(cfg.Repository) == "" || strings.TrimSpace(cfg.PasswordFile) == "" {
		return nil, fmt.Errorf("%s is missing repository or password_file", path)
	}
	return cfg, nil
}

// RequireInitializedResticRepo verifies the password file and repository
// look usable. Returns a non-nil error with actionable guidance if not.
func RequireInitializedResticRepo(cfg *ResticConfig) error {
	if _, err := os.Stat(cfg.PasswordFile); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("password file %s does not exist", cfg.PasswordFile)
		}
		return fmt.Errorf("check password file %s: %w", cfg.PasswordFile, err)
	}
	// For local repositories we can sanity-check the on-disk layout. For
	// remote backends (s3:, sftp:, b2:, …) trust restic to surface errors at
	// invocation time.
	if isLocalRepo(cfg.Repository) {
		if _, err := os.Stat(filepath.Join(cfg.Repository, "config")); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("restic repository at %s is not initialized", cfg.Repository)
			}
			return fmt.Errorf("check restic repository %s: %w", cfg.Repository, err)
		}
	}
	return nil
}

func isLocalRepo(repo string) bool {
	// restic URLs are scheme-prefixed (s3:, sftp:, b2:, azure:, gs:, rclone:, rest:).
	// Anything else (absolute or relative path) is treated as local.
	if i := strings.Index(repo, ":"); i > 0 {
		scheme := repo[:i]
		// On Windows a drive letter like "C:" could be confused with a scheme;
		// SKM doesn't currently target Windows so this is a non-issue, but we
		// keep the check explicit for clarity.
		if len(scheme) > 1 {
			return false
		}
	}
	return true
}

// InitResticRepository drives the interactive setup for a new restic backend:
// prompt for repository URL and password, write the password file, run
// `restic init`, and only on success persist restic.json.
func InitResticRepository(env *models.Environment) error {
	cfgPath := resticConfigPath(env)
	if _, err := os.Stat(cfgPath); err == nil {
		color.Red("%s%s already exists. Delete it to re-initialize, or edit it directly.", utils.CrossSymbol, cfgPath)
		return nil
	}

	color.Cyan("Configuring restic backup for SKM.")
	fmt.Println("Examples:")
	fmt.Println("  Local:  /Users/me/.skm-backups")
	fmt.Println("  S3:     s3:s3.amazonaws.com/my-bucket/skm")
	fmt.Println("  R2:     s3:https://<account>.r2.cloudflarestorage.com/my-bucket/skm")
	fmt.Println("  SFTP:   sftp:user@host:/data/skm")
	fmt.Println("  B2:     b2:my-bucket/skm")
	fmt.Println()
	fmt.Println("For S3, R2, and B2, set the relevant credential env vars (e.g.")
	fmt.Println("AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY) before running backup.")
	fmt.Println()

	defaultRepo := filepath.Join(os.Getenv("HOME"), ".skm-backups")
	repo, err := promptString("Restic repository", defaultRepo, func(in string) error {
		if strings.TrimSpace(in) == "" {
			return errors.New("repository cannot be empty")
		}
		return nil
	})
	if err != nil {
		return nil
	}
	repo = strings.TrimSpace(repo)

	defaultPwFile := filepath.Join(os.Getenv("HOME"), ".skm-backups.passwd")
	pwFile, err := promptString("Path to write the password file", defaultPwFile, func(in string) error {
		if strings.TrimSpace(in) == "" {
			return errors.New("password file path cannot be empty")
		}
		return nil
	})
	if err != nil {
		return nil
	}
	pwFile = strings.TrimSpace(pwFile)
	if _, statErr := os.Stat(pwFile); statErr == nil {
		color.Red("%spassword file %s already exists; refusing to overwrite", utils.CrossSymbol, pwFile)
		return nil
	}

	pw1, err := promptMasked("Restic password", func(in string) error {
		if len(in) < 8 {
			return errors.New("password must be at least 8 characters")
		}
		return nil
	})
	if err != nil {
		return nil
	}
	pw2, err := promptMasked("Confirm password", nil)
	if err != nil {
		return nil
	}
	if pw1 != pw2 {
		color.Red("%spasswords do not match", utils.CrossSymbol)
		return nil
	}

	if err := writeSecretFile(pwFile, []byte(pw1+"\n")); err != nil {
		color.Red("%swrite %s: %s", utils.CrossSymbol, pwFile, err.Error())
		return nil
	}

	// Run `restic init`. If the repo URL or credentials are wrong, restic
	// will print a clear error — propagate that and clean up the password
	// file so the user can re-run.
	if err := runResticInit(env.ResticPath, repo, pwFile); err != nil {
		color.Red("%srestic init failed: %s", utils.CrossSymbol, err.Error())
		_ = os.Remove(pwFile)
		return nil
	}

	cfg := ResticConfig{Repository: repo, PasswordFile: pwFile}
	if err := writeConfigFile(cfgPath, &cfg); err != nil {
		color.Red("%swrite %s: %s", utils.CrossSymbol, cfgPath, err.Error())
		return nil
	}

	color.Green("%sRestic repository initialized at %s", utils.CheckSymbol, repo)
	color.Yellow("")
	color.Yellow("⚠ IMPORTANT: store the restic password somewhere OTHER than this machine.")
	color.Yellow("  If this laptop is lost and the password only lives in %s,", pwFile)
	color.Yellow("  the backup will be unrecoverable. Write it down, save it in a password")
	color.Yellow("  manager, or print it — anywhere off-machine.")
	return nil
}

func runResticInit(resticBin, repo, passwordFile string) error {
	cmd := exec.Command(resticBin, "init", "--repo", repo, "--password-file", passwordFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func writeSecretFile(path string, data []byte) error {
	fp, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	if _, err := fp.Write(data); err != nil {
		fp.Close()
		return err
	}
	return fp.Close()
}

func writeConfigFile(path string, cfg *ResticConfig) error {
	fp, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer fp.Close()
	enc := json.NewEncoder(fp)
	enc.SetIndent("", "  ")
	return enc.Encode(cfg)
}

func promptString(label, def string, validate promptui.ValidateFunc) (string, error) {
	p := promptui.Prompt{Label: label, Default: def, Validate: validate, AllowEdit: true}
	return p.Run()
}

func promptMasked(label string, validate promptui.ValidateFunc) (string, error) {
	p := promptui.Prompt{Label: label, Mask: '*', Validate: validate}
	return p.Run()
}

// Deprecated shims preserved so existing callers still compile while
// callers migrate to the new helpers. They map onto the new API.
func MustLoadOrCreateResticSettings(env *models.Environment, _ *cli.Context) *ResticConfig {
	cfg, err := LoadResticSettings(env)
	if err == nil {
		return cfg
	}
	utils.Fatalf("No restic configuration found. Run `skm backup --restic --init` first.\n")
	return nil
}

func EnsureInitializedResticRepo(cfg *ResticConfig, _ *models.Environment) {
	if err := RequireInitializedResticRepo(cfg); err != nil {
		utils.Fatalf("%s. Run `skm backup --restic --init` to (re)configure.\n", err.Error())
	}
}
