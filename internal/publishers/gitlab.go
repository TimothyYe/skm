package publishers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type gitLabPublisher struct {
	baseURL string
}

func (g *gitLabPublisher) Name() string { return "gitlab" }

func (g *gitLabPublisher) TokenHint() string {
	return "set $GITLAB_TOKEN, pass --token, or run `glab auth login` (scope: api)"
}

type gitLabKey struct {
	ID    int    `json:"id"`
	Key   string `json:"key"`
	Title string `json:"title"`
}

func (g *gitLabPublisher) Existing(ctx context.Context, token, publicKey string) (string, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.baseURL+"/api/v4/user/keys", nil)
	if err != nil {
		return "", false, err
	}
	g.applyAuth(req, token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return "", false, decodeError(resp.StatusCode, body)
	}

	var keys []gitLabKey
	if err := json.Unmarshal(body, &keys); err != nil {
		return "", false, fmt.Errorf("decode gitlab keys: %w", err)
	}
	target := CanonicalKey(publicKey)
	for _, k := range keys {
		if CanonicalKey(k.Key) == target {
			return k.Title, true, nil
		}
	}
	return "", false, nil
}

func (g *gitLabPublisher) Publish(ctx context.Context, token, title, publicKey string) error {
	payload, _ := json.Marshal(map[string]string{
		"title": title,
		"key":   publicKey,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/api/v4/user/keys", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	g.applyAuth(req, token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return decodeError(resp.StatusCode, body)
	}
	return nil
}

func (g *gitLabPublisher) applyAuth(req *http.Request, token string) {
	// GitLab accepts both PRIVATE-TOKEN (PAT) and Authorization: Bearer (OAuth).
	// PRIVATE-TOKEN works for both; using it keeps the call simple.
	req.Header.Set("PRIVATE-TOKEN", token)
	req.Header.Set("Accept", "application/json")
}
