package main

import (
	"os"

	"github.com/fatih/color"
	cli "gopkg.in/urfave/cli.v1"
)

func initialize(c *cli.Context) error {
	//Check the existence of key store
	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		err := os.Mkdir(storePath, 0755)

		if err == nil {
			color.Green("%sSSH key store initialized!", checkSymbol)
			return nil
		}
	}

	if _, err := os.Stat(storePath); !os.IsNotExist(err) {
		color.Green("%sSSH key store already exists.", checkSymbol)
	}

	return nil
}

func list(c *cli.Context) error {
	keyMap := loadSSHKeys()

	if len(keyMap) == 0 {
		color.Green("%s No SSH key found!", checkSymbol)
		return nil
	}

	for k, _ := range keyMap {
		color.Blue(k)
	}

	return nil
}
