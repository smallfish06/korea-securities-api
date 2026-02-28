package kis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/smallfish06/korea-securities-api/pkg/broker"
)

// TokenResponse represents the KIS token response
type TokenResponse struct {
	AccessToken           string `json:"access_token"`
	AccessTokenExpired    string `json:"access_token_token_expired"`
	TokenType             string `json:"token_type"`
	ExpiresIn             int    `json:"expires_in"`
	AccessTokenExpiresStr string `json:"access_token_expires"`
}

// Authenticate authenticates with KIS and returns a token
func (c *Client) Authenticate(ctx context.Context, creds broker.Credentials) (*broker.Token, error) {
	c.SetCredentials(creds.AppKey, creds.AppSecret)

	// Check token manager first (shared cache across clients)
	tm := c.tokenManager
	if tm == nil {
		tm = GetTokenManager()
	}
	if token, expiresAt, ok := tm.GetToken(creds.AppKey); ok {
		c.setToken(token, expiresAt)
		return &broker.Token{
			AccessToken: token,
			TokenType:   "Bearer",
			ExpiresAt:   expiresAt,
		}, nil
	}

	// Apply token issuance rate limit (1/minute per appkey)
	tm.WaitForAuth(creds.AppKey)

	reqBody := map[string]string{
		"grant_type": "client_credentials",
		"appkey":     creds.AppKey,
		"appsecret":  creds.AppSecret,
	}

	bodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := c.baseURL + "/oauth2/tokenP"
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("authentication failed: HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// Calculate expiration time
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// Store token in client
	c.setToken(tokenResp.AccessToken, expiresAt)

	// Store token in manager (shared cache + optional persistence)
	if err := tm.SetToken(creds.AppKey, tokenResp.AccessToken, expiresAt); err != nil {
		// Log error but don't fail - we have the token in memory
		fmt.Fprintf(os.Stderr, "Warning: failed to save token to disk: %v\n", err)
	}

	return &broker.Token{
		AccessToken: tokenResp.AccessToken,
		TokenType:   tokenResp.TokenType,
		ExpiresAt:   expiresAt,
	}, nil
}
