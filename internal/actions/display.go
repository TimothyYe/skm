package actions

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TimothyYe/skm/internal/utils"
	"gopkg.in/urfave/cli.v1"
)

func Display(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)
	keys := utils.LoadSSHKeys(env)

	if c.NArg() > 0 {
		alias := c.Args().Get(0)
		if key, exists := keys[alias]; exists {
			fmt.Print(getKeyPayload(key.PublicKey))
			return nil
		}
		return errors.New("Key alias not found")
	}

	// When stdout is a TTY, offer the picker. When piped (e.g. `skm display |
	// pbcopy`), preserve the original behaviour of printing the default key so
	// existing scripts keep working.
	if isInteractive() {
		alias, err := pickKey("Select an SSH key to display", keys)
		if err != nil {
			return nil
		}
		if key, exists := keys[alias]; exists {
			fmt.Print(getKeyPayload(key.PublicKey))
		}
		return nil
	}

	for _, key := range keys {
		if key.IsDefault {
			keyPath := utils.ParsePath(filepath.Join(env.SSHPath, key.Type.PublicKey()))
			fmt.Print(getKeyPayload(keyPath))
			return nil
		}
	}

	return nil
}

func isInteractive() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
