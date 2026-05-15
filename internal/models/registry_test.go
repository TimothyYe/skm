package models

import (
	"testing"
)

func TestKeyTypeRegistry(t *testing.T) {
	testcases := []struct {
		name  string
		input string
		found bool
	}{
		{
			name:  "rsa",
			input: "id_rsa",
			found: true,
		},
		{
			name:  "ed25519",
			input: "id_ed25519",
			found: true,
		},
		{
			name:  "",
			input: "unsupported",
			found: false,
		},
	}
	for _, testcase := range testcases {
		kt, found := SupportedKeyTypes.GetByFilename(testcase.input)
		if found != testcase.found {
			t.Fatalf("%s should have returned %v, but got %v instead", testcase.input, testcase.found, found)
		}
		if kt.Name != testcase.name {
			t.Fatalf("%s should have returned %v, but got %v instead",
				testcase.input,
				testcase.name,
				kt.Name)
		}
	}
}

func TestKeyTypeFilenames(t *testing.T) {
	testcases := []struct {
		registryKey string
		priv        string
		pub         string
	}{
		{registryKey: "rsa", priv: "id_rsa", pub: "id_rsa.pub"},
		{registryKey: "ed25519", priv: "id_ed25519", pub: "id_ed25519.pub"},
	}
	for _, tc := range testcases {
		kt, ok := SupportedKeyTypes[tc.registryKey]
		if !ok {
			t.Fatalf("registry missing entry for %q", tc.registryKey)
		}
		if got := kt.PrivateKey(); got != tc.priv {
			t.Errorf("%s PrivateKey() = %q, want %q", tc.registryKey, got, tc.priv)
		}
		if got := kt.PublicKey(); got != tc.pub {
			t.Errorf("%s PublicKey() = %q, want %q", tc.registryKey, got, tc.pub)
		}
	}
}

func TestED25519DoesNotSupportVariableBitsize(t *testing.T) {
	kt, ok := SupportedKeyTypes["ed25519"]
	if !ok {
		t.Fatal("ed25519 missing from registry")
	}
	if kt.SupportsVariableBitsize {
		t.Error("ed25519 keys are fixed-size; SupportsVariableBitsize should be false")
	}
}
