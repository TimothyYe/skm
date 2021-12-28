package actions

import (
	"github.com/TimothyYe/skm/internal/utils"
	"github.com/fatih/color"
	"gopkg.in/urfave/cli.v1"
)

func Delete(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)
	var alias string

	if c.NArg() > 0 {
		alias = c.Args().Get(0)
	} else {
		color.Red("%sPlease input key alias name!", utils.CrossSymbol)
		return nil
	}

	keyMap := utils.LoadSSHKeys(env)
	key, ok := keyMap[alias]

	if !ok {
		color.Red("Key alias: %s doesn't exist!", alias)
		return nil
	}

	// Set key with related alias as default used key
	utils.DeleteKey(alias, key, env)
	return nil
}
