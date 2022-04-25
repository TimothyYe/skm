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
		// partial match key
		canPartialMatch := false

		for k, _ := range keyMap {
			if strings.Index(k, alias) >= 0 {
				alias = k
				canPartialMatch = true
				break
			}
		}

		if !canPartialMatch {
			color.Red("Key alias: %s doesn't exist!", alias)
			return nil
		}
	}

	// Set key with related alias as default used key
	utils.CreateLink(alias, keyMap, env)
	// Run a potential hook
	utils.RunHook(alias, env)
	color.Green("Now using SSH key: [%s]", alias)
	return nil
}
