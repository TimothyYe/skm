package publishers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type bitbucketPublisher struct {
	baseURL string
	user    string // workspace / account that owns the key list
}

func (b *bitbucketPublisher) Name() string { return "bitbucket" }

func (b *bitbucketPublisher) TokenHint() string {
	return "set $BITBUCKET_TOKEN or pass --token (Bitbucket app password, scope: Account Write)"
}

type bitbucketKey struct {
	UUID  string `json:"uuid"`
	Key   string `json:"key"`
	Label string `json:"label"`
}

type bitbucketKeyList struct {
	Values []bitbucketKey `json:"values"`
	Next   string         `json:"next"`
}

func (b *bitbucketPublisher) Existing(ctx context.Context, token, publicKey string) (string, bool, error) {
	target := CanonicalKey(publicKey)
	url := fmt.Sprintf("%s/2.0/users/%s/ssh-keys", b.baseURL, b.user)
	for url != "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return "", false, err
		}
		b.applyAuth(req, token)
		resp, err := httpClient.Do(req)
		if err != nil {
			return "", false, err
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode/100 != 2 {
			return "", false, decodeError(resp.StatusCode, body)
		}
		var page bitbucketKeyList
		if err := json.Unmarshal(body, &page); err != nil {
			return "", false, fmt.Errorf("decode bitbucket keys: %w", err)
		}
		for _, k := range page.Values {
			if CanonicalKey(k.Key) == target {
				return k.Label, true, nil
			}
		}
		url = page.Next
	}
	return "", false, nil
}

func (b *bitbucketPublisher) Publish(ctx context.Context, token, title, publicKey string) error {
	payload, _ := json.Marshal(map[string]string{
		"label": title,
		"key":   publicKey,
	})
	url := fmt.Sprintf("%s/2.0/users/%s/ssh-keys", b.baseURL, b.user)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	b.applyAuth(req, token)
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

func (b *bitbucketPublisher) applyAuth(req *http.Request, token string) {
	// Bitbucket app passwords use HTTP Basic with the workspace user; we reuse
	// b.user as both the URL segment and the basic-auth username, which matches
	// the common app-password setup.
	req.SetBasicAuth(b.user, token)
	req.Header.Set("Accept", "application/json")
}
