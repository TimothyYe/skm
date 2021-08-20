package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/TimothyYe/skm"
	"github.com/fatih/color"

	cli "gopkg.in/urfave/cli.v1"
)
var home,_ = os.UserHomeDir();
var defaultStorePath = filepath.Join(home, ".skm")
var defaultSSHPath = filepath.Join(home, ".ssh")

func init() {
	// initialize the store path
	if envStorePath := os.Getenv("SKM_STORE_PATH"); envStorePath != "" {
		defaultStorePath = envStorePath
	}
	if d, err := os.Readlink(defaultStorePath); err == nil {
		defaultStorePath = d
	}
}

func main() {
	parseArgs()

	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "store-path",
			Value: defaultStorePath,
			Usage: "Path where SKM should store its profiles",
		},
		cli.StringFlag{
			Name:  "ssh-path",
			Value: defaultSSHPath,
			Usage: "Path to a .ssh folder",
		},
		cli.StringFlag{
			Name:  "restic-path",
			Value: "",
			Usage: "Path to the restic binary",
		},
	}
	app.Name = skm.Name
	app.Usage = skm.Usage
	app.Version = Version
	app.Commands = initCommands()
	if err := app.Run(os.Args); err != nil {
		fmt.Println("Failed to run skm:", err)
	}
}

func mustGetEnvironment(ctx *cli.Context) *skm.Environment {
	storePath := ctx.GlobalString("store-path")
	sshPath := ctx.GlobalString("ssh-path")
	resticPath := ctx.GlobalString("restic-path")
	if storePath == "" || sshPath == "" {
		log.Fatalf("store-path (%v) and ssh-path (%v) have to be set!", storePath, sshPath)
	}

	// create SSH path if it doesn't exist
	if _, err := os.Stat(sshPath); os.IsNotExist(err) {
		err := os.Mkdir(sshPath, 0755)

		if err != nil {
			color.Red("%sFailed to initialize SSH path!", skm.CrossSymbol)
			return nil
		}
	}

	if resticPath == "" {
		resticPath, _ = exec.LookPath("restic")
	}
	return &skm.Environment{
		StorePath:  storePath,
		SSHPath:    sshPath,
		ResticPath: resticPath,
	}
}

// ParseArgs parses input arguments and displays the program logo
func parseArgs() {
	if len(os.Args) == 1 {
		displayLogo()
	} else if len(os.Args) == 2 {
		if os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "h" || os.Args[1] == "help" {
			displayLogo()
		}
	}
}
