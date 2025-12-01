package kra

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type authProvider struct {
	config *Config
	client *http.Client

	token     string
	expiresAt time.Time
	mu        sync.RWMutex
}

func newAuthProvider(config *Config) *authProvider {
	return &authProvider{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

func (a *authProvider) Token(ctx context.Context) (string, error) {
	if a.config.APIKey != "" {
		return a.config.APIKey, nil
	}

	a.mu.RLock()
	if a.token != "" && time.Until(a.expiresAt) > 30*time.Second {
		token := a.token
		a.mu.RUnlock()
		return token, nil
	}
	a.mu.RUnlock()

	return a.refresh(ctx)
}

func (a *authProvider) refresh(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.token != "" && time.Until(a.expiresAt) > 30*time.Second {
		return a.token, nil
	}

	if a.config.ClientID == "" || a.config.ClientSecret == "" {
		return "", fmt.Errorf("client credentials not set")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.config.TokenURL, nil)
	if err != nil {
		return "", err
	}

	authHeader := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", a.config.ClientID, a.config.ClientSecret)))
	req.Header.Set("Authorization", "Basic "+authHeader)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token endpoint returned status %d", resp.StatusCode)
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}

	token, _ := payload["access_token"].(string)
	if token == "" {
		return "", fmt.Errorf("token response missing access_token")
	}

	expiresIn := parseExpiresIn(payload["expires_in"])
	if expiresIn <= 0 {
		expiresIn = 3600
	}

	a.token = token
	a.expiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)

	return a.token, nil
}

func parseExpiresIn(value interface{}) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case json.Number:
		i, _ := v.Int64()
		return int(i)
	case string:
		if v == "" {
			return 0
		}
		i, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return 0
		}
		return i
	default:
		return 0
	}
}
