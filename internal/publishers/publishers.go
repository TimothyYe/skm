// Package publishers uploads SSH public keys to Git hosting providers
// (GitHub, GitLab, Bitbucket) via their HTTP APIs.
//
// Each Publisher implementation handles one provider. Callers go through the
// Resolve constructor instead of instantiating types directly so the action
// layer stays agnostic about which providers exist.
package publishers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Publisher uploads a single SSH public key to a provider and reports whether
// the same key is already published (compared by canonical type+base64 form).
type Publisher interface {
	// Name returns the provider's short identifier ("github", "gitlab", "bitbucket").
	Name() string
	// Existing returns the title of an already-uploaded key matching publicKey,
	// or ("", false) if not present. publicKey is the raw .pub file contents.
	Existing(ctx context.Context, token, publicKey string) (title string, found bool, err error)
	// Publish uploads publicKey under the given title.
	Publish(ctx context.Context, token, title, publicKey string) error
	// TokenHint describes how to obtain a token for this provider; surfaced in
	// the "no token" error message so users know what to do.
	TokenHint() string
}

// Resolve returns a Publisher for the named provider. baseURL overrides the
// default endpoint (used for GitHub Enterprise / self-hosted GitLab). For
// Bitbucket, extra["user"] supplies the workspace/account name embedded in
// the API URL.
func Resolve(name, baseURL string, extra map[string]string) (Publisher, error) {
	switch strings.ToLower(name) {
	case "github":
		return &gitHubPublisher{baseURL: firstNonEmpty(baseURL, "https://api.github.com")}, nil
	case "gitlab":
		return &gitLabPublisher{baseURL: firstNonEmpty(baseURL, "https://gitlab.com")}, nil
	case "bitbucket":
		user := extra["user"]
		if user == "" {
			return nil, fmt.Errorf("bitbucket requires --user <workspace>")
		}
		return &bitbucketPublisher{baseURL: firstNonEmpty(baseURL, "https://api.bitbucket.org"), user: user}, nil
	default:
		return nil, fmt.Errorf("unknown provider %q (expected github|gitlab|bitbucket)", name)
	}
}

// CanonicalKey strips the comment from a public key line so two keys differing
// only in their trailing comment compare equal. Providers vary in whether they
// preserve the comment — comparing type+base64 is the reliable form.
func CanonicalKey(pubKey string) string {
	line, _, _ := strings.Cut(pubKey, "\n")
	fields := strings.Fields(strings.TrimSpace(line))
	if len(fields) < 2 {
		return strings.TrimSpace(line)
	}
	return fields[0] + " " + fields[1]
}

// httpClient is the shared client used by all providers. The 30s timeout
// catches stuck connections without being short enough to break slow networks.
var httpClient = &http.Client{Timeout: 30 * time.Second}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// decodeError extracts a short message from a provider's error response body
// so we can surface something more useful than "HTTP 422".
func decodeError(status int, body []byte) error {
	msg := strings.TrimSpace(string(body))
	if len(msg) > 200 {
		msg = msg[:200] + "..."
	}
	if msg == "" {
		return fmt.Errorf("HTTP %d", status)
	}
	return fmt.Errorf("HTTP %d: %s", status, msg)
}
