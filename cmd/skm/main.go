package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/TimothyYe/skm"

	cli "gopkg.in/urfave/cli.v1"
)

var defaultStorePath = filepath.Join(os.Getenv("HOME"), ".skm")
var defaultSSHPath = filepath.Join(os.Getenv("HOME"), ".ssh")

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
	app.Run(os.Args)
}

func mustGetEnvironment(ctx *cli.Context) *skm.Environment {
	storePath := ctx.GlobalString("store-path")
	sshPath := ctx.GlobalString("ssh-path")
	resticPath := ctx.GlobalString("restic-path")
	if storePath == "" || sshPath == "" {
		log.Fatalf("store-path (%v) and ssh-path (%v) have to be set!", storePath, sshPath)
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
