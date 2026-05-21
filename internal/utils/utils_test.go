package utils

import (
	"fmt"
	"github.com/TimothyYe/skm/internal/models"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"math/rand"
)

var random *rand.Rand

func init() {
	source := rand.NewSource(time.Now().Unix())
	random = rand.New(source)
}

func setupTestEnvironment(t *testing.T) *models.Environment {
	testRoot := filepath.Join(os.TempDir(), fmt.Sprintf("skm-testsuite-%d", random.Int()))
	sshPath := filepath.Join(testRoot, ".ssh")
	storePath := filepath.Join(testRoot, ".skm")
	os.RemoveAll(testRoot)
	if err := os.MkdirAll(sshPath, 0700); err != nil {
		t.Fatalf("Failed to create %s: %s", sshPath, err.Error())
	}
	if err := os.MkdirAll(storePath, 0700); err != nil {
		t.Fatalf("Failed to create %s: %s", storePath, err.Error())
	}
	return &models.Environment{
		SSHPath:   sshPath,
		StorePath: storePath,
	}
}

func tearDownTestEnvironment(t *testing.T, env *models.Environment) {
	rootPath := filepath.Dir(env.SSHPath)
	os.RemoveAll(rootPath)
}

func TestExecute(t *testing.T) {
	result := Execute("/home", "ls")
	if !result {
		t.Error("should return true")
	}

	result = Execute("/home", "aaa")
	if result {
		t.Error("should return false")
	}
}

func TestParsePath(t *testing.T) {
	path := ParsePath("/etc/passwd")

	if path != "/etc/passwd" {
		t.Error("path are not equal")
	}

	// parse symbol link via a unique tmp path so re-runs don't collide
	linkPath := filepath.Join(t.TempDir(), "passwd-link")
	if err := os.Symlink("/etc/passwd", linkPath); err != nil {
		t.Fatalf("failed to create symbol link: %v", err)
	}
	path = ParsePath(linkPath)

	if path != "/etc/passwd" {
		t.Errorf("expected /etc/passwd, got %q", path)
	}
}

func TestParsePath_NonExistent(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	if got := ParsePath(missing); got != "" {
		t.Errorf("expected empty string for non-existent path, got %q", got)
	}
}

func TestParsePath_RegularFile(t *testing.T) {
	regular := filepath.Join(t.TempDir(), "regular")
	if err := os.WriteFile(regular, []byte("data"), 0600); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if got := ParsePath(regular); got != regular {
		t.Errorf("regular file should be returned unchanged, got %q", got)
	}
}

func TestLoadSSHKeys(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	// Create a test key
	Execute("", "mkdir", filepath.Join(env.StorePath, "testkey123"))
	Execute("", "touch", filepath.Join(env.StorePath, "testkey123", "id_rsa"))
	Execute("", "touch", filepath.Join(env.StorePath, "testkey123", "id_rsa.pub"))

	keyMap := LoadSSHKeys(env)

	// Length of key map should greater than 0
	if len(keyMap) == 0 {
		t.Error("key map should not be empty")
	}

	// cleanup
	os.RemoveAll(filepath.Join(env.StorePath, "testkey123"))
}

func TestClearKey(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)
	if err := ClearKey(env); err != nil {
		t.Fatalf("ClearKey returned error: %v", err)
	}

	PublicKeyPath := filepath.Join(env.SSHPath, PublicKey)
	if _, err := os.Stat(PublicKeyPath); !os.IsNotExist(err) {
		t.Error("should public key should be removed")
	}

	PrivateKeyPath := filepath.Join(env.SSHPath, PrivateKey)
	if _, err := os.Stat(PrivateKeyPath); !os.IsNotExist(err) {
		t.Error("should private key should be removed")
	}
}

func TestDeleteKey(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	//Create a test key
	Execute("", "mkdir", filepath.Join(env.StorePath, "testkey123"))
	Execute("", "touch", filepath.Join(env.StorePath, "testkey123", "id_rsa"))
	Execute("", "touch", filepath.Join(env.StorePath, "testkey123", "id_rsa.pub"))

	//Construct a key
	key := models.SSHKey{PrivateKey: filepath.Join(env.StorePath, "testkey123", "id_rsa"), PublicKey: filepath.Join(env.StorePath, "testkey123", "id_rsa.pub")}
	//Delete key
	DeleteKey("testkey123", &key, env, DeleteOptions{testMode: true})

	if _, err := os.Stat(filepath.Join(env.StorePath, "testkey123")); !os.IsNotExist(err) {
		t.Error("key should be deleted")
	}
}

func TestLoadSingleKey(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	Execute("", "touch", filepath.Join(env.SSHPath, "id_rsa"))
	Execute("", "touch", filepath.Join(env.SSHPath, "id_rsa.pub"))

	key := loadSingleKey(env.SSHPath, env)

	if key == nil {
		t.Error("key shouldn't be nil")
	}
}

func TestCreateLink(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	if err := CreateLink("abc", nil, env); err == nil {
		t.Error("expected error for unknown alias")
	}

	PublicKeyPath := filepath.Join(env.SSHPath, PublicKey)
	if _, err := os.Stat(PublicKeyPath); !os.IsNotExist(err) {
		t.Error("should create symbol link for public key")
	}

	PrivateKeyPath := filepath.Join(env.SSHPath, PrivateKey)
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

func TestIsEmpty(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)
	if ok, err := IsEmpty(env.StorePath); err != nil || !ok {
		t.Error("directory should be empty")
	}
}

func TestIsEmpty_NonEmpty(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)
	mustWriteFile(t, filepath.Join(env.StorePath, "marker"), "x")
	ok, err := IsEmpty(env.StorePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("directory containing a file should not be reported as empty")
	}
}

func TestIsEmpty_NonExistent(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	if _, err := IsEmpty(missing); err == nil {
		t.Error("expected error for non-existent path")
	}
}

func TestGetBakFileName_Format(t *testing.T) {
	name := GetBakFileName()
	if !strings.HasPrefix(name, "skm-") || !strings.HasSuffix(name, ".tar.gz") {
		t.Errorf("unexpected backup filename: %q", name)
	}
	// skm-YYYYMMDDhhmmss.tar.gz is 26 chars.
	if len(name) != len("skm-")+14+len(".tar.gz") {
		t.Errorf("backup filename has unexpected length: %q", name)
	}
}

func TestLoadSSHKeys_DetectsED25519(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	dir := filepath.Join(env.StorePath, "ed-alias")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}
	mustWriteFile(t, filepath.Join(dir, "id_ed25519"), "priv")
	mustWriteFile(t, filepath.Join(dir, "id_ed25519.pub"), "pub")

	keys := LoadSSHKeys(env)
	k, ok := keys["ed-alias"]
	if !ok {
		t.Fatal("expected ed-alias to be loaded")
	}
	if k.Type == nil || k.Type.Name != "ed25519" {
		t.Errorf("expected type ed25519, got %#v", k.Type)
	}
	if k.IsDefault {
		t.Error("key without an active ~/.ssh symlink should not be marked default")
	}
}

func TestLoadSSHKeys_DetectsDefault(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	alias := "active"
	dir := filepath.Join(env.StorePath, alias)
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}
	priv := filepath.Join(dir, "id_rsa")
	pub := filepath.Join(dir, "id_rsa.pub")
	mustWriteFile(t, priv, "priv")
	mustWriteFile(t, pub, "pub")
	// Symlink ~/.ssh/id_rsa to the stored key so IsDefault should be set.
	if err := os.Symlink(priv, filepath.Join(env.SSHPath, "id_rsa")); err != nil {
		t.Fatalf("setup symlink: %v", err)
	}

	keys := LoadSSHKeys(env)
	k, ok := keys[alias]
	if !ok {
		t.Fatal("expected alias to be loaded")
	}
	if !k.IsDefault {
		t.Error("expected IsDefault=true when ~/.ssh symlink points to the stored key")
	}
}

func TestCreateLink_CreatesSymlinks(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	alias := "linked"
	dir := filepath.Join(env.StorePath, alias)
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}
	priv := filepath.Join(dir, "id_rsa")
	pub := filepath.Join(dir, "id_rsa.pub")
	mustWriteFile(t, priv, "priv")
	mustWriteFile(t, pub, "pub")

	keys := LoadSSHKeys(env)
	if err := CreateLink(alias, keys, env); err != nil {
		t.Fatalf("CreateLink returned error: %v", err)
	}

	for filename, target := range map[string]string{
		"id_rsa":     priv,
		"id_rsa.pub": pub,
	} {
		linkPath := filepath.Join(env.SSHPath, filename)
		fi, err := os.Lstat(linkPath)
		if err != nil {
			t.Fatalf("expected symlink at %s: %v", linkPath, err)
		}
		if fi.Mode()&os.ModeSymlink == 0 {
			t.Errorf("%s is not a symlink", filename)
			continue
		}
		if got := ParsePath(linkPath); got != target {
			t.Errorf("symlink %s -> %q, want %q", filename, got, target)
		}
	}
}

func TestCreateLink_UnknownAliasClearsExisting(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	// Drop a stand-in key into ~/.ssh that should be wiped by ClearKey.
	mustWriteFile(t, filepath.Join(env.SSHPath, PrivateKey), "existing")

	if err := CreateLink("does-not-exist", map[string]*models.SSHKey{}, env); err == nil {
		t.Error("expected error for unknown alias")
	}

	if _, err := os.Stat(filepath.Join(env.SSHPath, PrivateKey)); !os.IsNotExist(err) {
		t.Error("ClearKey should run even when the requested alias is missing")
	}
}

func TestDeleteKey_InUseClearsSshSymlinks(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	alias := "to-delete"
	dir := filepath.Join(env.StorePath, alias)
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}
	priv := filepath.Join(dir, "id_rsa")
	pub := filepath.Join(dir, "id_rsa.pub")
	mustWriteFile(t, priv, "priv")
	mustWriteFile(t, pub, "pub")
	if err := os.Symlink(priv, filepath.Join(env.SSHPath, "id_rsa")); err != nil {
		t.Fatalf("setup priv symlink: %v", err)
	}
	if err := os.Symlink(pub, filepath.Join(env.SSHPath, "id_rsa.pub")); err != nil {
		t.Fatalf("setup pub symlink: %v", err)
	}

	keys := LoadSSHKeys(env)
	key, ok := keys[alias]
	if !ok {
		t.Fatal("setup: alias not loaded")
	}
	DeleteKey(alias, key, env, DeleteOptions{testMode: true})

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("alias directory should be removed")
	}
	for _, base := range []string{"id_rsa", "id_rsa.pub"} {
		if _, err := os.Lstat(filepath.Join(env.SSHPath, base)); !os.IsNotExist(err) {
			t.Errorf("expected ~/.ssh/%s to be cleared", base)
		}
	}
}

// writeHook drops an executable hook script at path that, when run, writes
// its $SKM_EVENT/$SKM_ALIAS/$1 to logFile so the test can verify it fired.
// extraExit is appended verbatim to the script body, letting individual tests
// force a non-zero exit.
func writeHook(t *testing.T, path, logFile, extraExit string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatalf("mkdir hook dir: %v", err)
	}
	body := "#!/bin/sh\n" +
		"printf '%s %s %s\\n' \"$SKM_EVENT\" \"$SKM_ALIAS\" \"$1\" >> " + logFile + "\n" +
		extraExit
	if err := os.WriteFile(path, []byte(body), 0755); err != nil {
		t.Fatalf("write hook: %v", err)
	}
}

func readLog(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return ""
	}
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	return string(b)
}

func TestRunHook_LegacyPostUseStillRuns(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	alias := "legacy"
	if err := os.MkdirAll(filepath.Join(env.StorePath, alias), 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}
	log := filepath.Join(env.StorePath, "fired.log")
	writeHook(t, filepath.Join(env.StorePath, alias, HookName), log, "")

	if err := RunHook(EventPostUse, alias, env); err != nil {
		t.Fatalf("RunHook returned error: %v", err)
	}
	got := readLog(t, log)
	want := "post-use legacy legacy\n"
	if got != want {
		t.Errorf("legacy hook output: got %q want %q", got, want)
	}
}

func TestRunHook_PerKeyEventFires(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	alias := "k1"
	log := filepath.Join(env.StorePath, "fired.log")
	writeHook(t, filepath.Join(env.StorePath, alias, HooksDir, EventPostCreate), log, "")

	if err := RunHook(EventPostCreate, alias, env); err != nil {
		t.Fatalf("RunHook: %v", err)
	}
	if got := readLog(t, log); got != "post-create k1 k1\n" {
		t.Errorf("unexpected log: %q", got)
	}
}

func TestRunHook_GlobalFiresWhenPerKeyAbsent(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	log := filepath.Join(env.StorePath, "fired.log")
	writeHook(t, filepath.Join(env.StorePath, HooksDir, EventPostUse), log, "")

	if err := RunHook(EventPostUse, "anything", env); err != nil {
		t.Fatalf("RunHook: %v", err)
	}
	if got := readLog(t, log); got != "post-use anything anything\n" {
		t.Errorf("unexpected log: %q", got)
	}
}

func TestRunHook_GlobalAndPerKeyBothFire(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	alias := "both"
	if err := os.MkdirAll(filepath.Join(env.StorePath, alias), 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}
	log := filepath.Join(env.StorePath, "fired.log")
	// Global hook tags its line "G", per-key tags "K", so the ordering is
	// observable in the log.
	writeGlobal := "#!/bin/sh\necho G $SKM_EVENT $SKM_ALIAS >> " + log + "\n"
	writePerKey := "#!/bin/sh\necho K $SKM_EVENT $SKM_ALIAS >> " + log + "\n"
	if err := os.MkdirAll(filepath.Join(env.StorePath, HooksDir), 0700); err != nil {
		t.Fatalf("mkdir global: %v", err)
	}
	if err := os.WriteFile(filepath.Join(env.StorePath, HooksDir, EventPostUse), []byte(writeGlobal), 0755); err != nil {
		t.Fatalf("write global: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(env.StorePath, alias, HooksDir), 0700); err != nil {
		t.Fatalf("mkdir perkey: %v", err)
	}
	if err := os.WriteFile(filepath.Join(env.StorePath, alias, HooksDir, EventPostUse), []byte(writePerKey), 0755); err != nil {
		t.Fatalf("write perkey: %v", err)
	}

	if err := RunHook(EventPostUse, alias, env); err != nil {
		t.Fatalf("RunHook: %v", err)
	}
	got := readLog(t, log)
	want := "G post-use both\nK post-use both\n"
	if got != want {
		t.Errorf("ordering wrong:\n got: %q\nwant: %q", got, want)
	}
}

func TestRunHook_NonExecutableSkippedNoError(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	alias := "noexec"
	dir := filepath.Join(env.StorePath, alias, HooksDir)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, EventPostUse), []byte("#!/bin/sh\necho hi\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := RunHook(EventPostUse, alias, env); err != nil {
		t.Errorf("non-executable hook should be silently skipped, got: %v", err)
	}
}

func TestRunHook_PreEventAbortsOnNonZeroExit(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	alias := "veto"
	if err := os.MkdirAll(filepath.Join(env.StorePath, alias), 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}
	log := filepath.Join(env.StorePath, "fired.log")
	writeHook(t, filepath.Join(env.StorePath, alias, HooksDir, EventPreDelete), log, "exit 7\n")

	err := RunHook(EventPreDelete, alias, env)
	if err == nil {
		t.Fatal("expected error from failing pre-delete hook")
	}
	if !strings.Contains(err.Error(), "pre-delete") {
		t.Errorf("error should mention event: %v", err)
	}
}

func TestRunHook_PostEventSwallowsFailure(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	alias := "boom"
	if err := os.MkdirAll(filepath.Join(env.StorePath, alias), 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}
	log := filepath.Join(env.StorePath, "fired.log")
	writeHook(t, filepath.Join(env.StorePath, alias, HooksDir, EventPostUse), log, "exit 3\n")

	if err := RunHook(EventPostUse, alias, env); err != nil {
		t.Errorf("post-* hook failures must not propagate: %v", err)
	}
}

func TestRunHook_ExtraEnvIsPassed(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	alias := "ctx"
	if err := os.MkdirAll(filepath.Join(env.StorePath, alias), 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}
	log := filepath.Join(env.StorePath, "fired.log")
	body := "#!/bin/sh\necho \"$SKM_REMOTE_HOST $SKM_REMOTE_PORT\" >> " + log + "\n"
	dir := filepath.Join(env.StorePath, alias, HooksDir)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, EventPostCopy), []byte(body), 0755); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := RunHook(EventPostCopy, alias, env, "SKM_REMOTE_HOST", "h.example", "SKM_REMOTE_PORT", "2222"); err != nil {
		t.Fatalf("RunHook: %v", err)
	}
	if got := readLog(t, log); got != "h.example 2222\n" {
		t.Errorf("extra env not propagated: %q", got)
	}
}

func TestDeleteKey_PreDeleteHookCanAbort(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	alias := "guarded"
	dir := filepath.Join(env.StorePath, alias)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}
	mustWriteFile(t, filepath.Join(dir, "id_rsa"), "priv")
	mustWriteFile(t, filepath.Join(dir, "id_rsa.pub"), "pub")
	log := filepath.Join(env.StorePath, "fired.log")
	writeHook(t, filepath.Join(dir, HooksDir, EventPreDelete), log, "exit 1\n")

	key := &models.SSHKey{
		PrivateKey: filepath.Join(dir, "id_rsa"),
		PublicKey:  filepath.Join(dir, "id_rsa.pub"),
	}
	DeleteKey(alias, key, env, DeleteOptions{testMode: true})

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("pre-delete hook returned non-zero; alias directory must survive")
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
