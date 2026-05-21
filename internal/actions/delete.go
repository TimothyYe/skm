package actions

import (
	"fmt"
	"strings"

	"github.com/TimothyYe/skm/internal/models"
	"github.com/TimothyYe/skm/internal/utils"
	"github.com/fatih/color"
	"gopkg.in/urfave/cli.v1"
)

func Delete(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)
	keyMap := utils.LoadSSHKeys(env)
	assumeYes := c.Bool("yes")
	purge := c.Bool("purge")

	aliases := c.Args()
	if len(aliases) == 0 {
		picked, err := pickKey("Select an SSH key to delete", keyMap)
		if err != nil {
			if err == ErrNoKeys {
				color.Red("%sNo SSH keys to delete", utils.CrossSymbol)
			}
			return nil
		}
		aliases = []string{picked}
	}

	resolved := make([]string, 0, len(aliases))
	keys := make([]*models.SSHKey, 0, len(aliases))
	for _, alias := range aliases {
		key, ok := keyMap[alias]
		if !ok {
			color.Red("Key alias: %s doesn't exist!", alias)
			continue
		}
		resolved = append(resolved, alias)
		keys = append(keys, key)
	}

	if len(resolved) == 0 {
		return nil
	}

	perKeyAssumeYes := assumeYes
	if !assumeYes && len(resolved) > 1 {
		verb := "Delete"
		hint := " (moved to trash)"
		if purge {
			verb = "PURGE"
			hint = " (cannot be undone)"
		}
		fmt.Print(color.BlueString("%s %d SSH keys [%s]%s? [y/n]: ", verb, len(resolved), strings.Join(resolved, ", "), hint))
		var input string
		fmt.Scan(&input)
		if input != "y" {
			return nil
		}
		perKeyAssumeYes = true
	}

	for i, alias := range resolved {
		utils.DeleteKey(alias, keys[i], env, utils.DeleteOptions{AssumeYes: perKeyAssumeYes, Purge: purge})
	}
	return nil
}
