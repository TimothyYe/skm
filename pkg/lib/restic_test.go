package lib

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/TimothyYe/skm/internal/models"
)

func newEnv(t *testing.T) *models.Environment {
	t.Helper()
	dir := t.TempDir()
	store := filepath.Join(dir, ".skm")
	if err := os.MkdirAll(store, 0700); err != nil {
		t.Fatalf("mkdir store: %v", err)
	}
	return &models.Environment{StorePath: store, SSHPath: filepath.Join(dir, ".ssh")}
}

func TestLoadResticSettings_MissingFile(t *testing.T) {
	env := newEnv(t)
	if _, err := LoadResticSettings(env); err == nil {
		t.Error("expected error when restic.json is missing")
	}
}

func TestLoadResticSettings_RoundTrip(t *testing.T) {
	env := newEnv(t)
	cfg := ResticConfig{Repository: "/tmp/repo", PasswordFile: "/tmp/pw"}
	data, _ := json.Marshal(cfg)
	if err := os.WriteFile(filepath.Join(env.StorePath, "restic.json"), data, 0600); err != nil {
		t.Fatalf("seed: %v", err)
	}
	got, err := LoadResticSettings(env)
	if err != nil {
		t.Fatalf("LoadResticSettings: %v", err)
	}
	if got.Repository != cfg.Repository || got.PasswordFile != cfg.PasswordFile {
		t.Errorf("round-trip mismatch: got %+v want %+v", got, cfg)
	}
}

func TestLoadResticSettings_RejectsIncomplete(t *testing.T) {
	env := newEnv(t)
	// Missing password_file field — caller should not silently get an empty path.
	if err := os.WriteFile(filepath.Join(env.StorePath, "restic.json"), []byte(`{"repository":"/tmp/repo"}`), 0600); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if _, err := LoadResticSettings(env); err == nil {
		t.Error("expected error for incomplete config")
	}
}

func TestRequireInitializedResticRepo_LocalMissingPasswordFile(t *testing.T) {
	dir := t.TempDir()
	cfg := &ResticConfig{Repository: filepath.Join(dir, "repo"), PasswordFile: filepath.Join(dir, "pw")}
	if err := RequireInitializedResticRepo(cfg); err == nil {
		t.Error("expected error when password file is missing")
	}
}

func TestRequireInitializedResticRepo_LocalUninitializedRepo(t *testing.T) {
	dir := t.TempDir()
	pw := filepath.Join(dir, "pw")
	if err := os.WriteFile(pw, []byte("x"), 0600); err != nil {
		t.Fatalf("seed pw: %v", err)
	}
	cfg := &ResticConfig{Repository: filepath.Join(dir, "repo"), PasswordFile: pw}
	if err := RequireInitializedResticRepo(cfg); err == nil {
		t.Error("expected error when local repository has no config file")
	}
}

func TestRequireInitializedResticRepo_RemoteRepoSkipsLocalChecks(t *testing.T) {
	dir := t.TempDir()
	pw := filepath.Join(dir, "pw")
	if err := os.WriteFile(pw, []byte("x"), 0600); err != nil {
		t.Fatalf("seed pw: %v", err)
	}
	// s3:/sftp:/b2: repositories cannot be statted locally; the helper must
	// not fail for them — let restic surface backend errors at invocation.
	for _, repo := range []string{
		"s3:s3.amazonaws.com/bucket/skm",
		"s3:https://abc.r2.cloudflarestorage.com/bucket",
		"sftp:user@host:/data/skm",
		"b2:bucket/skm",
		"azure:container/skm",
		"gs:bucket/skm",
		"rclone:remote:path",
	} {
		cfg := &ResticConfig{Repository: repo, PasswordFile: pw}
		if err := RequireInitializedResticRepo(cfg); err != nil {
			t.Errorf("repo %q should pass local checks, got: %v", repo, err)
		}
	}
}

func TestIsLocalRepo(t *testing.T) {
	cases := map[string]bool{
		"/Users/me/.skm-backups": true,
		"./relative":             true,
		"~/literal-tilde":        true,
		"s3:bucket/path":         false,
		"sftp:user@host:/path":   false,
		"b2:bucket":              false,
		"azure:c/p":              false,
		"gs:b/p":                 false,
		"rclone:r:p":             false,
		"rest:https://example":   false,
	}
	for input, want := range cases {
		if got := isLocalRepo(input); got != want {
			t.Errorf("isLocalRepo(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestInitResticRepository_RefusesIfConfigExists(t *testing.T) {
	env := newEnv(t)
	if err := os.WriteFile(filepath.Join(env.StorePath, "restic.json"), []byte("{}"), 0600); err != nil {
		t.Fatalf("seed: %v", err)
	}
	// InitResticRepository should bail out before any prompt — we'd hang on
	// stdin otherwise. If this test hangs, the early-return guard is broken.
	if err := InitResticRepository(env); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
