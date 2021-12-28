package actions

import (
	"github.com/TimothyYe/skm/internal/utils"
	"github.com/TimothyYe/skm/pkg/lib"
	"github.com/fatih/color"
	"gopkg.in/urfave/cli.v1"
	"os"
	"path/filepath"
)

func Backup(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)
	if c.Bool("restic") {
		lib.MustHaveRestic(env)
		resticCfg := lib.MustLoadOrCreateResticSettings(env, c)
		lib.EnsureInitializedResticRepo(resticCfg, env)
		result := utils.Execute(env.StorePath, env.ResticPath, "backup", ".", "--password-file", resticCfg.PasswordFile, "--repo", resticCfg.Repository)
		if result {
			color.Green("%s Backup to %s complete", utils.CheckSymbol, resticCfg.Repository)
		}
		return nil
	}
	fileName := utils.GetBakFileName()
	dstFile := filepath.Join(os.Getenv("HOME"), fileName)

	result := utils.Execute(env.StorePath, "tar", "-czvf", dstFile, ".")
	if result {
		color.Green("%s All SSH keys backup to: %s", utils.CheckSymbol, dstFile)
	}

	return nil
}
