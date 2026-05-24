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
			Usage:   "Create a new SSH key (defaults to ed25519; rsa, ed25519-sk, ecdsa-sk also supported)",
			Action:  actions.Create,
			Flags: []cli.Flag{
				cli.StringFlag{Name: "b", Usage: "Key size in bits (RSA only; minimum 3072, default 3072)"},
				cli.StringFlag{Name: "C", Usage: "Key comment passed to ssh-keygen"},
				cli.StringFlag{Name: "t", Usage: "Key type: ed25519 (default), rsa, ed25519-sk, ecdsa-sk"},
			},
		},
		{
			Name:    "ls",
			Aliases: []string{"l"},
			Usage:   "List all the available SSH keys (alias + comment by default; use -l for full details)",
			Action:  actions.List,
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "long, l", Usage: "Show full details (type, bits, fingerprint, agent, created, comment)"},
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
			Usage:   "Delete specific SSH key by alias name (moves to trash; restore with `skm trash restore`)",
			Action:  actions.Delete,
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "yes, y", Usage: "Skip confirmation prompt"},
				cli.BoolFlag{Name: "purge", Usage: "Delete outright instead of moving to trash"},
			},
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
			Usage:   "Backup all SSH keys to an archive file (or a restic repository with --restic)",
			Action:  actions.Backup,
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "encrypt", Usage: "Encrypt the tar bundle with openssl AES-256-CBC"},
				cli.StringFlag{Name: "password-file", Usage: "Read the encryption passphrase from this file (skips the prompt)"},
				cli.BoolFlag{Name: "restic", Usage: "Use restic to generate backup"},
				cli.BoolFlag{Name: "init", Usage: "Interactively configure a restic repository (use with --restic)"},
			},
		},
		{
			Name:    "restore",
			Aliases: []string{"r"},
			Usage:   "Restore SSH keys from an existing archive file (.enc bundles are auto-decrypted)",
			Action:  actions.Restore,
			Flags: []cli.Flag{
				cli.StringFlag{Name: "password-file", Usage: "Read the decryption passphrase from this file"},
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
			Name:    "fingerprint",
			Aliases: []string{"fp"},
			Usage:   "Print the SHA256 fingerprint of an SSH key (default: active key)",
			Action:  actions.Fingerprint,
		},
		{
			Name:    "info",
			Aliases: []string{"in"},
			Usage:   "Show detailed information about an SSH key (default: active key)",
			Action:  actions.Info,
		},
		{
			Name:    "passphrase",
			Aliases: []string{"pp"},
			Usage:   "Add, rotate, or remove the passphrase on an SSH key",
			Action:  actions.Passphrase,
		},
		{
			Name:    "publish",
			Aliases: []string{"pub"},
			Usage:   "Upload an SSH public key to GitHub, GitLab, or Bitbucket",
			Action:  actions.Publish,
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "github", Usage: "Publish to GitHub"},
				cli.BoolFlag{Name: "gitlab", Usage: "Publish to GitLab"},
				cli.BoolFlag{Name: "bitbucket", Usage: "Publish to Bitbucket (requires --user)"},
				cli.StringFlag{Name: "url", Usage: "Base URL for self-hosted GitHub Enterprise or GitLab"},
				cli.StringFlag{Name: "user, u", Usage: "Account/workspace name (Bitbucket)"},
				cli.StringFlag{Name: "token", Usage: "API token (else read from env or gh/glab CLI)"},
				cli.StringFlag{Name: "title", Usage: "Title for the uploaded key (default: skm-<alias>-<host>-<date>)"},
				cli.BoolFlag{Name: "dry-run", Usage: "Print what would be sent without uploading"},
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
			Name:    "audit",
			Aliases: []string{"au"},
			Usage:   "Audit stored keys for weak strength, missing passphrases, and age",
			Action:  actions.Audit,
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "json", Usage: "Emit findings as JSON"},
				cli.BoolFlag{Name: "strict", Usage: "Treat warnings as failures (exit non-zero)"},
				cli.StringFlag{Name: "max-age", Value: "1y", Usage: "Flag keys older than this (e.g. 30d, 6m, 1y)"},
				cli.IntFlag{Name: "rsa-min", Value: 3072, Usage: "Minimum acceptable RSA key size in bits"},
			},
		},
		{
			Name:  "trash",
			Usage: "Manage soft-deleted SSH keys",
			Subcommands: []cli.Command{
				{
					Name:    "ls",
					Aliases: []string{"list"},
					Usage:   "List keys currently in the trash",
					Action:  actions.TrashList,
				},
				{
					Name:      "restore",
					Usage:     "Restore a trashed key back into the store",
					ArgsUsage: "<name>",
					Action:    actions.TrashRestore,
					Flags: []cli.Flag{
						cli.StringFlag{Name: "as", Usage: "Restore under a different alias than the original"},
					},
				},
				{
					Name:   "empty",
					Usage:  "Permanently delete every key in the trash",
					Action: actions.TrashEmpty,
					Flags: []cli.Flag{
						cli.BoolFlag{Name: "yes, y", Usage: "Skip confirmation prompt"},
					},
				},
			},
		},
		{
			Name:  "hook",
			Usage: "Inspect hooks wired to SKM events",
			Subcommands: []cli.Command{
				{
					Name:      "ls",
					Aliases:   []string{"list"},
					Usage:     "List hook scripts registered for SKM events",
					ArgsUsage: "[alias]",
					Action:    actions.HookList,
					Flags: []cli.Flag{
						cli.BoolFlag{Name: "all, a", Usage: "List hooks for every alias in the store"},
					},
				},
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
