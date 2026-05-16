package actions

import (
	"sort"
	"strings"

	"github.com/TimothyYe/skm/internal/utils"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"gopkg.in/urfave/cli.v1"
)

func Use(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)
	var alias string
	keyMap := utils.LoadSSHKeys(env)

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

	// complete match key
	_, ok := keyMap[alias]

	if !ok {
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

	// Set key with related alias as default used key
	if err := utils.CreateLink(alias, keyMap, env); err != nil {
		color.Red("%s%s", utils.CrossSymbol, err.Error())
		return nil
	}
	// Run a potential hook
	utils.RunHook(alias, env)
	color.Green("Now using SSH key: [%s]", alias)
	return nil
}
