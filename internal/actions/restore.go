package actions

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/TimothyYe/skm/internal/utils"
	"github.com/TimothyYe/skm/pkg/lib"
	"github.com/fatih/color"
	"gopkg.in/urfave/cli.v1"
)

func Restore(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)
	if c.Bool("restic") {
		lib.MustHaveRestic(env)
		resticCfg, err := lib.LoadResticSettings(env)
		if err != nil {
			color.Red("%sNo restic configuration yet. Run `skm backup --restic --init` to set one up.", utils.CrossSymbol)
			return nil
		}
		if err := lib.RequireInitializedResticRepo(resticCfg); err != nil {
			color.Red("%s%s", utils.CrossSymbol, err.Error())
			return nil
		}
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

	if c.NArg() == 0 {
		color.Red("%sPlease input the correct backup file path!", utils.CrossSymbol)
		return nil
	}
	filePath := c.Args().Get(0)

	var confirm string
	color.Green("This operation will overwrite all you current SSH keys, please make sure you want to do this operation?")
	fmt.Print("(Y/n):")
	fmt.Scan(&confirm)

	if confirm != "Y" {
		os.Exit(0)
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		color.Red("%sFailed to get the file path!", utils.CrossSymbol)
		return nil
	}

	// If the bundle is encrypted (`.enc` suffix), decrypt it into a temp file
	// first, then run the regular tar-extract flow against that.
	tarPath := absPath
	if strings.HasSuffix(strings.ToLower(absPath), ".enc") {
		if _, err := exec.LookPath("openssl"); err != nil {
			color.Red("%sopenssl not found in PATH; required to decrypt %s", utils.CrossSymbol, absPath)
			return nil
		}
		tmp, err := os.CreateTemp("", "skm-restore-*.tar.gz")
		if err != nil {
			color.Red("%s%s", utils.CrossSymbol, err.Error())
			return nil
		}
		tmp.Close()
		defer os.Remove(tmp.Name())
		tarPath = tmp.Name()

		args := []string{"enc", "-d", "-aes-256-cbc", "-pbkdf2", "-in", absPath, "-out", tarPath}
		if pf := strings.TrimSpace(c.String("password-file")); pf != "" {
			args = append(args, "-pass", "file:"+pf)
		}
		if ok := utils.Execute("", "openssl", args...); !ok {
			color.Red("%sFailed to decrypt %s", utils.CrossSymbol, absPath)
			return nil
		}
	}

	// Clear the key store first
	if err := os.RemoveAll(env.StorePath); err != nil {
		color.Red("%sClear store path failed: %s", utils.CrossSymbol, err.Error())
	}

	if err := utils.ClearKey(env); err != nil {
		color.Red("%s%s", utils.CrossSymbol, err.Error())
		return nil
	}
	if err := os.Mkdir(env.StorePath, 0700); err != nil {
		color.Red("%sFailed to initialize SSH key store!", utils.CrossSymbol)
		return nil
	}

	if utils.Execute(env.StorePath, "tar", "zxvf", tarPath, "-C", env.StorePath) {
		fmt.Println()
		color.Green("%s All SSH keys restored to %s", utils.CheckSymbol, env.StorePath)
	}
	return nil
}
