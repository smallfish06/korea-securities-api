package kiwoom

import (
	internalkiwoom "github.com/smallfish06/krsec/internal/kiwoom"
	tokencache "github.com/smallfish06/krsec/pkg/token"
)

// NewFileTokenManager creates a Kiwoom file-backed token manager.
func NewFileTokenManager() tokencache.Manager {
	return internalkiwoom.NewFileTokenManager()
}

// NewFileTokenManagerWithDir creates a Kiwoom file-backed token manager with a fixed directory.
func NewFileTokenManagerWithDir(dir string) tokencache.Manager {
	return internalkiwoom.NewFileTokenManagerWithDir(dir)
}
