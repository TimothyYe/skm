package skm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
)

const (
	// Name is the program name
	Name = "SKM"
	// Usage is for simple description
	Usage = "Manage your multiple SSH keys easily"

	// CheckSymbol is the code for check symbol
	CheckSymbol = "\u2714 "
	// CrossSymbol is the code for check symbol
	CrossSymbol = "\u2716 "

	// PublicKey is the default name of SSH public key
	PublicKey = "id_rsa.pub"
	// PrivateKey is the default name of SSH private key
	PrivateKey = "id_rsa"
	// DefaultKey is the default alias name of SSH key
	DefaultKey = "default"
)

var (
	// StorePath is the default SKM key path to store all the SSH keys
	StorePath = filepath.Join(os.Getenv("HOME"), ".skm")
	// SSHPath is the default SSH key path
	SSHPath = filepath.Join(os.Getenv("HOME"), ".ssh")
)

// Execute executes shell commands with arguments
func Execute(workDir, script string, args ...string) bool {
	cmd := exec.Command(script, args...)

	if workDir != "" {
		cmd.Dir = workDir
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		color.Red("%s%s", CrossSymbol, err.Error())
		return false
	}

	return true
}

// ClearKey clears both private & public keys from SSH key path
func ClearKey() {
	// Remove private key if exists
	PrivateKeyPath := filepath.Join(SSHPath, PrivateKey)
	os.Remove(PrivateKeyPath)

	// Remove public key if exists
	PublicKeyPath := filepath.Join(SSHPath, PublicKey)
	os.Remove(PublicKeyPath)
}

// DeleteKey delete key by its alias name
func DeleteKey(alias string, key *SSHKey, forTest ...bool) {
	inUse := key.PrivateKey == parsePath(filepath.Join(SSHPath, PrivateKey))
	var testMode bool
	var input string

	if len(forTest) > 0 {
		testMode = true
	} else {
		testMode = false
	}

	if !testMode {
		if inUse {
			fmt.Print(color.BlueString("SSH key [%s] is currently in use, please confirm to delete it [y/n]: ", alias))
		} else {
			fmt.Print(color.BlueString("Please confirm to delete SSH key [%s] [y/n]: ", alias))
		}
		fmt.Scan(&input)
	} else {
		input = "y"
		inUse = true
	}

	if input == "y" {
		if inUse {
			ClearKey()
		}

		//Remove specified key by alias name
		if err := os.RemoveAll(filepath.Join(StorePath, alias)); err == nil {
			color.Green("%sSSH key [%s] deleted!", CheckSymbol, alias)
		} else {
			color.Red("%sFailed to delete SSH key [%s]!", CrossSymbol, alias)
		}
	}
}

// CreateLink creates symbol link for specified SSH key
func CreateLink(alias string) {
	ClearKey()

	//Create symlink for private key
	os.Symlink(filepath.Join(StorePath, alias, PrivateKey), filepath.Join(SSHPath, PrivateKey))

	//Create symlink for public key
	os.Symlink(filepath.Join(StorePath, alias, PublicKey), filepath.Join(SSHPath, PublicKey))
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
			key.PublicKey = path
			return nil
		}

		//Check if key is in use
		key.PrivateKey = path

		if path == parsePath(filepath.Join(SSHPath, PrivateKey)) {
			key.IsDefault = true
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

func parsePath(path string) string {
	fileInfo, err := os.Lstat(path)

	if err != nil {
		return ""
	}

	if fileInfo.Mode()&os.ModeSymlink != 0 {
		originFile, err := os.Readlink(path)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		return originFile
	}
	return path
}

// LoadSSHKeys loads all the SSH keys from key store
func LoadSSHKeys() map[string]*SSHKey {
	keys := map[string]*SSHKey{}

	//Walkthrough SSH key store and load all the keys
	err := filepath.Walk(StorePath, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}

		if path == StorePath {
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

// GetBakFileName generates a backup file name by current date and time
func GetBakFileName() string {
	return fmt.Sprintf("skm-%s.tar.gz", time.Now().Format("20060102150405"))
}
