package main

import (
	"os"

	"github.com/TimothyYe/skm"

	cli "gopkg.in/urfave/cli.v1"
)

func main() {
	parseArgs()

	app := cli.NewApp()
	app.Name = skm.Name
	app.Usage = skm.Usage
	app.Version = Version
	app.Commands = initCommands()
	app.Run(os.Args)
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
