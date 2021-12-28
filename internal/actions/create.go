package actions

import (
	"github.com/TimothyYe/skm/internal/models"
	"github.com/TimothyYe/skm/internal/utils"
	"github.com/fatih/color"
	"gopkg.in/urfave/cli.v1"
	"os"
	"path/filepath"
)

func Create(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)
	var alias string
	args := []string{}

	if c.NArg() > 0 {
		alias = c.Args().Get(0)
	} else {
		color.Red("%sPlease input key alias name!", utils.CrossSymbol)
		return nil
	}

	// Check alias name existence
	keyMap := utils.LoadSSHKeys(env)

	if len(keyMap) > 0 {
		if _, ok := keyMap[alias]; ok {
			color.Red("%sSSH key alias already exists, please choose another one!", utils.CrossSymbol)
			return nil
		}
	}

	// Remove existing empty alias directory if exists
	if _, err := utils.IsEmpty(filepath.Join(env.StorePath, alias)); !os.IsNotExist(err) {
		err := os.Remove(filepath.Join(env.StorePath, alias))

		if err != nil {
			color.Red("%sFailed to remove existing empty alias directory!", utils.CrossSymbol)
			return nil
		}
	}

	// Create alias directory
	err := os.Mkdir(filepath.Join(env.StorePath, alias), 0755)

	if err != nil {
		color.Red("%sCreate SSH key failed!", utils.CrossSymbol)
		return nil
	}

	keyType := c.String("t")
	if keyType == "" {
		keyType = "rsa"
	}
	args = append(args, "-t")
	args = append(args, keyType)

	keyTypeSettings, ok := models.SupportedKeyTypes[keyType]
	if !ok {
		color.Red("%s is not a supported key type.", keyType)
		return nil
	}

	args = append(args, "-f")
	fileName := keyTypeSettings.KeyBaseName
	args = append(args, filepath.Join(env.StorePath, alias, fileName))

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

	utils.Execute("", "ssh-keygen", args...)
	color.Green("%sSSH key [%s] created!", utils.CheckSymbol, alias)
	return nil
}
