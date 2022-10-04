package models

import "fmt"

// Environment abstracts away things like the path the .skm and .ssh folder
// which allows us to simulate them for testing.
type Environment struct {
	StorePath    string
	SSHPath      string
	ResticPath   string
	KeepTypeKeys bool
}

// SSHKey struct includes both private/public keys & isDefault flag
type SSHKey struct {
	PublicKey  string
	PrivateKey string
	IsDefault  bool
	Type       *KeyType
}

// KeyType abstracts configurations for various SSH key types like RSA and
// ED25519
type KeyType struct {
	Name                    string
	KeyBaseName             string
	SupportsVariableBitsize bool
}

// PrivateKey returns the filename used by a keytype for the private component.
func (kt KeyType) PrivateKey() string {
	return kt.KeyBaseName
}

// PublicKey returns the filename used by a keytype for the public component.
func (kt KeyType) PublicKey() string {
	return fmt.Sprintf("%s.pub", kt.KeyBaseName)
}
