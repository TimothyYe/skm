package main

import (
	"os"

	cli "gopkg.in/urfave/cli.v1"
)

func main() {
	parseArgs()

	app := cli.NewApp()
	app.Name = Name
	app.Usage = Usage
	app.Version = Version
	app.Commands = initCommands()
	app.Run(os.Args)
}
