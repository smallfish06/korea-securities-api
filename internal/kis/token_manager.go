package kis

import (
	"strings"
	"sync"
	"time"

	"github.com/smallfish06/krsec/internal/filetoken"
	tokencache "github.com/smallfish06/krsec/pkg/token"
)

// FileTokenManager stores tokens in memory and persists them to disk.
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

// NewFileTokenManagerWithDir creates a file-backed token manager with an optional fixed directory.
// When dir is empty, the default directory resolution is used.
func NewFileTokenManagerWithDir(dir string) *FileTokenManager {
	return &FileTokenManager{
		Manager: filetoken.New(filetoken.Options{
			Dir:                 strings.TrimSpace(dir),
			AuthLimiterName:     "kis-auth",
			ValidityBuffer:      5 * time.Minute,
			AllowFileName:       filetoken.JSONFileOnly,
			BuildFileName:       filetoken.DefaultHashedFileName,
			RequireAppKeyOnLoad: false,
		}),
	}
}

// GetTokenManager returns the global token manager.
// The default implementation is file-backed.
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

// tokenDir returns the directory where tokens are stored.
func (tm *FileTokenManager) tokenDir() (string, error) {
	return tm.TokenDir()
}

func defaultTokenDir(cwd string) (string, error) {
	return filetoken.DefaultTokenDir(cwd)
}

// findProjectRoot walks up from start and returns the directory containing go.mod.
func findProjectRoot(start string) (string, bool) {
	return filetoken.FindProjectRoot(start)
}
