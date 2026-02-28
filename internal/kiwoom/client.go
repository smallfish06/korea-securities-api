package kiwoom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/smallfish06/korea-securities-api/pkg/broker"
)

const (
	// BaseURLReal is Kiwoom production REST domain.
	BaseURLReal = "https://api.kiwoom.com"
	// BaseURLSandbox is Kiwoom mock REST domain.
	BaseURLSandbox = "https://mockapi.kiwoom.com"
)

// Client is a Kiwoom HTTP client with token caching and API-id routing.
type Client struct {
	baseURL    string
	httpClient *http.Client

	mu          sync.RWMutex
	appKey      string
	appSecret   string
	accessToken string
	expiresAt   time.Time

	apiLimiter   *RateLimiter
	tokenManager TokenManager
}

// callOptions controls optional Kiwoom continuation headers.
type callOptions struct {
	ContYN  string
	NextKey string
	Headers map[string]string
}

// callResult wraps parsed body and raw response headers.
type callResult struct {
	Headers http.Header
	Body    map[string]interface{}
}

// NewClient creates a new Kiwoom client.
func NewClient(sandbox bool) *Client {
	return NewClientWithTokenManager(sandbox, nil)
}

// NewClientWithTokenManager creates a client with injected token manager.
func NewClientWithTokenManager(sandbox bool, tm TokenManager) *Client {
	baseURL := BaseURLReal
	if sandbox {
		baseURL = BaseURLSandbox
	}
	if tm == nil {
		tm = GetTokenManager()
	}
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiLimiter:   NewRateLimiter(8.0),
		tokenManager: tm,
	}
}

// Name returns broker name.
func (c *Client) Name() string {
	return "KIWOOM"
}

// SetCredentials sets app credentials on this client.
func (c *Client) SetCredentials(appKey, appSecret string) {
	c.mu.Lock()
	c.appKey = strings.TrimSpace(appKey)
	c.appSecret = strings.TrimSpace(appSecret)
	c.mu.Unlock()
}

// SetBaseURL overrides the API base URL.
// Primarily useful for tests or private/proxy deployments.
func (c *Client) SetBaseURL(baseURL string) {
	c.mu.Lock()
	c.baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	c.mu.Unlock()
}

func (c *Client) setToken(token string, expiresAt time.Time) {
	c.mu.Lock()
	c.accessToken = token
	c.expiresAt = expiresAt
	c.mu.Unlock()
}

func (c *Client) getToken() (string, time.Time) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.accessToken, c.expiresAt
}

func (c *Client) getCredentials() (string, string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.appKey, c.appSecret
}

func (c *Client) isTokenValid() bool {
	appKey, _ := c.getCredentials()
	tm := c.tokenManager
	if tm == nil {
		tm = GetTokenManager()
	}

	if appKey != "" {
		if token, expiresAt, ok := tm.GetToken(appKey); ok {
			cached, _ := c.getToken()
			if cached != token {
				c.setToken(token, expiresAt)
			}
			return true
		}
	}

	_, expiresAt := c.getToken()
	return time.Now().Before(expiresAt.Add(-2 * time.Minute))
}

// Catalog returns all known Kiwoom API specs.
func (c *Client) Catalog() ([]APISpec, error) {
	return ListAPISpecs()
}

// call executes one Kiwoom REST API request by api-id.
func (c *Client) call(ctx context.Context, apiID string, body map[string]interface{}, opts callOptions) (*callResult, error) {
	spec, ok, err := LookupAPISpec(apiID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("unknown kiwoom api-id: %s", apiID)
	}
	if strings.Contains(strings.ToLower(spec.Path), "websocket") || strings.HasPrefix(strings.ToLower(spec.RealDomain), "wss://") {
		return nil, fmt.Errorf("api-id %s is websocket-only and not supported by REST call", apiID)
	}
	if body == nil {
		body = map[string]interface{}{}
	}

	headers, result, err := c.doRequest(ctx, spec, apiID, body, opts)
	if err != nil {
		return nil, err
	}

	if code := parseReturnCode(result["return_code"]); code != 0 {
		msg := strings.TrimSpace(toString(result["return_msg"]))
		return nil, wrapCallError(apiID, code, msg)
	}

	return &callResult{Headers: headers, Body: result}, nil
}

func (c *Client) doRequest(ctx context.Context, spec APISpec, apiID string, body map[string]interface{}, opts callOptions) (http.Header, map[string]interface{}, error) {
	c.apiLimiter.Wait()

	if !c.isTokenValid() {
		appKey, appSecret := c.getCredentials()
		if appKey == "" || appSecret == "" {
			return nil, nil, fmt.Errorf("missing credentials for token refresh")
		}
		if _, err := c.Authenticate(ctx, broker.Credentials{AppKey: appKey, AppSecret: appSecret}); err != nil {
			return nil, nil, fmt.Errorf("token refresh failed: %w", err)
		}
	}

	url := strings.TrimRight(c.baseURL, "/") + spec.Path
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, spec.Method, url, bytes.NewReader(payload))
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}

	token, _ := c.getToken()
	req.Header.Set("Authorization", "Bearer "+token)
	if strings.TrimSpace(spec.ContentType) != "" {
		req.Header.Set("Content-Type", spec.ContentType)
	} else {
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	}
	req.Header.Set("api-id", apiID)

	if cont := strings.TrimSpace(opts.ContYN); cont != "" {
		req.Header.Set("cont-yn", cont)
	}
	if next := strings.TrimSpace(opts.NextKey); next != "" {
		req.Header.Set("next-key", next)
	}
	for k, v := range opts.Headers {
		key := strings.TrimSpace(k)
		if key == "" {
			continue
		}
		req.Header.Set(key, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		if code, msg, ok := parseErrorPayload(bodyBytes); ok {
			if code == 0 {
				code = resp.StatusCode
			}
			return nil, nil, wrapCallError(apiID, code, msg)
		}
		msg := strings.TrimSpace(string(bodyBytes))
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		return nil, nil, wrapCallError(apiID, resp.StatusCode, msg)
	}

	out := make(map[string]interface{})
	if len(bytes.TrimSpace(bodyBytes)) > 0 {
		if err := json.Unmarshal(bodyBytes, &out); err != nil {
			return nil, nil, fmt.Errorf("decode response: %w", err)
		}
	}

	return resp.Header.Clone(), out, nil
}

func parseReturnCode(v interface{}) int {
	switch t := v.(type) {
	case nil:
		return 0
	case int:
		return t
	case int8:
		return int(t)
	case int16:
		return int(t)
	case int32:
		return int(t)
	case int64:
		return int(t)
	case float32:
		return int(t)
	case float64:
		return int(t)
	case json.Number:
		n, err := t.Int64()
		if err == nil {
			return int(n)
		}
		f, err := t.Float64()
		if err == nil {
			return int(f)
		}
		return 0
	default:
		s := strings.TrimSpace(toString(v))
		if s == "" {
			return 0
		}
		code, err := strconv.Atoi(s)
		if err != nil {
			return 0
		}
		return code
	}
}

func toString(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case json.Number:
		return t.String()
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%f", t)
	case int:
		return fmt.Sprintf("%d", t)
	case int64:
		return fmt.Sprintf("%d", t)
	case bool:
		if t {
			return "true"
		}
		return "false"
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return ""
		}
		return string(b)
	}
}
