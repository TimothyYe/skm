package actions

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/TimothyYe/skm/internal/models"
	"github.com/TimothyYe/skm/internal/utils"
	"github.com/fatih/color"
	"gopkg.in/urfave/cli.v1"
)

func Rename(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)
	keyMap := utils.LoadSSHKeys(env)

	var alias, newAlias string

	switch c.NArg() {
	case 2:
		alias = c.Args().Get(0)
		newAlias = c.Args().Get(1)
	case 1:
		alias = c.Args().Get(0)
		input, err := promptText("New alias name", validateNewAlias(keyMap))
		if err != nil {
			return nil
		}
		newAlias = input
	case 0:
		picked, err := pickKey("Select an SSH key to rename", keyMap)
		if err != nil {
			if err == ErrNoKeys {
				color.Red("%sNo SSH keys to rename", utils.CrossSymbol)
			}
			return nil
		}
		alias = picked
		input, err := promptText("New alias name", validateNewAlias(keyMap))
		if err != nil {
			return nil
		}
		newAlias = input
	default:
		color.Red("%s Please input current alias name and new alias name", utils.CrossSymbol)
		return nil
	}

	if err := os.Rename(filepath.Join(env.StorePath, alias), filepath.Join(env.StorePath, newAlias)); err == nil {
		color.Green("%s SSH key [%s] renamed to [%s]", utils.CheckSymbol, alias, newAlias)
	} else {
		color.Red("%s Failed to rename SSH key!", utils.CrossSymbol)
	}
	return nil
}

func validateNewAlias(keyMap map[string]*models.SSHKey) func(string) error {
	return func(input string) error {
		input = strings.TrimSpace(input)
		if input == "" {
			return errors.New("alias cannot be empty")
		}
		if strings.ContainsAny(input, "/\\ ") {
			return errors.New("alias cannot contain spaces or path separators")
		}
		if _, exists := keyMap[input]; exists {
			return errors.New("alias already exists")
		}
		return nil
	}
}
