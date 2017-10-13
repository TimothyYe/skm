package main

import (
	"os"
	"path/filepath"

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

func create(c *cli.Context) error {
	var alias string
	args := []string{}

	if c.NArg() > 0 {
		alias = c.Args().Get(0)
	} else {
		color.Red("%sPlease input key alias name!")
		os.Exit(1)
	}

	//Create alias directory
	err := os.Mkdir(filepath.Join(storePath, alias), 0755)

	if err != nil {
		color.Green("%sCreate SSH key failed!", checkSymbol)
		return nil
	}

	bits := c.String("b")
	if bits != "" {
		args = append(args, "-b")
		args = append(args, bits)
	}

	comment := c.String("C")
	if bits != "" {
		args = append(args, "-C")
		args = append(args, comment)
	}

	args = append(args, "-f")
	args = append(args, filepath.Join(storePath, alias, "id_rsa"))

	execute("ssh-keygen", args...)
	return nil
}

func list(c *cli.Context) error {
	keyMap := loadSSHKeys()

	if len(keyMap) == 0 {
		color.Green("%s No SSH key found!", checkSymbol)
		return nil
	}

	for k, _ := range keyMap {
		color.Blue("\t%s", k)
	}

	return nil
}
