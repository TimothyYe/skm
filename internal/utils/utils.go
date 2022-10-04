package utils

import (
	"errors"
	"fmt"
	"github.com/TimothyYe/skm/internal/models"
	"gopkg.in/urfave/cli.v1"
	"io"
	"log"
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

	// Ed25519PublicKey is the default name of ed25519 SSH public key
	Ed25519PublicKey = "id_ed25519.pub"
	// Ed25519PrivateKey is the default name of ed25519 SSH private key
	Ed25519PrivateKey = "id_ed25519"

	// DefaultKey is the default alias name of SSH key
	DefaultKey = "default"

	// HookName is the name of a hook that is called when present after using a key
	HookName = "hook"
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
func ClearKey(env *models.Environment, ty string) {
	for _, kt := range models.SupportedKeyTypes {
		if env.KeepTypeKeys && ty != "" && ty != kt.Name {
			continue
		}

		// Remove private key if exists
		PrivateKeyPath := filepath.Join(env.SSHPath, kt.PrivateKey())
		os.Remove(PrivateKeyPath)

		// Remove public key if exists
		PublicKeyPath := filepath.Join(env.SSHPath, kt.PublicKey())
		os.Remove(PublicKeyPath)
	}
}

// DeleteKey delete key by its alias name
func DeleteKey(alias string, key *models.SSHKey, env *models.Environment, forTest ...bool) {
	inUse := key.PrivateKey == ParsePath(filepath.Join(env.SSHPath, key.PrivateKey))
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
			ClearKey(env, key.Type.Name)
		}

		//Remove specified key by alias name
		if err := os.RemoveAll(filepath.Join(env.StorePath, alias)); err == nil {
			color.Green("%sSSH key [%s] deleted!", CheckSymbol, alias)
		} else {
			color.Red("%sFailed to delete SSH key [%s]!", CrossSymbol, alias)
		}
	}
}

// RunHook runs hook file after switching SSH key
func RunHook(alias string, env *models.Environment) {
	if info, err := os.Stat(filepath.Join(env.StorePath, alias, HookName)); !os.IsNotExist(err) {
		if info.Mode()&0111 != 0 {
			Execute("", filepath.Join(env.StorePath, alias, HookName), alias)
		}
	}
}

// AddCache adds SSH to ssh agent cache via key alias
func AddCache(alias string, keyMap map[string]*models.SSHKey, env *models.Environment) error {
	key, found := keyMap[alias]

	if !found {
		return fmt.Errorf("SSH key [%s] not found", alias)
	}

	// Add key to SSH agent cache
	privateKeyPath := filepath.Join(env.StorePath, alias, key.Type.PrivateKey())
	args := []string{privateKeyPath}
	result := Execute("", "ssh-add", args...)

	if !result {
		return errors.New("Failed to remove SSH key from cache")
	}

	return nil
}

// DeleteCache removes SSH key from SSH agent cache via key alias
func DeleteCache(alias string, keyMap map[string]*models.SSHKey, env *models.Environment) error {
	key, found := keyMap[alias]

	if !found {
		return fmt.Errorf("SSH key [%s] not found", alias)
	}

	// Remove key from SSH agent cache
	privateKeyPath := filepath.Join(env.StorePath, alias, key.Type.PrivateKey())
	args := []string{"-d", privateKeyPath}
	result := Execute("", "ssh-add", args...)

	if !result {
		return errors.New("Failed to remove SSH key from cache")
	}

	return nil
}

// ListCache lists cached SSH key from SSH agent cache
func ListCache() error {
	args := []string{"-l"}
	result := Execute("", "ssh-add", args...)

	if !result {
		return errors.New("Failed to list SSH keys from cache")
	}

	return nil
}

// CreateLink creates symbol link for specified SSH key
func CreateLink(alias string, keyMap map[string]*models.SSHKey, env *models.Environment) {
	key, found := keyMap[alias]
	ty := ""
	if found {
		ty = key.Type.Name
	}
	ClearKey(env, ty)

	if !found {
		return
	}

	relStorePath, err := filepath.Rel(env.SSHPath, env.StorePath)
	if err != nil {
		fmt.Println("Failed to find rel store path")
	}

	//Create symlink for private key
	if err := os.Symlink(filepath.Join(relStorePath, alias, key.Type.PrivateKey()), filepath.Join(env.SSHPath, key.Type.PrivateKey())); err != nil {
		fmt.Println("Failed to create symble link for private key")
		return
	}

	//Create symlink for public key
	if err := os.Symlink(filepath.Join(relStorePath, alias, key.Type.PublicKey()), filepath.Join(env.SSHPath, key.Type.PublicKey())); err != nil {
		fmt.Println("Failed to create symble link for private key")
		return
	}
}

func loadSingleKey(keyPath string, env *models.Environment) *models.SSHKey {
	key := &models.SSHKey{}

	//Walk-through SSH key store and load all the keys
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

		kt, ok := models.SupportedKeyTypes.GetByFilename(f.Name())
		if !ok {
			return nil
		}
		key.Type = &kt

		//Check if key is in use
		key.PrivateKey = path

		parsedPath := ParsePath(filepath.Join(env.SSHPath, kt.KeyBaseName))

		if path == parsedPath {
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

// ParsePath return the original SSH key path if it is a symbol link
func ParsePath(path string) string {
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

		if !filepath.IsAbs(originFile) {
			originFile = filepath.Join(filepath.Dir(path), originFile)
		}

		return originFile
	}
	return path
}

// LoadSSHKeys loads all the SSH keys from key store
func LoadSSHKeys(env *models.Environment) map[string]*models.SSHKey {
	keys := map[string]*models.SSHKey{}

	//Walk-through SSH key store and load all the keys
	err := filepath.Walk(env.StorePath, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}

		if path == env.StorePath {
			return nil
		}

		if f.IsDir() {
			//Load private/public keys
			key := loadSingleKey(path, env)

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

// IsEmpty checks if directory in path is empty
func IsEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

// Fatalf output formatted fatal error info
func Fatalf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	color.New(color.FgRed).Fprintf(os.Stderr, "%s %s", CrossSymbol, msg)
	os.Exit(1)
}

func MustGetEnvironment(ctx *cli.Context) *models.Environment {
	storePath := ctx.GlobalString("store-path")
	sshPath := ctx.GlobalString("ssh-path")
	resticPath := ctx.GlobalString("restic-path")
	keepTypeKeys := ctx.GlobalBool("keep-type-keys")
	if storePath == "" || sshPath == "" {
		log.Fatalf("store-path (%v) and ssh-path (%v) have to be set!", storePath, sshPath)
	}

	// create SSH path if it doesn't exist
	if _, err := os.Stat(sshPath); os.IsNotExist(err) {
		err := os.Mkdir(sshPath, 0755)

		if err != nil {
			color.Red("%sFailed to initialize SSH path!", CrossSymbol)
			return nil
		}
	}

	if resticPath == "" {
		resticPath, _ = exec.LookPath("restic")
	}
	return &models.Environment{
		StorePath:    storePath,
		SSHPath:      sshPath,
		ResticPath:   resticPath,
		KeepTypeKeys: keepTypeKeys,
	}
}
