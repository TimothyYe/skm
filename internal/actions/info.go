package actions

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/TimothyYe/skm/internal/models"
	"github.com/TimothyYe/skm/internal/utils"
	"github.com/fatih/color"
	"gopkg.in/urfave/cli.v1"
)

// Fingerprint prints the SHA256 fingerprint of an SSH key. When no alias is
// given, it falls back to the currently active default. Output is plain so
// the command is usable in scripts (e.g. `skm fingerprint work | xargs ...`).
func Fingerprint(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)
	keys := utils.LoadSSHKeys(env)

	key, alias, err := resolveAliasOrDefault(c, keys)
	if err != nil {
		color.Red("%s%s", utils.CrossSymbol, err.Error())
		return nil
	}
	details := inspectKey(key.PublicKey)
	if details.Fingerprint == "" {
		color.Red("%sCould not compute fingerprint for [%s]", utils.CrossSymbol, alias)
		return nil
	}
	fmt.Println(details.Fingerprint)
	return nil
}

// Info prints a multi-line summary of a key: type, bits, fingerprint, comment,
// agent state, on-disk paths, and modification date. Designed for human reading;
// `skm ls --json` is the scripting path.
func Info(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)
	keys := utils.LoadSSHKeys(env)

	key, alias, err := resolveAliasOrDefault(c, keys)
	if err != nil {
		color.Red("%s%s", utils.CrossSymbol, err.Error())
		return nil
	}
	details := inspectKey(key.PublicKey)
	agent := loadAgentFingerprints()

	keyType := "-"
	if key.Type != nil {
		keyType = key.Type.Name
	}
	created := "-"
	if info, err := os.Stat(key.PrivateKey); err == nil {
		created = info.ModTime().Format("2006-01-02 15:04:05")
	}
	inAgent := "no"
	if details.Fingerprint != "" && agent[details.Fingerprint] {
		inAgent = "yes"
	}
	defaultMark := "no"
	if key.IsDefault {
		defaultMark = "yes"
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Alias\t%s\n", alias)
	fmt.Fprintf(w, "Default\t%s\n", defaultMark)
	fmt.Fprintf(w, "Type\t%s\n", keyType)
	fmt.Fprintf(w, "Bits\t%s\n", dashIfEmpty(details.Bits))
	fmt.Fprintf(w, "Fingerprint\t%s\n", dashIfEmpty(details.Fingerprint))
	fmt.Fprintf(w, "Comment\t%s\n", dashIfEmpty(details.Comment))
	fmt.Fprintf(w, "In agent\t%s\n", inAgent)
	fmt.Fprintf(w, "Private\t%s\n", key.PrivateKey)
	fmt.Fprintf(w, "Public\t%s\n", key.PublicKey)
	fmt.Fprintf(w, "Modified\t%s\n", created)
	w.Flush()

	fmt.Print(buf.String())
	return nil
}

func dashIfEmpty(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

// resolveAliasOrDefault picks the key to operate on: the positional alias if
// supplied, otherwise the currently active default. Returns a friendly error
// when neither is available.
func resolveAliasOrDefault(c *cli.Context, keys map[string]*models.SSHKey) (*models.SSHKey, string, error) {
	if c.NArg() > 0 {
		alias := c.Args().Get(0)
		key, ok := keys[alias]
		if !ok {
			return nil, "", fmt.Errorf("SSH key alias [%s] not found", alias)
		}
		return key, alias, nil
	}
	for alias, key := range keys {
		if key.IsDefault {
			return key, alias, nil
		}
	}
	return nil, "", fmt.Errorf("no active SSH key; pass an alias or run `skm use <alias>` first")
}
