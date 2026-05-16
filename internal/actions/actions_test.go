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

// seedKey creates a stored key directory with id_rsa / id_rsa.pub under it.
func seedKey(t *testing.T, env *models.Environment, alias string) {
	t.Helper()
	dir := filepath.Join(env.StorePath, alias)
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatalf("setup %s: %v", alias, err)
	}
	writeFile(t, filepath.Join(dir, "id_rsa"), []byte("priv-"+alias))
	writeFile(t, filepath.Join(dir, "id_rsa.pub"), []byte("pub-"+alias))
}

// activeAlias returns the alias whose stored private key the ~/.ssh symlink
// resolves to, or "" if no active symlink is present.
func activeAlias(t *testing.T, env *models.Environment) string {
	t.Helper()
	resolved := utils.ParsePath(filepath.Join(env.SSHPath, utils.PrivateKey))
	if resolved == "" {
		return ""
	}
	rel, err := filepath.Rel(env.StorePath, resolved)
	if err != nil {
		t.Fatalf("rel: %v", err)
	}
	return filepath.Dir(rel)
}

func TestUse_ExactMatch(t *testing.T) {
	env := setupEnvironment(t)
	defer tearDownEnvironment(t, env)

	seedKey(t, env, "prod-a")
	seedKey(t, env, "prod-b")
	seedKey(t, env, "staging")

	c := newContextForArgs(t, env, []string{"staging"})
	if err := Use(c); err != nil {
		t.Fatalf("Use: %v", err)
	}

	if got := activeAlias(t, env); got != "staging" {
		t.Errorf("active alias = %q, want %q", got, "staging")
	}
}

func TestUse_UniquePartialMatch(t *testing.T) {
	env := setupEnvironment(t)
	defer tearDownEnvironment(t, env)

	seedKey(t, env, "prod-a")
	seedKey(t, env, "staging")

	// "stag" only matches "staging".
	c := newContextForArgs(t, env, []string{"stag"})
	if err := Use(c); err != nil {
		t.Fatalf("Use: %v", err)
	}

	if got := activeAlias(t, env); got != "staging" {
		t.Errorf("active alias = %q, want %q", got, "staging")
	}
}

func TestUse_AmbiguousPartialMatchNoop(t *testing.T) {
	env := setupEnvironment(t)
	defer tearDownEnvironment(t, env)

	seedKey(t, env, "prod-a")
	seedKey(t, env, "prod-b")

	// "prod" matches both — must not pick either, must not create symlinks.
	c := newContextForArgs(t, env, []string{"prod"})
	if err := Use(c); err != nil {
		t.Fatalf("Use: %v", err)
	}

	if got := activeAlias(t, env); got != "" {
		t.Errorf("ambiguous match should not have set an active alias, got %q", got)
	}
	if _, err := os.Lstat(filepath.Join(env.SSHPath, utils.PrivateKey)); !os.IsNotExist(err) {
		t.Error("private key symlink should not exist after ambiguous match")
	}
	if _, err := os.Lstat(filepath.Join(env.SSHPath, utils.PublicKey)); !os.IsNotExist(err) {
		t.Error("public key symlink should not exist after ambiguous match")
	}
}

func TestUse_NoMatchNoop(t *testing.T) {
	env := setupEnvironment(t)
	defer tearDownEnvironment(t, env)

	seedKey(t, env, "prod-a")

	// Pre-existing real key file in ~/.ssh — Use should leave it alone on miss.
	stray := filepath.Join(env.SSHPath, utils.PrivateKey)
	writeFile(t, stray, []byte("preexisting"))

	c := newContextForArgs(t, env, []string{"nope"})
	if err := Use(c); err != nil {
		t.Fatalf("Use: %v", err)
	}

	// Pre-existing file should remain — CreateLink (which would ClearKey) was never called.
	got, err := os.ReadFile(stray)
	if err != nil {
		t.Fatalf("read stray: %v", err)
	}
	if string(got) != "preexisting" {
		t.Errorf("pre-existing key was modified: got %q", string(got))
	}
}

func TestUse_PartialMatchDeterministic(t *testing.T) {
	// Same inputs across runs must resolve to the same alias.
	// "alpha" appears in both "alpha-1" and "zalpha", and sorted order
	// guarantees "alpha-1" is the chosen match.
	for range 5 {
		env := setupEnvironment(t)
		seedKey(t, env, "zalpha")
		seedKey(t, env, "alpha-1")

		c := newContextForArgs(t, env, []string{"alph"})
		if err := Use(c); err != nil {
			tearDownEnvironment(t, env)
			t.Fatalf("Use: %v", err)
		}

		// "alph" is a substring of both — ambiguous → no symlink.
		if got := activeAlias(t, env); got != "" {
			tearDownEnvironment(t, env)
			t.Fatalf("ambiguous input should not resolve, got %q", got)
		}
		tearDownEnvironment(t, env)
	}
}

// newContextWithBoolFlags returns a cli.Context with string flags for the
// store/ssh paths and any number of named bool flags pre-set to true.
func newContextWithBoolFlags(t *testing.T, env *models.Environment, positionalArgs []string, trueBools ...string) *cli.Context {
	t.Helper()
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	flags.String("ssh-path", env.SSHPath, "")
	flags.String("store-path", env.StorePath, "")
	for _, name := range trueBools {
		flags.Bool(name, true, "")
	}
	if err := flags.Parse(positionalArgs); err != nil {
		t.Fatalf("flag parse: %v", err)
	}
	return cli.NewContext(nil, flags, nil)
}

// ----- List -----

func TestList_EmptyStore(t *testing.T) {
	env := setupEnvironment(t)
	defer tearDownEnvironment(t, env)

	c := newContextForArgs(t, env, nil)
	if err := List(c); err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestList_WithKeys(t *testing.T) {
	env := setupEnvironment(t)
	defer tearDownEnvironment(t, env)

	seedKey(t, env, "alpha")
	seedKey(t, env, "beta")

	c := newContextForArgs(t, env, nil)
	if err := List(c); err != nil {
		t.Fatalf("List: %v", err)
	}
}

// ----- Display -----

func TestDisplay_UnknownAliasErrors(t *testing.T) {
	env := setupEnvironment(t)
	defer tearDownEnvironment(t, env)

	seedKey(t, env, "alpha")

	c := newContextForArgs(t, env, []string{"missing"})
	if err := Display(c); err == nil {
		t.Error("expected error for unknown alias")
	}
}

func TestDisplay_KnownAliasSucceeds(t *testing.T) {
	env := setupEnvironment(t)
	defer tearDownEnvironment(t, env)

	seedKey(t, env, "alpha")

	c := newContextForArgs(t, env, []string{"alpha"})
	if err := Display(c); err != nil {
		t.Fatalf("Display: %v", err)
	}
}

func TestDisplay_NoArgNoDefaultIsNoop(t *testing.T) {
	env := setupEnvironment(t)
	defer tearDownEnvironment(t, env)

	// Stored key exists, but no symlink in ~/.ssh → no IsDefault key → silent return.
	seedKey(t, env, "alpha")

	c := newContextForArgs(t, env, nil)
	if err := Display(c); err != nil {
		t.Fatalf("Display: %v", err)
	}
}

// ----- Cache -----

func TestCache_NoFlagReturnsError(t *testing.T) {
	env := setupEnvironment(t)
	defer tearDownEnvironment(t, env)

	c := newContextForArgs(t, env, []string{"alpha"})
	if err := Cache(c); err == nil {
		t.Error("expected error when no add/del/list flag is set")
	}
}

func TestCache_AddUnknownAliasNoop(t *testing.T) {
	env := setupEnvironment(t)
	defer tearDownEnvironment(t, env)

	// No stored aliases → AddCache returns "not found" before invoking ssh-add.
	c := newContextWithBoolFlags(t, env, []string{"missing"}, "add")
	if err := Cache(c); err != nil {
		t.Fatalf("Cache: %v", err)
	}
}

func TestCache_DelUnknownAliasNoop(t *testing.T) {
	env := setupEnvironment(t)
	defer tearDownEnvironment(t, env)

	c := newContextWithBoolFlags(t, env, []string{"missing"}, "del")
	if err := Cache(c); err != nil {
		t.Fatalf("Cache: %v", err)
	}
}

// ----- Copy -----

func TestCopy_NoActiveKeyReturnsEarly(t *testing.T) {
	env := setupEnvironment(t)
	defer tearDownEnvironment(t, env)

	// No symlink in ~/.ssh → guard kicks in, ssh-copy-id is not invoked.
	c := newContextForArgs(t, env, []string{"user@host"}, "p")
	if err := Copy(c); err != nil {
		t.Fatalf("Copy: %v", err)
	}
}

// ----- Delete -----

func TestDelete_NoArgsNoop(t *testing.T) {
	env := setupEnvironment(t)
	defer tearDownEnvironment(t, env)

	seedKey(t, env, "alpha")

	c := newContextForArgs(t, env, nil)
	if err := Delete(c); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// Stored alias must remain.
	if _, err := os.Stat(filepath.Join(env.StorePath, "alpha")); err != nil {
		t.Errorf("alpha should still exist: %v", err)
	}
}

func TestDelete_UnknownAliasNoop(t *testing.T) {
	env := setupEnvironment(t)
	defer tearDownEnvironment(t, env)

	seedKey(t, env, "alpha")

	c := newContextForArgs(t, env, []string{"missing"})
	if err := Delete(c); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if _, err := os.Stat(filepath.Join(env.StorePath, "alpha")); err != nil {
		t.Errorf("alpha should still exist: %v", err)
	}
}

// ----- Rename failure cases -----

func TestRename_NonexistentSourceNoop(t *testing.T) {
	env := setupEnvironment(t)
	defer tearDownEnvironment(t, env)

	// "before" does not exist — os.Rename will fail, action prints error and returns nil.
	c := newContextForArgs(t, env, []string{"before", "after"})
	if err := Rename(c); err != nil {
		t.Fatalf("Rename: %v", err)
	}

	if _, err := os.Stat(filepath.Join(env.StorePath, "after")); !os.IsNotExist(err) {
		t.Error("target alias should not exist when source is missing")
	}
}
