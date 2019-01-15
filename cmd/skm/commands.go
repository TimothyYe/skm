package main

import (
	cli "gopkg.in/urfave/cli.v1"
)

func initCommands() []cli.Command {
	return []cli.Command{
		{
			Name:    "init",
			Aliases: []string{"i"},
			Usage:   "Initialize SSH keys store for the first time use",
			Action:  initialize,
		},
		{
			Name:    "create",
			Aliases: []string{"c"},
			Usage:   "Create a new SSH key.",
			Action:  create,
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
			Action:  list,
		},
		{
			Name:    "use",
			Aliases: []string{"u"},
			Usage:   "Set specific SSH key as default by its alias name",
			Action:  use,
		},
		{
			Name:    "delete",
			Aliases: []string{"d"},
			Usage:   "Delete specific SSH key by alias name",
			Action:  delete,
		},
		{
			Name:    "rename",
			Aliases: []string{"rn"},
			Usage:   "Rename SSH key alias name to a new one",
			Action:  rename,
		},
		{
			Name:    "copy",
			Aliases: []string{"cp"},
			Usage:   "Copy current SSH public key to a remote host",
			Action:  copy,
			Flags: []cli.Flag{
				cli.StringFlag{Name: "p", Usage: "SSH port"},
			},
		},
		{
			Name:    "display",
			Aliases: []string{"dp"},
			Usage:   "Display the current SSH public key or specific SSH public key by alias name",
			Action:  display,
		},
		{
			Name:    "backup",
			Aliases: []string{"b"},
			Usage:   "Backup all SSH keys to an archive file",
			Action:  backup,
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "restic", Usage: "Use restic to generate backup"},
			},
		},
		{
			Name:    "restore",
			Aliases: []string{"r"},
			Usage:   "Restore SSH keys from an existing archive file",
			Action:  restore,
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "restic", Usage: "Use restic to generate backup"},
				cli.StringFlag{Name: "restic-snapshot", Usage: "The snapshot to be restored"},
			},
		},
		{
			Name:   "cache",
			Usage:  "Add your SSH to SSH agent cache via alias name",
			Action: cache,
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "add", Usage: "Add SSH key to SSH agent cache"},
				cli.BoolFlag{Name: "del", Usage: "Remove SSH key from SSH agent cache"},
				cli.BoolFlag{Name: "list", Usage: "List all SSH keys from SSH agent cache"},
			},
		},
	}
}
