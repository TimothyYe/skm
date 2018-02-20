package skm

import "testing"

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
			t.Fatalf("%s should have returned %v, but got %v instead", testcase.input, testcase.name, kt.Name)
		}
	}
}
