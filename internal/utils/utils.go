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
	// DefaultKey is the default alias name of SSH key
	DefaultKey = "default"

	// HookName is the legacy single-file hook (treated as post-use).
	HookName = "hook"
	// HooksDir is the directory holding event-named hook scripts.
	HooksDir = "hooks"
)

// Hook event names.
const (
	EventPostUse    = "post-use"
	EventPostCreate = "post-create"
	EventPreDelete  = "pre-delete"
	EventPostCopy   = "post-copy"
)

// KnownHookEvents lists every event SKM may fire. Used by `skm hook ls` and tests.
var KnownHookEvents = []string{
	EventPostUse,
	EventPostCreate,
	EventPreDelete,
	EventPostCopy,
}

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
func ClearKey(env *models.Environment) error {
	for _, kt := range models.SupportedKeyTypes {
		// Remove private key if exists
		privateKeyPath := filepath.Join(env.SSHPath, kt.PrivateKey())
		if err := os.Remove(privateKeyPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove %s: %w", privateKeyPath, err)
		}

		// Remove public key if exists
		publicKeyPath := filepath.Join(env.SSHPath, kt.PublicKey())
		if err := os.Remove(publicKeyPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove %s: %w", publicKeyPath, err)
		}
	}
	return nil
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
		if err := RunHook(EventPreDelete, alias, env); err != nil {
			color.Red("%spre-delete hook aborted deletion of [%s]: %s", CrossSymbol, alias, err.Error())
			return
		}

		if inUse {
			if err := ClearKey(env); err != nil {
				color.Red("%s%s", CrossSymbol, err.Error())
				return
			}
		}

		//Remove specified key by alias name
		if err := os.RemoveAll(filepath.Join(env.StorePath, alias)); err == nil {
			color.Green("%sSSH key [%s] deleted!", CheckSymbol, alias)
		} else {
			color.Red("%sFailed to delete SSH key [%s]!", CrossSymbol, alias)
		}
	}
}

// RunHook executes any hook scripts registered for the given event.
//
// Lookup order: global ~/<store>/hooks/<event> first, then per-key
// ~/<store>/<alias>/hooks/<event>. For backward compatibility, the legacy
// per-key ~/<store>/<alias>/hook file fires for the post-use event.
//
// Scripts receive the alias as argv[1] (legacy contract) and a set of SKM_*
// environment variables. Additional event-specific context can be supplied as
// key,value pairs in extraEnv.
//
// For pre-* events, the first non-zero exit returns an error so the caller can
// abort. For post-* events, hook failures are logged but never propagate.
func RunHook(event, alias string, env *models.Environment, extraEnv ...string) error {
	paths := HookPaths(event, alias, env)
	if len(paths) == 0 {
		return nil
	}

	envVars := buildHookEnv(event, alias, env, extraEnv)
	abortable := strings.HasPrefix(event, "pre-")

	for _, p := range paths {
		if err := runHookScript(p, alias, envVars); err != nil {
			if abortable {
				return fmt.Errorf("%s hook %s: %w", event, p, err)
			}
			color.Yellow("⚠ %s hook %s failed: %s", event, p, err)
		}
	}
	return nil
}

// HookPaths returns the ordered list of executable hook scripts for the given
// event and alias (global first, then per-key, then legacy fallback).
func HookPaths(event, alias string, env *models.Environment) []string {
	var paths []string
	if p := filepath.Join(env.StorePath, HooksDir, event); isExecutableFile(p) {
		paths = append(paths, p)
	}
	if alias != "" {
		if p := filepath.Join(env.StorePath, alias, HooksDir, event); isExecutableFile(p) {
			paths = append(paths, p)
		}
		if event == EventPostUse {
			if p := filepath.Join(env.StorePath, alias, HookName); isExecutableFile(p) {
				paths = append(paths, p)
			}
		}
	}
	return paths
}

func isExecutableFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	return info.Mode()&0111 != 0
}

func buildHookEnv(event, alias string, env *models.Environment, extraEnv []string) []string {
	out := append(os.Environ(),
		"SKM_EVENT="+event,
		"SKM_ALIAS="+alias,
		"SKM_STORE_PATH="+env.StorePath,
		"SKM_SSH_PATH="+env.SSHPath,
	)
	if alias != "" {
		if keys := LoadSSHKeys(env); keys != nil {
			if k, ok := keys[alias]; ok {
				if k.Type != nil {
					out = append(out, "SKM_KEY_TYPE="+k.Type.Name)
				}
				out = append(out,
					"SKM_PRIVATE_KEY="+k.PrivateKey,
					"SKM_PUBLIC_KEY="+k.PublicKey,
				)
			}
		}
	}
	for i := 0; i+1 < len(extraEnv); i += 2 {
		out = append(out, extraEnv[i]+"="+extraEnv[i+1])
	}
	return out
}

func runHookScript(path, alias string, envVars []string) error {
	cmd := exec.Command(path, alias)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = envVars
	return cmd.Run()
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
		return errors.New("Failed to add SSH key to cache")
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
func CreateLink(alias string, keyMap map[string]*models.SSHKey, env *models.Environment) error {
	if err := ClearKey(env); err != nil {
		return err
	}

	key, found := keyMap[alias]

	if !found {
		return fmt.Errorf("SSH key [%s] not found", alias)
	}

	relStorePath, err := filepath.Rel(env.SSHPath, env.StorePath)
	if err != nil {
		return fmt.Errorf("failed to find rel store path: %w", err)
	}

	//Create symlink for private key
	if err := os.Symlink(filepath.Join(relStorePath, alias, key.Type.PrivateKey()), filepath.Join(env.SSHPath, key.Type.PrivateKey())); err != nil {
		return fmt.Errorf("failed to create symbol link for private key: %w", err)
	}

	//Create symlink for public key
	if err := os.Symlink(filepath.Join(relStorePath, alias, key.Type.PublicKey()), filepath.Join(env.SSHPath, key.Type.PublicKey())); err != nil {
		return fmt.Errorf("failed to create symbol link for public key: %w", err)
	}

	return nil
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
		color.Red("%sfilepath.Walk() returned %v", CrossSymbol, err)
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
			color.Red("%s%s", CrossSymbol, err.Error())
			return ""
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
		color.Red("%sfilepath.Walk() returned %v", CrossSymbol, err)
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
	if storePath == "" || sshPath == "" {
		log.Fatalf("store-path (%v) and ssh-path (%v) have to be set!", storePath, sshPath)
	}

	// create SSH path if it doesn't exist
	if _, err := os.Stat(sshPath); os.IsNotExist(err) {
		err := os.Mkdir(sshPath, 0700)

		if err != nil {
			color.Red("%sFailed to initialize SSH path!", CrossSymbol)
			return nil
		}
	}

	if resticPath == "" {
		resticPath, _ = exec.LookPath("restic")
	}
	return &models.Environment{
		StorePath:  storePath,
		SSHPath:    sshPath,
		ResticPath: resticPath,
	}
}
