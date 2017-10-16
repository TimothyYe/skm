package main

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/fatih/color"
	cli "gopkg.in/urfave/cli.v1"
)

func initialize(c *cli.Context) error {
	//Check the existence of key store
	if _, err := os.Stat(storePath); !os.IsNotExist(err) {
		color.Green("%sSSH key store already exists.", checkSymbol)
		return nil
	}

	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		err := os.Mkdir(storePath, 0755)

		if err != nil {
			color.Red("%sFailed to initialize SSH key store!", checkSymbol)
			return nil
		}
	}

	//Check & move existing keys into default folder
	if _, err := os.Stat(filepath.Join(sshPath, privateKey)); !os.IsNotExist(err) {
		//Create alias directory
		err := os.Mkdir(filepath.Join(storePath, defaultKey), 0755)
		if err != nil {
			color.Green("%sFailed to create default key store!", checkSymbol)
			return nil
		}

		//Move key to default key store
		os.Rename(filepath.Join(sshPath, privateKey), filepath.Join(storePath, defaultKey, privateKey))
		os.Rename(filepath.Join(sshPath, publicKey), filepath.Join(storePath, defaultKey, publicKey))

		//Create symbol link
		createLink(defaultKey)
	}

	color.Green("%sSSH key store initialized!", checkSymbol)
	color.Green("Key store location is: %s", storePath)
	return nil
}

func create(c *cli.Context) error {
	var alias string
	args := []string{}

	if c.NArg() > 0 {
		alias = c.Args().Get(0)
	} else {
		color.Red("%sPlease input key alias name!")
		return nil
	}

	//Check alias name existance
	keyMap := loadSSHKeys()

	if len(keyMap) > 0 {
		if _, ok := keyMap[alias]; ok {
			color.Red("%sSSH key alias already exists, please choose another one!", crossSymbol)
			return nil
		}
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

	execute("", "ssh-keygen", args...)
	color.Green("%sSSH key [%s] created!", checkSymbol, alias)
	return nil
}

func list(c *cli.Context) error {
	keyMap := loadSSHKeys()

	if len(keyMap) == 0 {
		color.Green("%s No SSH key found!", checkSymbol)
		return nil
	}

	color.Green("%sFound %d SSH key(s)!", checkSymbol, len(keyMap))

	var keys []string
	for k := range keyMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		key := keyMap[k]
		if key.IsDefault {
			color.Green("->\t%s", k)
		} else {
			color.Blue("\t%s", k)
		}
	}

	return nil
}

func use(c *cli.Context) error {
	var alias string

	if c.NArg() > 0 {
		alias = c.Args().Get(0)
	} else {
		color.Red("%sPlease input key alias name!")
		return nil
	}

	keyMap := loadSSHKeys()
	_, ok := keyMap[alias]

	if !ok {
		color.Red("Key alias: %s doesn't exist!", alias)
	}

	//Set key with related alias as default used key
	createLink(alias)
	color.Green("Now using SSH key: %s", alias)
	return nil
}

func delete(c *cli.Context) error {
	var alias string

	if c.NArg() > 0 {
		alias = c.Args().Get(0)
	} else {
		color.Red("%sPlease input key alias name!")
		return nil
	}

	keyMap := loadSSHKeys()
	key, ok := keyMap[alias]

	if !ok {
		color.Red("Key alias: %s doesn't exist!", alias)
	}

	//Set key with related alias as default used key
	deleteKey(alias, key)
	return nil
}

func backup(c *cli.Context) error {
	fileName := getBakFileName()
	dstFile := filepath.Join(os.Getenv("HOME"), fileName)

	result := execute(storePath, "tar", "-czvf", dstFile, ".")
	if result {
		color.Green("%s All SSH keys backup to: %s", checkSymbol, dstFile)
	}

	return nil
}
