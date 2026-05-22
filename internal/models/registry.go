package models

// KeyTypeRegistry is used to store all the supported key types.
type KeyTypeRegistry map[string]KeyType

// GetByFilename returns a key type object given the name of the private key's
// file. If no matching key type could be found, then the second return value
// is false.
func (r KeyTypeRegistry) GetByFilename(name string) (KeyType, bool) {
	for _, kt := range r {
		if name == kt.KeyBaseName {
			return kt, true
		}
	}
	return KeyType{}, false
}

// SupportedKeyTypes contains all key types supported by skm.
var SupportedKeyTypes KeyTypeRegistry

func init() {
	SupportedKeyTypes = KeyTypeRegistry{}
	SupportedKeyTypes["rsa"] = KeyType{
		Name:                    "rsa",
		SupportsVariableBitsize: true,
		KeyBaseName:             "id_rsa",
	}

	SupportedKeyTypes["ed25519"] = KeyType{
		Name:                    "ed25519",
		SupportsVariableBitsize: false,
		KeyBaseName:             "id_ed25519",
	}

	// FIDO2 hardware-backed keys. The on-disk file is a key handle, not the
	// actual private material — that lives on the security key itself.
	// ssh-keygen 8.2+ and libfido2 are required at create time.
	SupportedKeyTypes["ed25519-sk"] = KeyType{
		Name:                    "ed25519-sk",
		SupportsVariableBitsize: false,
		KeyBaseName:             "id_ed25519_sk",
	}

	SupportedKeyTypes["ecdsa-sk"] = KeyType{
		Name:                    "ecdsa-sk",
		SupportsVariableBitsize: false,
		KeyBaseName:             "id_ecdsa_sk",
	}
}
