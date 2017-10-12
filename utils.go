package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	Name  = "SKM"
	Usage = "Manage your multiple SSH keys easily"

	CheckSymbol = "\u2714 "
	CrossSymbol = "\u2716 "
)

var (
	storePath = filepath.Join(os.Getenv("HOME"), ".skm")
	sshPath   = filepath.Join(os.Getenv("HOME"), ".ssh")
)

func parseArgs() {
	if len(os.Args) == 1 {
		displayLogo()
	} else if len(os.Args) == 2 {
		if os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "h" || os.Args[1] == "help" {
			displayLogo()
		}
	}
}

func execute(script string, args ...string) {
	cmd := exec.Command(script, args...)

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
}

func createLink(alias string) {
	public := "id_rsa.pub"
	private := "id_rsa"

	//Create symlink for private key
	os.Symlink(filepath.Join(storePath, alias, private), filepath.Join(sshPath, private))
	//Create symlink for public key
	os.Symlink(filepath.Join(storePath, alias, public), filepath.Join(sshPath, public))
}
