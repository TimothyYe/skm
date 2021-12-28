package actions

import (
	"fmt"
	"github.com/TimothyYe/skm/internal/utils"
	"github.com/fatih/color"
	"gopkg.in/urfave/cli.v1"
	"sort"
	"strings"
)

func List(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)
	keyMap := utils.LoadSSHKeys(env)

	if len(keyMap) == 0 {
		color.Green("%s No SSH key found!", utils.CheckSymbol)
		return nil
	}

	color.Green("\r\n%sFound %d SSH key(s)!", utils.CheckSymbol, len(keyMap))
	fmt.Println()

	var keys []string
	for k := range keyMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		key := keyMap[k]
		keyDesc := ""
		keyType := ""

		keyStr := strings.SplitAfterN(getKeyPayload(key.PublicKey), " ", 3)
		if len(keyStr) >= 3 {
			keyDesc = strings.TrimSpace(keyStr[2])
			keyType = strings.TrimSpace(keyStr[0])
		}
		if key.IsDefault {
			color.Green("->\t%s\t[%s]\t[%s]", k, keyType, keyDesc)
		} else {
			color.Blue("\t%s\t[%s]\t[%s]", k, keyType, keyDesc)
		}
	}
	return nil

}
