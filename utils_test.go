package skm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseArgs(t *testing.T) {
	os.Args = append(os.Args, "abc")
	ParseArgs()
}

func TestExecute(t *testing.T) {
	result := Execute("/home", "ls")
	if !result {
		t.Error("should return true")
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
