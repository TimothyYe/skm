package actions

import (
	"os"
	"path/filepath"
	"strings"
)

const skmrcFileName = ".skmrc"

// findSkmrcAlias walks from dir up to the filesystem root and returns the key
// alias written in the nearest .skmrc: its first line that is neither blank
// nor a "#" comment.
func findSkmrcAlias(dir string) (string, bool) {
	for {
		if alias, ok := readSkmrc(filepath.Join(dir, skmrcFileName)); ok {
			return alias, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func readSkmrc(path string) (string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			return line, true
		}
	}
	return "", false
}
