package actions

import (
	"errors"
	"github.com/TimothyYe/skm/internal/utils"
	"github.com/fatih/color"
	"gopkg.in/urfave/cli.v1"
)

func Cache(c *cli.Context) error {
	alias := c.Args().Get(0)
	env := utils.MustGetEnvironment(c)
	keyMap := utils.LoadSSHKeys(env)

	if c.Bool("add") {
		// add SSH key into SSH agent cache
		if err := utils.AddCache(alias, keyMap, env); err != nil {
			color.Red("%s"+err.Error(), utils.CrossSymbol)
			return nil
		}
		color.Green("%s SSH key [%s] already added into cache", utils.CheckSymbol, alias)
	} else if c.Bool("del") {
		// delete SSH key from SSH agent cache
		if err := utils.DeleteCache(alias, keyMap, env); err != nil {
			color.Red("%s"+err.Error(), utils.CrossSymbol)
			return nil
		}
		color.Green("%s SSH key [%s] removed from cache", utils.CheckSymbol, alias)
	} else if c.Bool("list") {
		// list all cached SSH keys from SSH agent cache
		if err := utils.ListCache(); err != nil {
			return nil
		}
	} else {
		color.Red("%s Invalid parameter!", utils.CrossSymbol)
		return errors.New("invalid parameter")
	}
	return nil
}
