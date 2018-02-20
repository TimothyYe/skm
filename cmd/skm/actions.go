package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/TimothyYe/skm"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	cli "gopkg.in/urfave/cli.v1"
)

func initialize(c *cli.Context) error {
	env := mustGetEnvironment(c)
	// Remove existing empty key store if exists
	if _, err := skm.IsEmpty(env.StorePath); !os.IsNotExist(err) {
		err := os.Remove(env.StorePath)

		if err != nil {
			color.Red("%sFailed to remove existing empty key store!", skm.CrossSymbol)
			return nil
		}
	}

	// Check the existence of key store
	if _, err := os.Stat(env.StorePath); !os.IsNotExist(err) {
		color.Green("%sSSH key store already exists.", skm.CheckSymbol)
		return nil
	}

	if _, err := os.Stat(env.StorePath); os.IsNotExist(err) {
		err := os.Mkdir(env.StorePath, 0755)

		if err != nil {
			color.Red("%sFailed to initialize SSH key store!", skm.CrossSymbol)
			return nil
		}
	}

	// Check & move existing keys into default folder
	// TODO: Support different initial default keys
	for _, kt := range skm.SupportedKeyTypes {
		if _, err := os.Stat(filepath.Join(env.SSHPath, kt.PrivateKey())); !os.IsNotExist(err) {
			// Create alias directory
			err := os.Mkdir(filepath.Join(env.StorePath, skm.DefaultKey), 0755)
			if err != nil {
				color.Red("%sFailed to create default key store!", skm.CrossSymbol)
				return nil
			}

			// Move key to default key store
			os.Rename(filepath.Join(env.SSHPath, kt.PrivateKey()), filepath.Join(env.StorePath, skm.DefaultKey, kt.PrivateKey()))
			os.Rename(filepath.Join(env.SSHPath, kt.PublicKey()), filepath.Join(env.StorePath, skm.DefaultKey, kt.PublicKey()))

			// Once we have the old keys in place, we can load the key map.
			keyMap := skm.LoadSSHKeys(env)

			// Create symbol link
			skm.CreateLink(skm.DefaultKey, keyMap, env)
			break
		}
	}

	color.Green("%sSSH key store initialized!", skm.CheckSymbol)
	color.Green("Key store location is: %s", env.StorePath)
	return nil
}

func create(c *cli.Context) error {
	env := mustGetEnvironment(c)
	var alias string
	args := []string{}

	if c.NArg() > 0 {
		alias = c.Args().Get(0)
	} else {
		color.Red("%sPlease input key alias name!", skm.CrossSymbol)
		return nil
	}

	// Check alias name existence
	keyMap := skm.LoadSSHKeys(env)

	if len(keyMap) > 0 {
		if _, ok := keyMap[alias]; ok {
			color.Red("%sSSH key alias already exists, please choose another one!", skm.CrossSymbol)
			return nil
		}
	}

	// Create alias directory
	err := os.Mkdir(filepath.Join(env.StorePath, alias), 0755)

	if err != nil {
		color.Red("%sCreate SSH key failed!", skm.CrossSymbol)
		return nil
	}

	keyType := c.String("t")
	if keyType == "" {
		keyType = "rsa"
	}
	args = append(args, "-t")
	args = append(args, keyType)

	keyTypeSettings, ok := skm.SupportedKeyTypes[keyType]
	if !ok {
		color.Red("%s is not a supported key type.", keyType)
		return nil
	}

	args = append(args, "-f")

	if keyTypeSettings.SupportsVariableBitsize {
		bits := c.String("b")
		if bits != "" {
			args = append(args, "-b")
			args = append(args, bits)
		}
	}

	comment := c.String("C")
	if comment != "" {
		args = append(args, "-C")
		args = append(args, comment)
	}

	fileName := keyTypeSettings.KeyBaseName
	args = append(args, filepath.Join(env.StorePath, alias, fileName))

	skm.Execute("", "ssh-keygen", args...)
	color.Green("%sSSH key [%s] created!", skm.CheckSymbol, alias)
	return nil
}

func list(c *cli.Context) error {
	env := mustGetEnvironment(c)
	keyMap := skm.LoadSSHKeys(env)

	if len(keyMap) == 0 {
		color.Green("%s No SSH key found!", skm.CheckSymbol)
		return nil
	}

	color.Green("\r\n%sFound %d SSH key(s)!", skm.CheckSymbol, len(keyMap))
	fmt.Println()

	var keys []string
	for k := range keyMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		key := keyMap[k]
		keyDesc := ""

		keyStr := strings.SplitAfterN(getKeyPayload(key.PublicKey), " ", 3)
		if len(keyStr) >= 3 {
			keyDesc = strings.TrimSpace(keyStr[2])
		}
		if key.IsDefault {
			color.Green("->\t%s\t[%s]", k, keyDesc)
		} else {
			color.Blue("\t%s\t[%s]", k, keyDesc)
		}
	}
	return nil

}

func use(c *cli.Context) error {
	env := mustGetEnvironment(c)
	var alias string
	keyMap := skm.LoadSSHKeys(env)

	if c.NArg() > 0 {
		alias = c.Args().Get(0)
	} else {
		templates := &promptui.SelectTemplates{
			Active:   "{{ . | white | bgGreen }} ",
			Inactive: "{{ . }} ",
			Selected: "{{ . | bold }} ",
		}

		// Construct prompt menu items
		var names []string

		for k := range keyMap {
			names = append(names, k)
		}

		sort.Strings(names)

		prompt := promptui.Select{
			Label:     "Please select one SSH key",
			Items:     names,
			Templates: templates,
		}

		_, result, err := prompt.Run()

		if err != nil {
			return nil
		}

		alias = result
	}

	_, ok := keyMap[alias]

	if !ok {
		color.Red("Key alias: %s doesn't exist!", alias)
		return nil
	}

	// Set key with related alias as default used key
	skm.CreateLink(alias, keyMap, env)
	// Run a potential hook
	skm.RunHook(alias, env)
	color.Green("Now using SSH key: [%s]", alias)
	return nil
}

func delete(c *cli.Context) error {
	env := mustGetEnvironment(c)
	var alias string

	if c.NArg() > 0 {
		alias = c.Args().Get(0)
	} else {
		color.Red("%sPlease input key alias name!", skm.CrossSymbol)
		return nil
	}

	keyMap := skm.LoadSSHKeys(env)
	key, ok := keyMap[alias]

	if !ok {
		color.Red("Key alias: %s doesn't exist!", alias)
		return nil
	}

	// Set key with related alias as default used key
	skm.DeleteKey(alias, key, env)
	return nil
}

func rename(c *cli.Context) error {
	env := mustGetEnvironment(c)
	var alias, newAlias string

	if c.NArg() == 2 {
		alias = c.Args().Get(0)
		newAlias = c.Args().Get(1)

		err := os.Rename(filepath.Join(env.StorePath, alias), filepath.Join(env.StorePath, newAlias))

		if err == nil {
			color.Green("%s SSH key [%s] renamed to [%s]", skm.CheckSymbol, alias, newAlias)
		} else {
			color.Red("%s Failed to rename SSH key!", skm.CrossSymbol)
		}
	} else {
		color.Red("%s Please input current alias name and new alias name", skm.CrossSymbol)
	}

	return nil
}

func copy(c *cli.Context) error {
	env := mustGetEnvironment(c)
	host := c.Args().Get(0)
	args := []string{}

	port := c.String("p")
	if port != "" {
		args = append(args, "-p")
		args = append(args, port)
	}

	keyPath := skm.ParsePath(filepath.Join(env.SSHPath, skm.PrivateKey))
	args = append(args, "-i")
	args = append(args, keyPath)
	args = append(args, host)

	result := skm.Execute("", "ssh-copy-id", args...)

	if result {
		color.Green("%s Current SSH key already copied to remote host", skm.CheckSymbol)
	}

	return nil
}

func display(c *cli.Context) error {
	env := mustGetEnvironment(c)
	for _, key := range skm.LoadSSHKeys(env) {
		if key.IsDefault {
			keyPath := skm.ParsePath(filepath.Join(env.SSHPath, key.Type.PublicKey()))
			fmt.Print(getKeyPayload(keyPath))
			return nil
		}
	}
	return nil
}

func backup(c *cli.Context) error {
	env := mustGetEnvironment(c)
	fileName := skm.GetBakFileName()
	dstFile := filepath.Join(os.Getenv("HOME"), fileName)

	result := skm.Execute(env.StorePath, "tar", "-czvf", dstFile, ".")
	if result {
		color.Green("%s All SSH keys backup to: %s", skm.CheckSymbol, dstFile)
	}

	return nil
}

func restore(c *cli.Context) error {
	env := mustGetEnvironment(c)
	var filePath string

	if c.NArg() > 0 {
		filePath = c.Args().Get(0)
	} else {
		color.Red("%sPlease input the corrent backup file path!", skm.CrossSymbol)
		return nil
	}

	// Clear the key store first
	err := os.RemoveAll(env.StorePath)

	if err != nil {
		fmt.Println("Clear store path failed:", err.Error())
	}

	// Clear all keys
	skm.ClearKey(env)

	err = os.Mkdir(env.StorePath, 0755)

	if err != nil {
		color.Red("%sFailed to initialize SSH key store!", skm.CrossSymbol)
		return nil
	}

	// Extract backup file
	result := skm.Execute(env.StorePath, "tar", "zxvf", filePath, "-C", env.StorePath)

	if result {
		fmt.Println()
		color.Green("%s All SSH keys restored to %s", skm.CheckSymbol, env.StorePath)
	}

	return nil
}

func getKeyPayload(keyPath string) string {
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		fmt.Println("Failed to read ", keyPath)
		return ""
	}
	return string(key)
}
