package actions

import (
	"github.com/TimothyYe/skm/internal/utils"
	"github.com/fatih/color"
	"gopkg.in/urfave/cli.v1"
)

// Passphrase rotates, adds, or removes the passphrase on a private key by
// shelling out to `ssh-keygen -p`. ssh-keygen handles all three operations
// from its interactive prompt: empty input on "new passphrase" removes it,
// non-empty rotates or adds. Doing this through SKM saves the user from
// remembering the flag (`-p`) and the path to the private key.
func Passphrase(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)
	keys := utils.LoadSSHKeys(env)

	key, alias, err := resolveAliasOrDefault(c, keys)
	if err != nil {
		color.Red("%s%s", utils.CrossSymbol, err.Error())
		return nil
	}

	color.Blue("Updating passphrase for [%s] (%s)", alias, key.PrivateKey)
	if ok := utils.Execute("", "ssh-keygen", "-p", "-f", key.PrivateKey); !ok {
		return nil
	}
	color.Green("%sPassphrase updated for [%s]", utils.CheckSymbol, alias)
	return nil
}
