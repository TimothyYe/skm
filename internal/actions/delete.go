package actions

import (
	"github.com/TimothyYe/skm/internal/utils"
	"github.com/fatih/color"
	"gopkg.in/urfave/cli.v1"
)

func Delete(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)
	keyMap := utils.LoadSSHKeys(env)

	var alias string
	if c.NArg() > 0 {
		alias = c.Args().Get(0)
	} else {
		picked, err := pickKey("Select an SSH key to delete", keyMap)
		if err != nil {
			if err == ErrNoKeys {
				color.Red("%sNo SSH keys to delete", utils.CrossSymbol)
			}
			return nil
		}
		alias = picked
	}

	key, ok := keyMap[alias]
	if !ok {
		color.Red("Key alias: %s doesn't exist!", alias)
		return nil
	}

	utils.DeleteKey(alias, key, env)
	return nil
}
