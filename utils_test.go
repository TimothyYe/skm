package skm

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"
)

func getHomeDir() string {
	user, err := user.Current()
	if nil == err && user.HomeDir != "" {
		return user.HomeDir
	}
	return os.Getenv("HOME")
}

func TestParseArgs(t *testing.T) {
	os.Args = []string{"skm"}
	ParseArgs()
	os.Args = []string{"skm", "-h"}
	ParseArgs()
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
	path := parsePath("/etc/passwd")

	if path != "/etc/passwd" {
		t.Error("path are not equal")
	}

	// parse symbol link
	os.Symlink("/etc/passwd", "/tmp/passwd")
	path = parsePath("/tmp/passwd")

	if path != "/etc/passwd" {
		t.Error("path are not equal")
	}
}

func TestLoadSSHKeys(t *testing.T) {
	if _, err := os.Stat(StorePath); os.IsNotExist(err) {
		Execute("", "mkdir", "-p", StorePath)
	}

	// Create a test key
	Execute("", "mkdir", filepath.Join(StorePath, "testkey123"))
	Execute("", "touch", filepath.Join(StorePath, "testkey123", "id_rsa"))
	Execute("", "touch", filepath.Join(StorePath, "testkey123", "id_rsa.pub"))

	keyMap := LoadSSHKeys()

	// Length of key map should greater than 0
	if len(keyMap) == 0 {
		t.Error("key map should not be empty")
	}

	// cleanup
	os.RemoveAll(filepath.Join(StorePath, "testkey123"))
}

// WARNING: Make sure to backup your SSH keys before running this test case
func TestClearKey(t *testing.T) {
	ClearKey()

	PublicKeyPath := filepath.Join(SSHPath, PublicKey)
	if _, err := os.Stat(PublicKeyPath); !os.IsNotExist(err) {
		t.Error("should public key should be removed")
	}

	PrivateKeyPath := filepath.Join(SSHPath, PrivateKey)
	if _, err := os.Stat(PrivateKeyPath); !os.IsNotExist(err) {
		t.Error("should private key should be removed")
	}
}

func TestDeleteKey(t *testing.T) {
	if _, err := os.Stat(StorePath); os.IsNotExist(err) {
		Execute("", "mkdir", "-p", StorePath)
	}

	//Create a test key
	Execute("", "mkdir", filepath.Join(StorePath, "testkey123"))
	Execute("", "touch", filepath.Join(StorePath, "testkey123", "id_rsa"))
	Execute("", "touch", filepath.Join(StorePath, "testkey123", "id_rsa.pub"))

	//Construct a key
	key := SSHKey{PrivateKey: filepath.Join(StorePath, "testkey123", "id_rsa"), PublicKey: filepath.Join(StorePath, "testkey123", "id_rsa.pub")}
	//Delete key
	DeleteKey("testkey123", &key, true)

	if _, err := os.Stat(filepath.Join(StorePath, "testkey123")); !os.IsNotExist(err) {
		t.Error("key should be deleted")
	}
}

func TestLoadSingleKey(t *testing.T) {
	sshPath := filepath.Join(getHomeDir(), ".sshtest")

	if _, err := os.Stat(sshPath); os.IsNotExist(err) {
		Execute("", "mkdir", "-p", sshPath)
	}

	Execute("", "touch", filepath.Join(sshPath, "id_rsa"))
	Execute("", "touch", filepath.Join(sshPath, "id_rsa.pub"))

	key := loadSingleKey(sshPath)

	if key == nil {
		t.Error("key shouldn't be nil")
	}

	// cleanup
	os.RemoveAll(sshPath)
}

// WARNING: Make sure to backup your SSH keys before running this test case
func TestCreateLink(t *testing.T) {
	CreateLink("abc")

	PublicKeyPath := filepath.Join(SSHPath, PublicKey)
	if _, err := os.Stat(PublicKeyPath); !os.IsNotExist(err) {
		t.Error("should create symbol link for public key")
	}

	PrivateKeyPath := filepath.Join(SSHPath, PrivateKey)
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
	skmPath := filepath.Join(getHomeDir(), ".skmtest")
	os.RemoveAll(skmPath)

	if _, err := os.Stat(skmPath); os.IsNotExist(err) {
		Execute("", "mkdir", "-p", skmPath)
	}

	if ok, err := IsEmpty(skmPath); err != nil || !ok {
		t.Error("directory should be empty")
	}
}
