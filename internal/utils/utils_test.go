package utils

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/TimothyYe/skm/internal/models"
)

var random *rand.Rand

func init() {
	source := rand.NewSource(time.Now().Unix())
	random = rand.New(source)
}

func setupTestEnvironment(t *testing.T) *models.Environment {
	testRoot := filepath.Join(os.TempDir(), fmt.Sprintf("skm-testsuite-%d", random.Int()))
	sshPath := filepath.Join(testRoot, ".ssh")
	storePath := filepath.Join(testRoot, ".skm")
	os.RemoveAll(testRoot)
	if err := os.MkdirAll(sshPath, 0700); err != nil {
		t.Fatalf("Failed to create %s: %s", sshPath, err.Error())
	}
	if err := os.MkdirAll(storePath, 0700); err != nil {
		t.Fatalf("Failed to create %s: %s", storePath, err.Error())
	}
	return &models.Environment{
		SSHPath:   sshPath,
		StorePath: storePath,
	}
}

func tearDownTestEnvironment(t *testing.T, env *models.Environment) {
	rootPath := filepath.Dir(env.SSHPath)
	os.RemoveAll(rootPath)
}

func TestExecute(t *testing.T) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		t.Errorf("get user home dir error: %v", err)
	}
	result := Execute(homedir, "ls")
	if !result {
		t.Error("should return true")
	}

	result = Execute(homedir, "aaa")
	if result {
		t.Error("should return false")
	}
}

func TestParsePath(t *testing.T) {
	file, err := os.CreateTemp("", "path1")
	if err != nil {
		t.Error(err)
	}

	path1 := ParsePath(file.Name())
	defer os.Remove(path1)

	path2 := path.Join(os.TempDir(), "path2")
	defer os.Remove(path2)

	// parse symbol link
	if err := os.Symlink(path1, path2); err != nil {
		t.Error("failed to parse symbol link")
	}

	if path1 != ParsePath(path2) {
		t.Error("path are not equal")
	}
}

func TestLoadSSHKeys(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	// Create a test key
	Execute("", "mkdir", filepath.Join(env.StorePath, "testkey123"))
	Execute("", "touch", filepath.Join(env.StorePath, "testkey123", "id_rsa"))
	Execute("", "touch", filepath.Join(env.StorePath, "testkey123", "id_rsa.pub"))

	keyMap := LoadSSHKeys(env)

	// Length of key map should greater than 0
	if len(keyMap) == 0 {
		t.Error("key map should not be empty")
	}

	// cleanup
	os.RemoveAll(filepath.Join(env.StorePath, "testkey123"))
}

func TestClearKey(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	publicKeyPath := filepath.Join(env.SSHPath, PublicKey)
	privateKeyPath := filepath.Join(env.SSHPath, PrivateKey)
	ed25519PublicKeyPath := filepath.Join(env.SSHPath, Ed25519PublicKey)
	ed25519PrivateKeyPath := filepath.Join(env.SSHPath, Ed25519PrivateKey)
	createFile := func(p string) {
		f, err := os.Create(p)
		if err != nil {
			t.Errorf("Create file error: %v", err)
		}
		f.Close()
	}
	createFile(publicKeyPath)
	createFile(privateKeyPath)
	createFile(ed25519PublicKeyPath)
	createFile(ed25519PrivateKeyPath)

	// clear key with ssh type
	env.KeepTypeKeys = true
	ClearKey(env, "ed25519")
	if _, err := os.Stat(publicKeyPath); os.IsNotExist(err) {
		t.Error("public key should not be removed")
	}

	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		t.Error("private key should not be removed")
	}

	if _, err := os.Stat(ed25519PublicKeyPath); !os.IsNotExist(err) {
		t.Error("ed25519 public key should be removed")
	}

	if _, err := os.Stat(ed25519PrivateKeyPath); !os.IsNotExist(err) {
		t.Error("ed25519 private key should be removed")
	}

	// clear all keys
	createFile(ed25519PublicKeyPath)
	createFile(ed25519PrivateKeyPath)
	env.KeepTypeKeys = false
	ClearKey(env, "")
	if _, err := os.Stat(publicKeyPath); !os.IsNotExist(err) {
		t.Error("public key should be removed")
	}

	if _, err := os.Stat(privateKeyPath); !os.IsNotExist(err) {
		t.Error("private key should be removed")
	}

	if _, err := os.Stat(ed25519PublicKeyPath); !os.IsNotExist(err) {
		t.Error("ed25519 public key should be removed")
	}

	if _, err := os.Stat(ed25519PrivateKeyPath); !os.IsNotExist(err) {
		t.Error("ed25519 private key should be removed")
	}
}

func TestDeleteKey(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	//Create a test key
	Execute("", "mkdir", filepath.Join(env.StorePath, "testkey123"))
	Execute("", "touch", filepath.Join(env.StorePath, "testkey123", "id_rsa"))
	Execute("", "touch", filepath.Join(env.StorePath, "testkey123", "id_rsa.pub"))

	//Construct a key
	key := models.SSHKey{
		PrivateKey: filepath.Join(env.StorePath, "testkey123", "id_rsa"),
		PublicKey:  filepath.Join(env.StorePath, "testkey123", "id_rsa.pub"),
		Type: &models.KeyType{
			Name: "rsa",
		},
	}
	//Delete key
	DeleteKey("testkey123", &key, env, true)

	if _, err := os.Stat(filepath.Join(env.StorePath, "testkey123")); !os.IsNotExist(err) {
		t.Error("key should be deleted")
	}
}

func TestLoadSingleKey(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	Execute("", "touch", filepath.Join(env.SSHPath, "id_rsa"))
	Execute("", "touch", filepath.Join(env.SSHPath, "id_rsa.pub"))

	key := loadSingleKey(env.SSHPath, env)

	if key == nil {
		t.Error("key shouldn't be nil")
	}
}

func TestCreateLink(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	CreateLink("abc", nil, env)

	PublicKeyPath := filepath.Join(env.SSHPath, PublicKey)
	if _, err := os.Stat(PublicKeyPath); !os.IsNotExist(err) {
		t.Error("should create symbol link for public key")
	}

	PrivateKeyPath := filepath.Join(env.SSHPath, PrivateKey)
	if _, err := os.Stat(PrivateKeyPath); !os.IsNotExist(err) {
		t.Error("should create symbol link for private key")
	}
}

func TestGetBakFileName(t *testing.T) {
	fileName := GetBakFileName()

	if fileName == "" {
		t.Error("file name shouldn't be empty")
	}
}

func TestIsEmpty(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)
	if ok, err := IsEmpty(env.StorePath); err != nil || !ok {
		t.Error("directory should be empty")
	}
}
