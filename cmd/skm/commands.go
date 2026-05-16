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
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "quiet, q", Usage: "Only print key alias names"},
				cli.BoolFlag{Name: "json", Usage: "Output as JSON for scripting"},
				cli.StringFlag{Name: "type, t", Usage: "Filter by key type (e.g. rsa, ed25519)"},
			},
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
			Usage:   "Copy SSH public key to a remote host",
			Action:  actions.Copy,
			Flags: []cli.Flag{
				cli.StringFlag{Name: "p", Usage: "SSH port"},
				cli.StringFlag{Name: "key, k", Usage: "Push the key with this alias instead of the active default"},
				cli.BoolFlag{Name: "pick", Usage: "Interactively pick the key to push"},
				cli.BoolFlag{Name: "dry-run", Usage: "Print the ssh-copy-id command without executing it"},
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
			Name:    "import",
			Aliases: []string{"im"},
			Usage:   "Import an existing SSH key pair from a path into the store",
			Action:  actions.Import,
			Flags: []cli.Flag{
				cli.StringFlag{Name: "alias, a", Usage: "Alias name for the imported key"},
				cli.BoolFlag{Name: "move", Usage: "Delete the source files after a successful import"},
			},
		},
		{
			Name:    "export",
			Aliases: []string{"ex"},
			Usage:   "Export a single key as a tar.gz bundle (optionally encrypted)",
			Action:  actions.Export,
			Flags: []cli.Flag{
				cli.StringFlag{Name: "output, o", Usage: "Output file path"},
				cli.BoolFlag{Name: "encrypt", Usage: "Encrypt the bundle with openssl AES-256-CBC"},
			},
		},
		{
			Name:    "doctor",
			Aliases: []string{"dr"},
			Usage:   "Run diagnostics against the SKM environment, agent, and stored keys",
			Action:  actions.Doctor,
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "json", Usage: "Emit check results as JSON"},
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
