package main

import (
	"os"

	"github.com/TimothyYe/skm"

	cli "gopkg.in/urfave/cli.v1"
)

func main() {
	skm.ParseArgs()

	app := cli.NewApp()
	app.Name = skm.Name
	app.Usage = skm.Usage
	app.Version = skm.Version
	app.Commands = initCommands()
	app.Run(os.Args)
}
