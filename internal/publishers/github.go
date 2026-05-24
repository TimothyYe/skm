package publishers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type gitHubPublisher struct {
	baseURL string
}

func (g *gitHubPublisher) Name() string { return "github" }

func (g *gitHubPublisher) TokenHint() string {
	return "set $GITHUB_TOKEN, pass --token, or run `gh auth login` (scope: admin:public_key)"
}

type gitHubKey struct {
	ID    int    `json:"id"`
	Key   string `json:"key"`
	Title string `json:"title"`
}

func (g *gitHubPublisher) Existing(ctx context.Context, token, publicKey string) (string, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.baseURL+"/user/keys", nil)
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

	var keys []gitHubKey
	if err := json.Unmarshal(body, &keys); err != nil {
		return "", false, fmt.Errorf("decode github keys: %w", err)
	}
	target := CanonicalKey(publicKey)
	for _, k := range keys {
		if CanonicalKey(k.Key) == target {
			return k.Title, true, nil
		}
	}
	return "", false, nil
}

func (g *gitHubPublisher) Publish(ctx context.Context, token, title, publicKey string) error {
	payload, _ := json.Marshal(map[string]string{
		"title": title,
		"key":   publicKey,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/user/keys", bytes.NewReader(payload))
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

func (g *gitHubPublisher) applyAuth(req *http.Request, token string) {
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
}
