package actions

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/TimothyYe/skm/internal/utils"
	"github.com/TimothyYe/skm/pkg/lib"
	"github.com/fatih/color"
	"gopkg.in/urfave/cli.v1"
)

func Restore(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)
	if c.Bool("restic") {
		lib.MustHaveRestic(env)
		resticCfg := lib.MustLoadOrCreateResticSettings(env, c)
		lib.EnsureInitializedResticRepo(resticCfg, env)
		if c.String("restic-snapshot") == "" {
			fmt.Fprintf(os.Stderr, "No snapshot specified. The following snapshots are available:\n\n")
			utils.Execute(env.StorePath, env.ResticPath, "snapshots", "--password-file", resticCfg.PasswordFile, "--repo", resticCfg.Repository)
			fmt.Fprintln(os.Stderr, "")
			utils.Fatalf("Please specify a snapshot\n")
		}
		result := utils.Execute(env.StorePath, env.ResticPath, "restore", c.String("restic-snapshot"), "--target", env.StorePath, "--password-file", resticCfg.PasswordFile, "--repo", resticCfg.Repository)
		if result {
			color.Green("%s Backup restored to %s", utils.CheckSymbol, env.StorePath)
		}
		return nil
	}
	var filePath string

	if c.NArg() > 0 {
		filePath = c.Args().Get(0)
	} else {
		color.Red("%sPlease input the corrent backup file path!", utils.CrossSymbol)
		return nil
	}

	var confirm string
	color.Green("This operation will overwrite all you current SSH keys, please make sure you want to do this operation?")
	fmt.Print("(Y/n):")
	fmt.Scan(&confirm)

	if confirm != "Y" {
		os.Exit(0)
	}

	// Clear the key store first
	err := os.RemoveAll(env.StorePath)

	if err != nil {
		fmt.Println("Clear store path failed:", err.Error())
	}

	// Extract backup file
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		color.Red("%sFailed to get the file path!", utils.CrossSymbol)
	}

	// Clear all keys
	utils.ClearKey(env)
	err = os.Mkdir(env.StorePath, 0755)
	if err != nil {
		color.Red("%sFailed to initialize SSH key store!", utils.CrossSymbol)
		return nil
	}

	result := utils.Execute(env.StorePath, "tar", "zxvf", absPath, "-C", env.StorePath)
	if result {
		fmt.Println()
		color.Green("%s All SSH keys restored to %s", utils.CheckSymbol, env.StorePath)
	}

	return nil
}
