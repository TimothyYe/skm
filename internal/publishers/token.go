package publishers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ResolveToken picks a token using the documented priority order:
//
//  1. explicit (the --token flag)
//  2. each of envVars (first non-empty wins; SKM_*_TOKEN overrides take priority
//     so users can scope a token to skm without disturbing other tooling)
//  3. cliFallback: if non-empty, shell out to that command and use stdout
//     (e.g. `gh auth token`, `glab auth token`). Errors are swallowed so a
//     missing binary just falls through to the no-token error.
//
// Returns the trimmed token or an empty string if none was found.
func ResolveToken(explicit string, envVars []string, cliFallback []string) string {
	if explicit != "" {
		return explicit
	}
	for _, name := range envVars {
		if v := strings.TrimSpace(os.Getenv(name)); v != "" {
			return v
		}
	}
	if len(cliFallback) > 0 {
		if _, err := exec.LookPath(cliFallback[0]); err == nil {
			out, err := exec.Command(cliFallback[0], cliFallback[1:]...).Output()
			if err == nil {
				if v := strings.TrimSpace(string(out)); v != "" {
					return v
				}
			}
		}
	}
	return ""
}

// NoTokenError is the message returned when token resolution fails. The
// publisher's TokenHint gets folded in so the user knows what scopes / CLI
// commands work for that specific provider.
func NoTokenError(provider, hint string) error {
	return fmt.Errorf("no token available for %s — %s", provider, hint)
}
