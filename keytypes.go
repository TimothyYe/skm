package skm

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
		SupportsVariableBitsize: true,
		KeyBaseName:             "id_ed25519",
	}
}
