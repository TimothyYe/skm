package main

import (
	"github.com/TimothyYe/skm/internal/actions"
	cli "gopkg.in/urfave/cli.v1"
)

func initCommands() []cli.Command {
	return []cli.Command{
		{
			Name:    "init",
			Aliases: []string{"i"},
			Usage:   "Initialize SSH keys store for the first time use",
			Action:  actions.Initialize,
		},
		{
			Name:    "create",
			Aliases: []string{"c"},
			Usage:   "Create a new SSH key.",
			Action:  actions.Create,
			Flags: []cli.Flag{
				cli.StringFlag{Name: "b", Usage: "bits"},
				cli.StringFlag{Name: "C", Usage: "comment"},
				cli.StringFlag{Name: "t", Usage: "type"},
			},
		},
		{
			Name:    "ls",
			Aliases: []string{"l"},
			Usage:   "List all the available SSH keys",
			Action:  actions.List,
		},
		{
			Name:    "use",
			Aliases: []string{"u"},
			Usage:   "Set specific SSH key as default by its alias name",
			Action:  actions.Use,
		},
		{
			Name:    "delete",
			Aliases: []string{"d"},
			Usage:   "Delete specific SSH key by alias name",
			Action:  actions.Delete,
		},
		{
			Name:    "rename",
			Aliases: []string{"rn"},
			Usage:   "Rename SSH key alias name to a new one",
			Action:  actions.Rename,
		},
		{
			Name:    "copy",
			Aliases: []string{"cp"},
			Usage:   "Copy current SSH public key to a remote host",
			Action:  actions.Copy,
			Flags: []cli.Flag{
				cli.StringFlag{Name: "p", Usage: "SSH port"},
			},
		},
		{
			Name:    "display",
			Aliases: []string{"dp"},
			Usage:   "Display the current SSH public key or specific SSH public key by alias name",
			Action:  actions.Display,
		},
		{
			Name:    "backup",
			Aliases: []string{"b"},
			Usage:   "Backup all SSH keys to an archive file",
			Action:  actions.Backup,
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "restic", Usage: "Use restic to generate backup"},
			},
		},
		{
			Name:    "restore",
			Aliases: []string{"r"},
			Usage:   "Restore SSH keys from an existing archive file",
			Action:  actions.Restore,
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "restic", Usage: "Use restic to generate backup"},
				cli.StringFlag{Name: "restic-snapshot", Usage: "The snapshot to be restored"},
			},
		},
		{
			Name:   "cache",
			Usage:  "Add your SSH to SSH agent cache via alias name",
			Action: actions.Cache,
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "add", Usage: "Add SSH key to SSH agent cache"},
				cli.BoolFlag{Name: "del", Usage: "Remove SSH key from SSH agent cache"},
				cli.BoolFlag{Name: "list", Usage: "List all SSH keys from SSH agent cache"},
			},
		},
	}
}
