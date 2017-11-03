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
	os.Args = append(os.Args, "skm")
	ParseArgs()
	os.Args = append(os.Args, "-h")
	ParseArgs()
}

func TestExecute(t *testing.T) {
	result := Execute("/home", "ls")
	if !result {
		t.Error("should return true")
	}

	result = Execute("/home", "aaa")
	if result {
		t.Error("shoudl return false")
	}
}

func TestParsePath(t *testing.T) {
	path := parsePath("/etc/passwd")

	if path != "/etc/passwd" {
		t.Error("path are not equal")
	}
}

func TestLoadSSHKeys(t *testing.T) {
	keyMap := LoadSSHKeys()

	if _, err := os.Stat(StorePath); !os.IsNotExist(err) {
		if _, err := os.Stat(filepath.Join(StorePath, DefaultKey)); !os.IsNotExist(err) {
			// Length of key map should greater than 0
			if len(keyMap) == 0 {
				t.Error("key map should not be empty")
			}
		}
	} else {
		// Should return empty keyMap since store path doesn't exist
		if len(keyMap) > 0 {
			t.Error("key map should be empty")
		}
	}
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

func TestLoadSingleKey(t *testing.T) {
	sshPath := filepath.Join(getHomeDir(), ".ssh")
	if _, err := os.Stat(filepath.Join(getHomeDir(), ".ssh", "id_rsa")); os.IsExist(err) {
		key := loadSingleKey(sshPath)

		if key == nil {
			t.Error("key shouldn't be nil")
		}
	} else {
		fileInfo, err := os.Lstat(sshPath)
		if err == nil && fileInfo.Mode()&os.ModeSymlink != 0 {
			key := loadSingleKey(sshPath)

			if key != nil {
				t.Error("key shoudl be nil")
			}
		}
	}
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
