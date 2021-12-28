package actions

import (
	"fmt"
	"github.com/TimothyYe/skm/internal/models"
	"github.com/TimothyYe/skm/internal/utils"
	"github.com/fatih/color"
	cli "gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"os"
	"path/filepath"
)

func Initialize(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)
	// Remove existing empty key store if exists
	if _, err := utils.IsEmpty(env.StorePath); !os.IsNotExist(err) {
		err := os.Remove(env.StorePath)

		if err != nil {
			color.Red("%sFailed to remove existing empty key store!", utils.CrossSymbol)
			return nil
		}
	}

	// Check the existence of key store
	if _, err := os.Stat(env.StorePath); !os.IsNotExist(err) {
		color.Green("%sSSH key store already exists.", utils.CheckSymbol)
		return nil
	}

	if _, err := os.Stat(env.StorePath); os.IsNotExist(err) {
		err := os.Mkdir(env.StorePath, 0755)

		if err != nil {
			color.Red("%sFailed to initialize SSH key store!", utils.CrossSymbol)
			return nil
		}
	}

	// Check & move existing keys into default folder
	// TODO: Support different initial default keys
	for _, kt := range models.SupportedKeyTypes {
		if _, err := os.Stat(filepath.Join(env.SSHPath, kt.PrivateKey())); !os.IsNotExist(err) {
			// Create alias directory
			err := os.Mkdir(filepath.Join(env.StorePath, utils.DefaultKey), 0755)
			if err != nil {
				color.Red("%sFailed to create default key store!", utils.CrossSymbol)
				return nil
			}

			// Move key to default key store
			if err := os.Rename(filepath.Join(env.SSHPath, kt.PrivateKey()), filepath.Join(env.StorePath, utils.DefaultKey, kt.PrivateKey())); err != nil {
				color.Red("%sFailed to move key to default key store")
			}
			if err := os.Rename(filepath.Join(env.SSHPath, kt.PublicKey()), filepath.Join(env.StorePath, utils.DefaultKey, kt.PublicKey())); err != nil {
				color.Red("%sFailed to move key to default key store")
			}

			// Once we have the old keys in place, we can load the key map.
			keyMap := utils.LoadSSHKeys(env)

			// Create symbol link
			utils.CreateLink(utils.DefaultKey, keyMap, env)
			break
		}
	}

	color.Green("%sSSH key store initialized!", utils.CheckSymbol)
	color.Green("Key store location is: %s", env.StorePath)
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
