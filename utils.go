package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func loadSingleKey(keyPath string) *SSHKey {
	key := &SSHKey{}

	//Walkthrough SSH key store and load all the keys
	err := filepath.Walk(keyPath, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}

		if path == keyPath {
			return nil
		}

		if f.IsDir() {
			return nil
		}

		if strings.Contains(f.Name(), ".pub") {
			key.PublicKey = f.Name()
			return nil
		} else {
			key.PrivateKey = f.Name()
			return nil
		}

		return nil
	})

	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
		return nil
	}

	if key.PublicKey != "" && key.PrivateKey != "" {
		return key
	}

	return nil
}

func loadSSHKeys() map[string]*SSHKey {
	keys := map[string]*SSHKey{}

	//Walkthrough SSH key store and load all the keys
	err := filepath.Walk(storePath, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}

		if path == storePath {
			return nil
		}

		if f.IsDir() {
			//Load private/public keys
			key := loadSingleKey(path)

			if key != nil {
				keys[f.Name()] = key
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
	}

	return keys
}
