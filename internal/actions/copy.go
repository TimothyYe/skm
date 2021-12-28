package actions

import (
	"github.com/TimothyYe/skm/internal/utils"
	"github.com/fatih/color"
	"gopkg.in/urfave/cli.v1"
	"path/filepath"
)

func Copy(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)
	host := c.Args().Get(0)
	args := []string{}

	port := c.String("p")
	if port != "" {
		args = append(args, "-p")
		args = append(args, port)
	}

	keyPath := utils.ParsePath(filepath.Join(env.SSHPath, utils.PrivateKey))
	args = append(args, "-i")
	args = append(args, keyPath)
	args = append(args, host)

	result := utils.Execute("", "ssh-copy-id", args...)

	if result {
		color.Green("%s Current SSH key already copied to remote host", utils.CheckSymbol)
	}

	return nil
}
