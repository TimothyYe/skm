package actions

import (
	"flag"
	"fmt"
	"github.com/TimothyYe/skm/internal/models"
	"github.com/TimothyYe/skm/internal/utils"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	cli "gopkg.in/urfave/cli.v1"
)

var random *rand.Rand

func init() {
	random = rand.New(rand.NewSource(time.Now().Unix()))
}

func TestInitAction(t *testing.T) {
	testcases := []struct {
		code  string
		info  string
		err   bool
		check func(t *testing.T, env *models.Environment)
		setup func(t *testing.T, env *models.Environment)
	}{
		{
			code: "no-preexisting-key",
			info: "If the .ssh folder is empty, then no default profile should be created",
			err:  false,
			check: func(t *testing.T, env *models.Environment) {
				assertEmptyDir(t, env.StorePath, "no-preexisting-key")
			},
		},
		{
			code: "rsa-key",
			info: "If the .ssh folder contains an RSA key-pair, it should be copied into the default profile",
			setup: func(t *testing.T, env *models.Environment) {
				os.RemoveAll(env.StorePath)
				writeFile(t, filepath.Join(env.SSHPath, "id_rsa"), []byte("private"))
				writeFile(t, filepath.Join(env.SSHPath, "id_rsa.pub"), []byte("public"))
			},
			err: false,
			check: func(t *testing.T, env *models.Environment) {
				assertFileExists(t, filepath.Join(env.StorePath, "default", "id_rsa"), "rsa")
				assertFileExists(t, filepath.Join(env.StorePath, "default", "id_rsa.pub"), "rsa")
			},
		},
		{
			code: "ed25519-key",
			info: "If the .ssh folder contains an ED25519 key-pair, it should be copied into the default profile",
			setup: func(t *testing.T, env *models.Environment) {
				os.RemoveAll(env.StorePath)
				writeFile(t, filepath.Join(env.SSHPath, "id_ed25519"), []byte("private"))
				writeFile(t, filepath.Join(env.SSHPath, "id_ed25519.pub"), []byte("public"))
			},
			err: false,
			check: func(t *testing.T, env *models.Environment) {
				assertFileExists(t, filepath.Join(env.StorePath, "default", "id_ed25519"), "ed25519")
				assertFileExists(t, filepath.Join(env.StorePath, "default", "id_ed25519.pub"), "ed25519")
			},
		},
	}

	for _, testcase := range testcases {
		env := setupEnvironment(t)
		flags := flag.NewFlagSet("", flag.ContinueOnError)
		flags.String("ssh-path", env.SSHPath, "")
		flags.String("store-path", env.StorePath, "")
		if testcase.setup != nil {
			testcase.setup(t, env)
		}
		if !t.Failed() {
			c := cli.NewContext(nil, flags, nil)
			err := Initialize(c)
			if err == nil && testcase.err {
				t.Errorf("[%s] should have returned an error", testcase.code)
			}
			if err != nil && !testcase.err {
				t.Errorf("[%s] produced an unexpected error: %s", testcase.code, err.Error())
			}
			if testcase.check != nil {
				testcase.check(t, env)
			}
		}
		tearDownEnvironment(t, env)
		if t.Failed() {
			break
		}
	}
}

func setupEnvironment(t *testing.T) *models.Environment {
	rootFolder := filepath.Join(os.TempDir(), fmt.Sprintf("skm-testdir-%d", random.Int()))
	storePath := filepath.Join(rootFolder, ".skm")
	sshPath := filepath.Join(rootFolder, ".ssh")
	if err := os.MkdirAll(storePath, 0700); err != nil {
		t.Errorf("Failed to create store path: %s", err.Error())
		return nil
	}
	if err := os.MkdirAll(sshPath, 0700); err != nil {
		t.Errorf("Failed to create ssh path: %s", err.Error())
		return nil
	}
	return &models.Environment{
		StorePath: storePath,
		SSHPath:   sshPath,
	}
}

func tearDownEnvironment(t *testing.T, env *models.Environment) {
	rootFolder := filepath.Dir(env.StorePath)
	os.RemoveAll(rootFolder)
}

func assertFileExists(t *testing.T, path string, code string) {
	_, err := os.Stat(path)
	if err != nil {
		t.Fatalf("[%s] Failed to assert file existence: %s (err: %v)", code, path, err)
	}
}

func assertEmptyDir(t *testing.T, path string, code string) {
	empty, err := utils.IsEmpty(path)
	if err != nil || !empty {
		t.Fatalf("[%s] Failed to assert empty directory %s: %v", code, path, err)
	}
}

func writeFile(t *testing.T, path string, data []byte) {
	if t.Failed() {
		return
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("Failed to write %s: %s", path, err.Error())
	}
}

func newContextForArgs(t *testing.T, env *models.Environment, positionalArgs []string, stringFlags ...string) *cli.Context {
	t.Helper()
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	flags.String("ssh-path", env.SSHPath, "")
	flags.String("store-path", env.StorePath, "")
	for _, name := range stringFlags {
		flags.String(name, "", "")
	}
	if err := flags.Parse(positionalArgs); err != nil {
		t.Fatalf("flag parse: %v", err)
	}
	return cli.NewContext(nil, flags, nil)
}

func TestRename(t *testing.T) {
	env := setupEnvironment(t)
	defer tearDownEnvironment(t, env)

	src := filepath.Join(env.StorePath, "before")
	if err := os.Mkdir(src, 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}
	writeFile(t, filepath.Join(src, "id_rsa"), []byte("priv"))

	c := newContextForArgs(t, env, []string{"before", "after"})
	if err := Rename(c); err != nil {
		t.Fatalf("Rename: %v", err)
	}

	if _, err := os.Stat(filepath.Join(env.StorePath, "after", "id_rsa")); err != nil {
		t.Errorf("expected renamed dir to contain the key: %v", err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("old alias directory should be gone")
	}
}

func TestRename_MissingArgsNoop(t *testing.T) {
	env := setupEnvironment(t)
	defer tearDownEnvironment(t, env)

	src := filepath.Join(env.StorePath, "before")
	if err := os.Mkdir(src, 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Only one positional arg - Rename should print an error and leave state alone.
	c := newContextForArgs(t, env, []string{"before"})
	if err := Rename(c); err != nil {
		t.Fatalf("Rename: %v", err)
	}

	if _, err := os.Stat(src); err != nil {
		t.Errorf("source alias should remain untouched: %v", err)
	}
}

func TestCreate_DuplicateAliasSkipsKeygen(t *testing.T) {
	env := setupEnvironment(t)
	defer tearDownEnvironment(t, env)

	alias := "dup"
	dir := filepath.Join(env.StorePath, alias)
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}
	priv := filepath.Join(dir, "id_rsa")
	pub := filepath.Join(dir, "id_rsa.pub")
	writeFile(t, priv, []byte("PRIV-ORIGINAL"))
	writeFile(t, pub, []byte("PUB-ORIGINAL"))

	c := newContextForArgs(t, env, []string{alias}, "t", "b", "C")
	if err := Create(c); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Files must remain byte-identical: ssh-keygen was not invoked.
	for path, want := range map[string]string{priv: "PRIV-ORIGINAL", pub: "PUB-ORIGINAL"} {
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if string(got) != want {
			t.Errorf("%s was overwritten: got %q, want %q", path, string(got), want)
		}
	}
}

func TestCreate_MissingAliasNoop(t *testing.T) {
	env := setupEnvironment(t)
	defer tearDownEnvironment(t, env)

	c := newContextForArgs(t, env, nil, "t", "b", "C")
	if err := Create(c); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// No alias dirs should have been created.
	entries, err := os.ReadDir(env.StorePath)
	if err != nil {
		t.Fatalf("read store: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected store to be empty, got %d entries", len(entries))
	}
}
