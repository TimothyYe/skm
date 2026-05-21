package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/TimothyYe/skm/internal/models"
)

// TrashEntry describes a single key directory currently in the trash.
type TrashEntry struct {
	// Name is the on-disk directory name under .trash (e.g. "prod-20260521150405").
	Name string
	// Alias is the original alias the entry was deleted under.
	Alias string
	// DeletedAt is the timestamp parsed from the directory name.
	DeletedAt time.Time
	// Path is the absolute path to the trashed directory.
	Path string
}

// MoveToTrash relocates the alias directory into the store's .trash, suffixed
// with a timestamp so repeated deletions of the same alias coexist. Returns
// the on-disk name of the trashed entry.
func MoveToTrash(alias string, env *models.Environment) (string, error) {
	src := filepath.Join(env.StorePath, alias)
	if _, err := os.Stat(src); err != nil {
		return "", err
	}

	trashRoot := filepath.Join(env.StorePath, TrashDir)
	if err := os.MkdirAll(trashRoot, 0700); err != nil {
		return "", fmt.Errorf("create trash dir: %w", err)
	}

	name := fmt.Sprintf("%s-%s", alias, time.Now().Format(trashTimestampLayout))
	dst := filepath.Join(trashRoot, name)

	// On the off-chance two deletes happen in the same second, walk the suffix.
	for i := 2; ; i++ {
		if _, err := os.Stat(dst); os.IsNotExist(err) {
			break
		}
		dst = filepath.Join(trashRoot, fmt.Sprintf("%s-%d", name, i))
	}

	if err := os.Rename(src, dst); err != nil {
		return "", fmt.Errorf("move to trash: %w", err)
	}
	return filepath.Base(dst), nil
}

// ListTrash returns trashed entries, most recently deleted first.
func ListTrash(env *models.Environment) ([]TrashEntry, error) {
	trashRoot := filepath.Join(env.StorePath, TrashDir)
	entries, err := os.ReadDir(trashRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	out := make([]TrashEntry, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		alias, deletedAt := parseTrashName(name)
		out = append(out, TrashEntry{
			Name:      name,
			Alias:     alias,
			DeletedAt: deletedAt,
			Path:      filepath.Join(trashRoot, name),
		})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].DeletedAt.After(out[j].DeletedAt) })
	return out, nil
}

// RestoreFromTrash moves a trashed entry back into the live store under
// asAlias (defaults to the entry's original alias when empty). The name may
// be either the exact trash directory name (e.g. "prod-20260521150412") or
// the original alias — the latter only resolves when exactly one trashed
// entry matches. Returns the alias the key was restored under.
func RestoreFromTrash(name, asAlias string, env *models.Environment) (string, error) {
	src := filepath.Join(env.StorePath, TrashDir, name)
	info, err := os.Stat(src)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
		// Fall back to alias lookup.
		resolved, alErr := resolveTrashByAlias(name, env)
		if alErr != nil {
			return "", alErr
		}
		name = resolved
		src = filepath.Join(env.StorePath, TrashDir, name)
		info, err = os.Stat(src)
		if err != nil {
			return "", err
		}
	}
	if !info.IsDir() {
		return "", fmt.Errorf("not a trash entry: %s", name)
	}

	if asAlias == "" {
		asAlias, _ = parseTrashName(name)
	}
	if asAlias == "" {
		return "", fmt.Errorf("cannot infer alias from %q; pass --as", name)
	}

	dst := filepath.Join(env.StorePath, asAlias)
	if _, err := os.Stat(dst); err == nil {
		return "", fmt.Errorf("alias %q already exists; pass --as to choose a different name", asAlias)
	}

	if err := os.Rename(src, dst); err != nil {
		return "", fmt.Errorf("restore from trash: %w", err)
	}
	return asAlias, nil
}

// EmptyTrash deletes all entries in the trash directory and returns the count
// of entries removed.
func EmptyTrash(env *models.Environment) (int, error) {
	entries, err := ListTrash(env)
	if err != nil {
		return 0, err
	}
	removed := 0
	for _, e := range entries {
		if err := os.RemoveAll(e.Path); err != nil {
			return removed, fmt.Errorf("remove %s: %w", e.Name, err)
		}
		removed++
	}
	return removed, nil
}

// resolveTrashByAlias looks up trashed entries whose original alias matches.
// Returns the on-disk name when exactly one entry matches; errors otherwise.
func resolveTrashByAlias(alias string, env *models.Environment) (string, error) {
	entries, err := ListTrash(env)
	if err != nil {
		return "", err
	}
	matches := make([]TrashEntry, 0, 2)
	for _, e := range entries {
		if e.Alias == alias {
			matches = append(matches, e)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no trashed key matches %q", alias)
	case 1:
		return matches[0].Name, nil
	default:
		names := make([]string, 0, len(matches))
		for _, m := range matches {
			names = append(names, m.Name)
		}
		return "", fmt.Errorf("multiple trashed entries match %q; specify one of: %s", alias, strings.Join(names, ", "))
	}
}

// parseTrashName splits "<alias>-YYYYMMDDhhmmss" (with optional "-N" disambiguation
// suffix) back into the alias and timestamp. Falls back to zero time when the
// suffix is unparseable.
func parseTrashName(name string) (string, time.Time) {
	// Find the trailing -<14 digits> (optionally followed by -N).
	parts := strings.Split(name, "-")
	for i := len(parts) - 1; i >= 1; i-- {
		candidate := parts[i]
		if len(candidate) == len(trashTimestampLayout) {
			if ts, err := time.ParseInLocation(trashTimestampLayout, candidate, time.Local); err == nil {
				return strings.Join(parts[:i], "-"), ts
			}
		}
	}
	return name, time.Time{}
}
