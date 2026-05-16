package actions

import (
	"sort"
	"strings"

	"github.com/TimothyYe/skm/internal/utils"
	"github.com/fatih/color"
	"gopkg.in/urfave/cli.v1"
)

func Use(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)
	keyMap := utils.LoadSSHKeys(env)

	var alias string
	if c.NArg() > 0 {
		alias = c.Args().Get(0)
	} else {
		picked, err := pickKey("Please select one SSH key", keyMap)
		if err != nil {
			return nil
		}
		alias = picked
	}

	// complete match key
	if _, ok := keyMap[alias]; !ok {
		// partial match key — iterate in sorted order so the result is deterministic
		names := make([]string, 0, len(keyMap))
		for k := range keyMap {
			names = append(names, k)
		}
		sort.Strings(names)

		var matches []string
		for _, k := range names {
			if strings.Contains(k, alias) {
				matches = append(matches, k)
			}
		}

		switch len(matches) {
		case 0:
			color.Red("Key alias: %s doesn't exist!", alias)
			return nil
		case 1:
			alias = matches[0]
		default:
			color.Red("Key alias: %s is ambiguous, matches: %s", alias, strings.Join(matches, ", "))
			return nil
		}
	}

	if err := utils.CreateLink(alias, keyMap, env); err != nil {
		color.Red("%s%s", utils.CrossSymbol, err.Error())
		return nil
	}
	utils.RunHook(alias, env)
	color.Green("Now using SSH key: [%s]", alias)
	return nil
}
