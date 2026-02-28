package kis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/smallfish06/krsec/internal/ratelimit"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// TokenManager defines token cache and token-issuance throttling behavior.
// Custom implementations can be injected into Client/Adapter constructors.
type TokenManager interface {
	GetToken(appKey string) (string, time.Time, bool)
	SetToken(appKey, token string, expiresAt time.Time) error
	WaitForAuth(appKey string)
}

// FileTokenManager stores tokens in memory and persists them to disk.
type FileTokenManager struct {
	mu     sync.RWMutex
	tokens map[string]*tokenEntry // key: appkey

	authLimiters   map[string]*ratelimit.Limiter // appkey -> rate limiter for token issuance
	authLimitersMu sync.Mutex
	dir            string
}

// tokenEntry represents a cached token
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

// NewFileTokenManagerWithDir creates a file-backed token manager with an optional fixed directory.
// When dir is empty, the default directory resolution is used.
func NewFileTokenManagerWithDir(dir string) *FileTokenManager {
	tm := &FileTokenManager{
		tokens:       make(map[string]*tokenEntry),
		authLimiters: make(map[string]*ratelimit.Limiter),
		dir:          strings.TrimSpace(dir),
	}
	tm.loadAll()
	return tm
}

// GetTokenManager returns the global token manager.
// The default implementation is file-backed.
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
// If tm is nil, the default file-backed manager is used.
func SetGlobalTokenManager(tm TokenManager) {
	if tm == nil {
		tm = NewFileTokenManager()
	}
	globalTokenManagerMu.Lock()
	globalTokenManager = tm
	globalTokenManagerMu.Unlock()
}

// GetToken returns the cached token for the given appkey
// Returns (token, expiresAt, found)
func (tm *FileTokenManager) GetToken(appKey string) (string, time.Time, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	entry, exists := tm.tokens[appKey]
	if !exists {
		return "", time.Time{}, false
	}

	// Check if token is still valid (5 minute buffer)
	if time.Now().After(entry.ExpiresAt.Add(-5 * time.Minute)) {
		return "", time.Time{}, false
	}

	return entry.AccessToken, entry.ExpiresAt, true
}

// SetToken stores the token for the given appkey (in memory and on disk)
func (tm *FileTokenManager) SetToken(appKey, token string, expiresAt time.Time) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	entry := &tokenEntry{
		AccessToken: token,
		ExpiresAt:   expiresAt,
		AppKey:      appKey,
	}

	tm.tokens[appKey] = entry
	return tm.save(appKey, entry)
}

// WaitForAuth enforces the per-appkey token issuance rate limit (1/minute)
func (tm *FileTokenManager) WaitForAuth(appKey string) {
	tm.authLimitersMu.Lock()
	limiter, exists := tm.authLimiters[appKey]
	if !exists {
		// 1 request per 60 seconds = 1/60 per second
		limiter = ratelimit.New("kis-auth", 1.0/60.0, 1)
		tm.authLimiters[appKey] = limiter
	}
	tm.authLimitersMu.Unlock()

	_ = limiter.Wait(context.Background())
}

// tokenDir returns the directory where tokens are stored.
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
			return filepath.Join(root, ".krsec", "tokens"), nil
		}
	}

	cacheDir, err := os.UserCacheDir()
	if err == nil && strings.TrimSpace(cacheDir) != "" {
		return filepath.Join(cacheDir, "krsec", "tokens"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get user home dir: %w", err)
	}
	return filepath.Join(home, ".krsec", "tokens"), nil
}

// findProjectRoot walks up from start and returns the directory containing go.mod.
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

// hashAppKey returns the first 12 characters of the sha256 hash of the appkey
func hashAppKey(appKey string) string {
	h := sha256.Sum256([]byte(appKey))
	return hex.EncodeToString(h[:])[:12]
}

// loadAll loads all token files from disk
func (tm *FileTokenManager) loadAll() {
	dir, err := tm.tokenDir()
	if err != nil {
		return
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0700); err != nil {
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var te tokenEntry
		if err := json.Unmarshal(data, &te); err != nil {
			continue
		}

		// Only load if token is still valid
		if time.Now().Before(te.ExpiresAt.Add(-5 * time.Minute)) {
			tm.tokens[te.AppKey] = &te
		} else {
			// Delete expired token file
			_ = os.Remove(filePath)
		}
	}
}

// save writes the token to disk
func (tm *FileTokenManager) save(appKey string, entry *tokenEntry) error {
	dir, err := tm.tokenDir()
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create token dir: %w", err)
	}

	filename := hashAppKey(appKey) + ".json"
	filePath := filepath.Join(dir, filename)

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal token entry: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("write token file: %w", err)
	}

	return nil
}
