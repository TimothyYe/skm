package main

//SSHKey struct includes both private/public keys & isDefault flag
type SSHKey struct {
	PublicKey  string
	PrivateKey string
	IsDefault  bool
}
