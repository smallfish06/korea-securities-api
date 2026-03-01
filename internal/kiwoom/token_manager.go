package kiwoom

import (
	"strings"
	"sync"
	"time"

	"github.com/smallfish06/krsec/internal/filetoken"
	tokencache "github.com/smallfish06/krsec/pkg/token"
)

// FileTokenManager stores tokens in memory and on disk.
type FileTokenManager struct {
	*filetoken.Manager
}

var (
	globalTokenManager   tokencache.Manager
	globalTokenManagerMu sync.RWMutex
)

// NewFileTokenManager creates the default file-backed token manager.
func NewFileTokenManager() *FileTokenManager {
	return NewFileTokenManagerWithDir("")
}

// NewFileTokenManagerWithDir creates a file-backed manager with optional fixed directory.
func NewFileTokenManagerWithDir(dir string) *FileTokenManager {
	return &FileTokenManager{
		Manager: filetoken.New(filetoken.Options{
			Dir:                 strings.TrimSpace(dir),
			AuthLimiterName:     "kiwoom-auth",
			ValidityBuffer:      3 * time.Minute,
			AllowFileName:       filetoken.PrefixedJSONFileOnly("kiwoom-"),
			BuildFileName:       filetoken.PrefixedHashedFileName("kiwoom-"),
			RequireAppKeyOnLoad: true,
		}),
	}
}

// GetTokenManager returns the global token manager.
func GetTokenManager() tokencache.Manager {
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
