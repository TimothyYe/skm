package actions

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/TimothyYe/skm/internal/utils"
	"github.com/fatih/color"
	"gopkg.in/urfave/cli.v1"
)

func TrashList(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)
	entries, err := utils.ListTrash(env)
	if err != nil {
		color.Red("%s%s", utils.CrossSymbol, err.Error())
		return nil
	}
	if len(entries) == 0 {
		color.Green("Trash is empty")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tALIAS\tDELETED")
	for _, e := range entries {
		deleted := "-"
		if !e.DeletedAt.IsZero() {
			deleted = e.DeletedAt.Format("2006-01-02 15:04:05")
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", e.Name, e.Alias, deleted)
	}
	return w.Flush()
}

func TrashRestore(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)
	if c.NArg() == 0 {
		color.Red("%susage: skm trash restore <name> [--as <alias>]", utils.CrossSymbol)
		return nil
	}
	name := c.Args().Get(0)
	asAlias := c.String("as")

	restored, err := utils.RestoreFromTrash(name, asAlias, env)
	if err != nil {
		color.Red("%s%s", utils.CrossSymbol, err.Error())
		return nil
	}
	color.Green("%sRestored [%s] as alias [%s]", utils.CheckSymbol, name, restored)
	return nil
}

func TrashEmpty(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)
	entries, err := utils.ListTrash(env)
	if err != nil {
		color.Red("%s%s", utils.CrossSymbol, err.Error())
		return nil
	}
	if len(entries) == 0 {
		color.Green("Trash is already empty")
		return nil
	}

	if !c.Bool("yes") {
		fmt.Print(color.BlueString("Permanently delete %d trashed key(s)? This cannot be undone. [y/n]: ", len(entries)))
		var input string
		fmt.Scan(&input)
		if input != "y" {
			return nil
		}
	}

	removed, err := utils.EmptyTrash(env)
	if err != nil {
		color.Red("%s%s", utils.CrossSymbol, err.Error())
		return nil
	}
	color.Green("%sEmptied trash (%d entries removed)", utils.CheckSymbol, removed)
	return nil
}
