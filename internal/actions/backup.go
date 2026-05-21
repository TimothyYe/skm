package actions

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/TimothyYe/skm/internal/models"
	"github.com/TimothyYe/skm/internal/utils"
	"github.com/TimothyYe/skm/pkg/lib"
	"github.com/fatih/color"
	"gopkg.in/urfave/cli.v1"
)

func Backup(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)
	if c.Bool("restic") {
		return resticBackup(c, env)
	}

	encrypt := c.Bool("encrypt")
	fileName := utils.GetBakFileName()
	if encrypt {
		fileName += ".enc"
	}
	dstFile := filepath.Join(os.Getenv("HOME"), fileName)

	tarPath := dstFile
	if encrypt {
		tmp, err := os.CreateTemp("", "skm-backup-*.tar.gz")
		if err != nil {
			color.Red("%s%s", utils.CrossSymbol, err.Error())
			return nil
		}
		tmp.Close()
		tarPath = tmp.Name()
		defer os.Remove(tarPath)
	}

	if ok := utils.Execute(env.StorePath, "tar", "--exclude=./"+utils.TrashDir, "-czvf", tarPath, "."); !ok {
		return nil
	}

	if encrypt {
		if _, err := exec.LookPath("openssl"); err != nil {
			color.Red("%sopenssl not found in PATH; install it or omit --encrypt", utils.CrossSymbol)
			return nil
		}
		args := []string{"enc", "-aes-256-cbc", "-pbkdf2", "-salt", "-in", tarPath, "-out", dstFile}
		if pf := strings.TrimSpace(c.String("password-file")); pf != "" {
			args = append(args, "-pass", "file:"+pf)
		}
		if ok := utils.Execute("", "openssl", args...); !ok {
			_ = os.Remove(dstFile)
			return nil
		}
		color.Green("%s All SSH keys backup to: %s", utils.CheckSymbol, dstFile)
		color.Yellow("  Decrypt with: openssl enc -d -aes-256-cbc -pbkdf2 -in %s -out %s", dstFile, strings.TrimSuffix(filepath.Base(dstFile), ".enc"))
		return nil
	}

	color.Green("%s All SSH keys backup to: %s", utils.CheckSymbol, dstFile)
	color.Yellow("⚠ This bundle contains UNENCRYPTED private keys. If it leaves this")
	color.Yellow("  machine, anyone with the file can use your keys. Re-run with")
	color.Yellow("  --encrypt to produce an encrypted archive.")
	return nil
}

func resticBackup(c *cli.Context, env *models.Environment) error {
	lib.MustHaveRestic(env)

	if c.Bool("init") {
		return lib.InitResticRepository(env)
	}

	resticCfg, err := lib.LoadResticSettings(env)
	if err != nil {
		color.Red("%sNo restic configuration yet. Run `skm backup --restic --init` to set one up.", utils.CrossSymbol)
		return nil
	}
	if err := lib.RequireInitializedResticRepo(resticCfg); err != nil {
		color.Red("%s%s", utils.CrossSymbol, err.Error())
		color.Yellow("Hint: run `skm backup --restic --init` to (re)configure the repository.")
		return nil
	}

	result := utils.Execute(env.StorePath, env.ResticPath, "backup", ".", "--exclude", utils.TrashDir, "--password-file", resticCfg.PasswordFile, "--repo", resticCfg.Repository)
	if result {
		color.Green("%s Backup to %s complete", utils.CheckSymbol, resticCfg.Repository)
	}
	return nil
}
