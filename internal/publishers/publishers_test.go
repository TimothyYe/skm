package publishers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestCanonicalKey(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"with comment", "ssh-ed25519 AAAA timothy@laptop", "ssh-ed25519 AAAA"},
		{"no comment", "ssh-ed25519 AAAA", "ssh-ed25519 AAAA"},
		{"trailing newline", "ssh-ed25519 AAAA timothy@laptop\n", "ssh-ed25519 AAAA"},
		{"multi-word comment", "ssh-rsa BBBB me at host", "ssh-rsa BBBB"},
		{"only type", "ssh-rsa", "ssh-rsa"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanonicalKey(tt.in); got != tt.want {
				t.Errorf("CanonicalKey(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestResolveProvider(t *testing.T) {
	if _, err := Resolve("github", "", nil); err != nil {
		t.Errorf("github: %v", err)
	}
	if _, err := Resolve("gitlab", "https://gitlab.example.com", nil); err != nil {
		t.Errorf("gitlab self-hosted: %v", err)
	}
	if _, err := Resolve("bitbucket", "", nil); err == nil {
		t.Errorf("bitbucket without --user should error")
	}
	if _, err := Resolve("bitbucket", "", map[string]string{"user": "alice"}); err != nil {
		t.Errorf("bitbucket with user: %v", err)
	}
	if _, err := Resolve("sourcehut", "", nil); err == nil {
		t.Errorf("unknown provider should error")
	}
}

func TestResolveToken(t *testing.T) {
	t.Setenv("SKM_TEST_TOKEN_A", "")
	t.Setenv("SKM_TEST_TOKEN_B", "")

	// explicit wins
	if got := ResolveToken("explicit", []string{"SKM_TEST_TOKEN_A"}, nil); got != "explicit" {
		t.Errorf("expected explicit, got %q", got)
	}

	// env var picked when explicit empty
	t.Setenv("SKM_TEST_TOKEN_A", "env-a")
	if got := ResolveToken("", []string{"SKM_TEST_TOKEN_A"}, nil); got != "env-a" {
		t.Errorf("expected env-a, got %q", got)
	}

	// first non-empty env var wins
	t.Setenv("SKM_TEST_TOKEN_A", "")
	t.Setenv("SKM_TEST_TOKEN_B", "env-b")
	if got := ResolveToken("", []string{"SKM_TEST_TOKEN_A", "SKM_TEST_TOKEN_B"}, nil); got != "env-b" {
		t.Errorf("expected env-b, got %q", got)
	}

	// no token, no fallback → empty string
	t.Setenv("SKM_TEST_TOKEN_A", "")
	t.Setenv("SKM_TEST_TOKEN_B", "")
	if got := ResolveToken("", []string{"SKM_TEST_TOKEN_A"}, nil); got != "" {
		t.Errorf("expected empty, got %q", got)
	}

	// missing CLI binary falls through silently (echo is unlikely to be useful;
	// use a guaranteed-missing name)
	if got := ResolveToken("", nil, []string{"definitely-not-a-real-binary-xyz"}); got != "" {
		t.Errorf("missing CLI should yield empty token, got %q", got)
	}

	// stdout from a CLI fallback is used and trimmed
	if got := ResolveToken("", nil, []string{"printf", "  cli-tok  \n"}); got != "cli-tok" {
		t.Errorf("expected cli-tok, got %q", got)
	}
}

const testPubKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIabcdef timothy@laptop"

// fakeServer captures requests sent by the providers and returns canned
// responses. Used by all three provider tests.
type fakeServer struct {
	mu       sync.Mutex
	requests []*http.Request
	bodies   []string
	server   *httptest.Server
}

func newFakeServer(handler http.HandlerFunc) *fakeServer {
	fs := &fakeServer{}
	fs.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fs.mu.Lock()
		fs.requests = append(fs.requests, r)
		fs.bodies = append(fs.bodies, string(body))
		fs.mu.Unlock()
		// Re-attach body for the handler to read if needed.
		r.Body = io.NopCloser(strings.NewReader(string(body)))
		handler(w, r)
	}))
	return fs
}

func (fs *fakeServer) close() { fs.server.Close() }

func (fs *fakeServer) lastBody() string {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if len(fs.bodies) == 0 {
		return ""
	}
	return fs.bodies[len(fs.bodies)-1]
}

func TestGitHubPublish(t *testing.T) {
	fs := newFakeServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user/keys" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("auth header = %q", got)
		}
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`[]`))
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":1}`))
		}
	})
	defer fs.close()

	p := &gitHubPublisher{baseURL: fs.server.URL}
	ctx := context.Background()

	// Existing → not found
	if _, found, err := p.Existing(ctx, "test-token", testPubKey); err != nil || found {
		t.Fatalf("Existing: found=%v err=%v", found, err)
	}
	// Publish → success
	if err := p.Publish(ctx, "test-token", "my-title", testPubKey); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	var sent map[string]string
	if err := json.Unmarshal([]byte(fs.lastBody()), &sent); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if sent["title"] != "my-title" || sent["key"] != testPubKey {
		t.Errorf("body mismatch: %+v", sent)
	}
}

func TestGitHubExistingMatch(t *testing.T) {
	// Provider returns the same key with a different trailing comment; should
	// still be detected as a match by canonical form.
	fs := newFakeServer(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"id":7,"title":"old-title","key":"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIabcdef different-comment"}]`))
	})
	defer fs.close()

	p := &gitHubPublisher{baseURL: fs.server.URL}
	title, found, err := p.Existing(context.Background(), "tok", testPubKey)
	if err != nil {
		t.Fatalf("Existing: %v", err)
	}
	if !found || title != "old-title" {
		t.Errorf("expected found=true title=old-title, got found=%v title=%q", found, title)
	}
}

func TestGitHubErrorResponse(t *testing.T) {
	fs := newFakeServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Bad credentials"}`))
	})
	defer fs.close()

	p := &gitHubPublisher{baseURL: fs.server.URL}
	err := p.Publish(context.Background(), "bad", "t", testPubKey)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "401") || !strings.Contains(err.Error(), "Bad credentials") {
		t.Errorf("error should mention status + body: %v", err)
	}
}

func TestGitLabPublish(t *testing.T) {
	fs := newFakeServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v4/user/keys" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if got := r.Header.Get("PRIVATE-TOKEN"); got != "test-token" {
			t.Errorf("auth header = %q", got)
		}
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`[]`))
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":1}`))
		}
	})
	defer fs.close()

	p := &gitLabPublisher{baseURL: fs.server.URL}
	ctx := context.Background()
	if _, found, err := p.Existing(ctx, "test-token", testPubKey); err != nil || found {
		t.Fatalf("Existing: found=%v err=%v", found, err)
	}
	if err := p.Publish(ctx, "test-token", "my-title", testPubKey); err != nil {
		t.Fatalf("Publish: %v", err)
	}
}

func TestBitbucketPublish(t *testing.T) {
	fs := newFakeServer(func(w http.ResponseWriter, r *http.Request) {
		wantPath := "/2.0/users/alice/ssh-keys"
		if r.URL.Path != wantPath {
			t.Errorf("unexpected path %s, want %s", r.URL.Path, wantPath)
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != "alice" || pass != "app-pw" {
			t.Errorf("basic auth = (%q,%q,%v)", user, pass, ok)
		}
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"values":[],"next":""}`))
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"uuid":"x"}`))
		}
	})
	defer fs.close()

	p := &bitbucketPublisher{baseURL: fs.server.URL, user: "alice"}
	ctx := context.Background()
	if _, found, err := p.Existing(ctx, "app-pw", testPubKey); err != nil || found {
		t.Fatalf("Existing: found=%v err=%v", found, err)
	}
	if err := p.Publish(ctx, "app-pw", "my-label", testPubKey); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	// Body uses "label" (Bitbucket's field name), not "title".
	var sent map[string]string
	_ = json.Unmarshal([]byte(fs.lastBody()), &sent)
	if sent["label"] != "my-label" {
		t.Errorf("expected label=my-label, got %+v", sent)
	}
}

func TestMain(m *testing.M) {
	// Ensure no leaked env vars from the test host can influence ResolveToken
	// in subtests that don't override.
	for _, k := range []string{
		"GITHUB_TOKEN", "GH_TOKEN", "GITLAB_TOKEN", "GL_TOKEN", "BITBUCKET_TOKEN",
		"SKM_GITHUB_TOKEN", "SKM_GITLAB_TOKEN", "SKM_BITBUCKET_TOKEN",
	} {
		_ = os.Unsetenv(k)
	}
	os.Exit(m.Run())
}
