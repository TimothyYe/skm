package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMoveToTrash_RoundTrip(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	alias := "to-trash"
	dir := filepath.Join(env.StorePath, alias)
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "id_rsa"), []byte("priv"), 0600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	name, err := MoveToTrash(alias, env)
	if err != nil {
		t.Fatalf("MoveToTrash: %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("original alias dir should be gone after move")
	}
	if _, err := os.Stat(filepath.Join(env.StorePath, TrashDir, name)); err != nil {
		t.Errorf("trashed dir should exist: %v", err)
	}

	entries, err := ListTrash(env)
	if err != nil {
		t.Fatalf("ListTrash: %v", err)
	}
	if len(entries) != 1 || entries[0].Alias != alias {
		t.Fatalf("ListTrash returned %+v", entries)
	}

	restored, err := RestoreFromTrash(name, "", env)
	if err != nil {
		t.Fatalf("RestoreFromTrash: %v", err)
	}
	if restored != alias {
		t.Errorf("restored alias = %q, want %q", restored, alias)
	}
	if _, err := os.Stat(filepath.Join(env.StorePath, alias, "id_rsa")); err != nil {
		t.Errorf("restored key should exist: %v", err)
	}
}

func TestRestoreFromTrash_AliasCollision(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	alias := "dup"
	dir := filepath.Join(env.StorePath, alias)
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}
	name, err := MoveToTrash(alias, env)
	if err != nil {
		t.Fatalf("MoveToTrash: %v", err)
	}

	// Re-create a live alias with the same name.
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatalf("setup2: %v", err)
	}

	if _, err := RestoreFromTrash(name, "", env); err == nil {
		t.Error("expected collision error when restoring over existing alias")
	}

	// --as <new> path should succeed.
	if _, err := RestoreFromTrash(name, "dup-restored", env); err != nil {
		t.Fatalf("RestoreFromTrash with --as: %v", err)
	}
	if _, err := os.Stat(filepath.Join(env.StorePath, "dup-restored")); err != nil {
		t.Errorf("restored alias should exist: %v", err)
	}
}

func TestRestoreFromTrash_ByAliasShortName(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	alias := "shortcut"
	dir := filepath.Join(env.StorePath, alias)
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if _, err := MoveToTrash(alias, env); err != nil {
		t.Fatalf("MoveToTrash: %v", err)
	}

	// Passing just the alias (not the timestamped name) should resolve to the lone entry.
	restored, err := RestoreFromTrash(alias, "", env)
	if err != nil {
		t.Fatalf("RestoreFromTrash by alias: %v", err)
	}
	if restored != alias {
		t.Errorf("restored alias = %q, want %q", restored, alias)
	}
}

func TestRestoreFromTrash_AmbiguousAliasErrors(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	alias := "twice"
	dir := filepath.Join(env.StorePath, alias)
	for i := 0; i < 2; i++ {
		if err := os.Mkdir(dir, 0700); err != nil {
			t.Fatalf("setup %d: %v", i, err)
		}
		if _, err := MoveToTrash(alias, env); err != nil {
			t.Fatalf("MoveToTrash %d: %v", i, err)
		}
	}

	if _, err := RestoreFromTrash(alias, "", env); err == nil {
		t.Error("expected ambiguous-match error when alias has multiple trashed entries")
	}
}

func TestEmptyTrash(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	for _, alias := range []string{"a", "b", "c"} {
		dir := filepath.Join(env.StorePath, alias)
		if err := os.Mkdir(dir, 0700); err != nil {
			t.Fatalf("setup: %v", err)
		}
		if _, err := MoveToTrash(alias, env); err != nil {
			t.Fatalf("MoveToTrash: %v", err)
		}
	}

	removed, err := EmptyTrash(env)
	if err != nil {
		t.Fatalf("EmptyTrash: %v", err)
	}
	if removed != 3 {
		t.Errorf("removed = %d, want 3", removed)
	}

	entries, err := ListTrash(env)
	if err != nil {
		t.Fatalf("ListTrash: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("trash should be empty, got %d entries", len(entries))
	}
}

func TestLoadSSHKeys_IgnoresTrash(t *testing.T) {
	env := setupTestEnvironment(t)
	defer tearDownTestEnvironment(t, env)

	alias := "live"
	dir := filepath.Join(env.StorePath, alias)
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "id_rsa"), []byte("priv"), 0600); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "id_rsa.pub"), []byte("pub"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Trash the alias, then re-create a fresh "live" — the trashed copy must not
	// reappear in LoadSSHKeys.
	if _, err := MoveToTrash(alias, env); err != nil {
		t.Fatalf("MoveToTrash: %v", err)
	}
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatalf("setup2: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "id_rsa"), []byte("priv2"), 0600); err != nil {
		t.Fatalf("setup2: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "id_rsa.pub"), []byte("pub2"), 0644); err != nil {
		t.Fatalf("setup2: %v", err)
	}

	keys := LoadSSHKeys(env)
	if _, ok := keys[alias]; !ok {
		t.Error("live alias should be loaded")
	}
	// Verify no spurious aliases (no entry should come from the trash dir).
	if len(keys) != 1 {
		t.Errorf("LoadSSHKeys returned %d keys; want 1: %+v", len(keys), keys)
	}
}
