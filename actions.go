package main

import (
	"github.com/fatih/color"
	cli "gopkg.in/urfave/cli.v1"
)

func initialize(c *cli.Context) error {
	return nil
}

func list(c *cli.Context) error {
	keyMap := loadSSHKeys()

	for k, _ := range keyMap {
		color.Blue(k)
	}

	return nil
}
