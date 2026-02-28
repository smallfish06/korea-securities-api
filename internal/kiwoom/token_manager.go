package kiwoom

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// TokenManager defines token cache and token-issuance throttling behavior.
type TokenManager interface {
	GetToken(appKey string) (string, time.Time, bool)
	SetToken(appKey, token string, expiresAt time.Time) error
	WaitForAuth(appKey string)
}

// FileTokenManager stores tokens in memory and on disk.
type FileTokenManager struct {
	mu           sync.RWMutex
	tokens       map[string]*tokenEntry
	authLimiters map[string]*RateLimiter
	authMu       sync.Mutex
	dir          string
}

type tokenEntry struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
	AppKey      string    `json:"app_key"`
}

var (
	globalTokenManager   TokenManager
	globalTokenManagerMu sync.RWMutex
)

// NewFileTokenManager creates the default file-backed token manager.
func NewFileTokenManager() *FileTokenManager {
	return NewFileTokenManagerWithDir("")
}

// NewFileTokenManagerWithDir creates a file-backed manager with optional fixed directory.
func NewFileTokenManagerWithDir(dir string) *FileTokenManager {
	tm := &FileTokenManager{
		tokens:       make(map[string]*tokenEntry),
		authLimiters: make(map[string]*RateLimiter),
		dir:          strings.TrimSpace(dir),
	}
	tm.loadAll()
	return tm
}

// GetTokenManager returns the global token manager.
func GetTokenManager() TokenManager {
	globalTokenManagerMu.RLock()
	tm := globalTokenManager
	globalTokenManagerMu.RUnlock()
	if tm != nil {
		return tm
	}

	globalTokenManagerMu.Lock()
	defer globalTokenManagerMu.Unlock()
	if globalTokenManager == nil {
		globalTokenManager = NewFileTokenManager()
	}
	return globalTokenManager
}

// SetGlobalTokenManager overrides the global token manager implementation.
func SetGlobalTokenManager(tm TokenManager) {
	if tm == nil {
		tm = NewFileTokenManager()
	}
	globalTokenManagerMu.Lock()
	globalTokenManager = tm
	globalTokenManagerMu.Unlock()
}

// GetToken returns the cached token if still valid.
func (tm *FileTokenManager) GetToken(appKey string) (string, time.Time, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	entry, ok := tm.tokens[appKey]
	if !ok {
		return "", time.Time{}, false
	}
	if time.Now().After(entry.ExpiresAt.Add(-3 * time.Minute)) {
		return "", time.Time{}, false
	}
	return entry.AccessToken, entry.ExpiresAt, true
}

// SetToken stores the token in memory and on disk.
func (tm *FileTokenManager) SetToken(appKey, token string, expiresAt time.Time) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	entry := &tokenEntry{AccessToken: token, ExpiresAt: expiresAt, AppKey: appKey}
	tm.tokens[appKey] = entry
	return tm.save(appKey, entry)
}

// WaitForAuth enforces per-appkey token issuance limits.
func (tm *FileTokenManager) WaitForAuth(appKey string) {
	tm.authMu.Lock()
	limiter, ok := tm.authLimiters[appKey]
	if !ok {
		limiter = NewRateLimiter(1.0 / 60.0)
		tm.authLimiters[appKey] = limiter
	}
	tm.authMu.Unlock()
	limiter.Wait()
}

func (tm *FileTokenManager) tokenDir() (string, error) {
	if tm.dir != "" {
		return tm.dir, nil
	}
	cwd, _ := os.Getwd()
	return defaultTokenDir(cwd)
}

func defaultTokenDir(cwd string) (string, error) {
	if strings.TrimSpace(cwd) != "" {
		if root, ok := findProjectRoot(cwd); ok {
			return filepath.Join(root, ".kr-broker", "tokens"), nil
		}
	}

	cacheDir, err := os.UserCacheDir()
	if err == nil && strings.TrimSpace(cacheDir) != "" {
		return filepath.Join(cacheDir, "kr-broker", "tokens"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get user home dir: %w", err)
	}
	return filepath.Join(home, ".kr-broker", "tokens"), nil
}

func findProjectRoot(start string) (string, bool) {
	current := start
	for {
		if st, err := os.Stat(filepath.Join(current, "go.mod")); err == nil && !st.IsDir() {
			return current, true
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", false
		}
		current = parent
	}
}

func hashAppKey(appKey string) string {
	h := sha256.Sum256([]byte(appKey))
	return hex.EncodeToString(h[:])[:12]
}

func (tm *FileTokenManager) loadAll() {
	dir, err := tm.tokenDir()
	if err != nil {
		return
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "kiwoom-") || filepath.Ext(name) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}

		var te tokenEntry
		if err := json.Unmarshal(data, &te); err != nil {
			continue
		}
		if te.AppKey == "" || time.Now().After(te.ExpiresAt.Add(-3*time.Minute)) {
			_ = os.Remove(filepath.Join(dir, name))
			continue
		}
		tm.tokens[te.AppKey] = &te
	}
}

func (tm *FileTokenManager) save(appKey string, entry *tokenEntry) error {
	dir, err := tm.tokenDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create token dir: %w", err)
	}

	path := filepath.Join(dir, "kiwoom-"+hashAppKey(appKey)+".json")
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal token: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write token file: %w", err)
	}
	return nil
}
