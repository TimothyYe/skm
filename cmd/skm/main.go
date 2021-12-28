package main

import (
	"fmt"
	"github.com/TimothyYe/skm/internal/utils"
	"os"
	"path/filepath"

	cli "gopkg.in/urfave/cli.v1"
)

var home, _ = os.UserHomeDir()

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
	app.Name = utils.Name
	app.Usage = utils.Usage
	app.Version = Version
	app.Commands = initCommands()
	if err := app.Run(os.Args); err != nil {
		fmt.Println("Failed to run skm:", err)
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
