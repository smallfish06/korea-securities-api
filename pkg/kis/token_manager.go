package kis

import (
	internalkis "github.com/smallfish06/krsec/internal/kis"
	tokencache "github.com/smallfish06/krsec/pkg/token"
)

// NewFileTokenManager creates a KIS file-backed token manager.
func NewFileTokenManager() tokencache.Manager {
	return internalkis.NewFileTokenManager()
}

// NewFileTokenManagerWithDir creates a KIS file-backed token manager with a fixed directory.
func NewFileTokenManagerWithDir(dir string) tokencache.Manager {
	return internalkis.NewFileTokenManagerWithDir(dir)
}
