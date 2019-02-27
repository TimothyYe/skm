package skm

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"math/rand"
)

var random *rand.Rand

func init() {
	source := rand.NewSource(time.Now().Unix())
	random = rand.New(source)
}

func setupTestEnvironment(t *testing.T) *Environment {
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
	return &Environment{
		SSHPath:   sshPath,
		StorePath: storePath,
	}
}

func tearDownTestEnvironment(t *testing.T, env *Environment) {
	rootPath := filepath.Dir(env.SSHPath)
	os.RemoveAll(rootPath)
}

func TestExecute(t *testing.T) {
	result := Execute("/home", "ls")
	if !result {
		t.Error("should return true")
	}

	result = Execute("/home", "aaa")
	if result {
		t.Error("should return false")
	}
}

func TestParsePath(t *testing.T) {
	path := ParsePath("/etc/passwd")

	if path != "/etc/passwd" {
		t.Error("path are not equal")
	}

	// parse symbol link
	if err := os.Symlink("/etc/passwd", "/tmp/passwd"); err != nil {
		t.Error("failed to parse symbol link")
	}
	path = ParsePath("/tmp/passwd")

	if path != "/etc/passwd" {
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
	ClearKey(env)

	PublicKeyPath := filepath.Join(env.SSHPath, PublicKey)
	if _, err := os.Stat(PublicKeyPath); !os.IsNotExist(err) {
		t.Error("should public key should be removed")
	}

	PrivateKeyPath := filepath.Join(env.SSHPath, PrivateKey)
	if _, err := os.Stat(PrivateKeyPath); !os.IsNotExist(err) {
		t.Error("should private key should be removed")
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
	key := SSHKey{PrivateKey: filepath.Join(env.StorePath, "testkey123", "id_rsa"), PublicKey: filepath.Join(env.StorePath, "testkey123", "id_rsa.pub")}
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
