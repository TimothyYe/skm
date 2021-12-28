package actions

import (
	"errors"
	"fmt"
	"github.com/TimothyYe/skm/internal/utils"
	"gopkg.in/urfave/cli.v1"
	"path/filepath"
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

	for _, key := range keys {
		if key.IsDefault {
			keyPath := utils.ParsePath(filepath.Join(env.SSHPath, key.Type.PublicKey()))
			fmt.Print(getKeyPayload(keyPath))
			return nil
		}
	}

	return nil
}
