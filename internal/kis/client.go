package kis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/smallfish06/kr-broker-api/pkg/broker"
)

const (
	// BaseURLReal is the production base URL
	BaseURLReal = "https://openapi.koreainvestment.com:9443"
	// BaseURLSandbox is the sandbox base URL
	BaseURLSandbox = "https://openapivts.koreainvestment.com:29443"
)

// Client is the KIS HTTP client
type Client struct {
	baseURL    string
	httpClient *http.Client
	appKey     string
	appSecret  string

	mu          sync.RWMutex
	accessToken string
	expiresAt   time.Time

	apiLimiter   *RateLimiter // Rate limiter for API requests (15/sec)
	tokenManager TokenManager
}

// NewClient creates a new KIS client
func NewClient(sandbox bool) *Client {
	return NewClientWithTokenManager(sandbox, nil)
}

// NewClientWithTokenManager creates a new KIS client with an injected token manager.
// When tokenManager is nil, the global default manager is used.
func NewClientWithTokenManager(sandbox bool, tokenManager TokenManager) *Client {
	baseURL := BaseURLReal
	if sandbox {
		baseURL = BaseURLSandbox
	}
	if tokenManager == nil {
		tokenManager = GetTokenManager()
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiLimiter:   NewRateLimiter(15.0), // 15 requests/sec (conservative, KIS allows 20)
		tokenManager: tokenManager,
	}
}

// Name returns the broker name
func (c *Client) Name() string {
	return "KIS"
}

// SetCredentials sets the app key and secret
func (c *Client) SetCredentials(appKey, appSecret string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.appKey = appKey
	c.appSecret = appSecret
}

// SetToken sets the access token
func (c *Client) setToken(token string, expiresAt time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.accessToken = token
	c.expiresAt = expiresAt
}

// GetToken returns the current access token
func (c *Client) getToken() (string, time.Time) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.accessToken, c.expiresAt
}

// IsTokenValid checks if the current token is valid.
// It checks the injected token manager first, then falls back to local token.
func (c *Client) isTokenValid() bool {
	c.mu.RLock()
	appKey := c.appKey
	c.mu.RUnlock()

	tm := c.tokenManager
	if tm == nil {
		tm = GetTokenManager()
	}

	// Check token manager first
	if appKey != "" {
		if token, expiresAt, ok := tm.GetToken(appKey); ok {
			// Update local cache if different
			localToken, _ := c.getToken()
			if localToken != token {
				c.setToken(token, expiresAt)
			}
			return true
		}
	}

	// Fall back to local token check
	_, expiresAt := c.getToken()
	return time.Now().Before(expiresAt.Add(-5 * time.Minute)) // 5분 여유
}

// doRequest performs an HTTP request with authentication headers
func (c *Client) doRequest(ctx context.Context, method, path string, trID string, body interface{}, result interface{}) error {
	// Apply rate limiting
	c.apiLimiter.Wait()

	// Check token validity
	if !c.isTokenValid() {
		c.mu.RLock()
		creds := broker.Credentials{
			AppKey:    c.appKey,
			AppSecret: c.appSecret,
		}
		c.mu.RUnlock()

		if _, err := c.Authenticate(ctx, creds); err != nil {
			return fmt.Errorf("token refresh failed: %w", err)
		}
	}

	url := c.baseURL + path
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	log.Printf("DEBUG: %s %s (tr_id: %s)", method, url, trID)
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	token, _ := c.getToken()
	c.mu.RLock()
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("appkey", c.appKey)
	req.Header.Set("appsecret", c.appSecret)
	c.mu.RUnlock()

	if trID != "" {
		req.Header.Set("tr_id", trID)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}
