package kis

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTokenDir_UsesConfiguredDir(t *testing.T) {
	tm := NewFileTokenManagerWithDir("/tmp/krsec-token-cache")
	got, err := tm.tokenDir()
	if err != nil {
		t.Fatalf("tokenDir returned error: %v", err)
	}
	if got != "/tmp/krsec-token-cache" {
		t.Fatalf("unexpected token dir: got=%q", got)
	}
}

func TestFindProjectRoot(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "repo")
	sub := filepath.Join(root, "internal", "kis")

	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatalf("write go.mod failed: %v", err)
	}

	got, ok := findProjectRoot(sub)
	if !ok {
		t.Fatalf("findProjectRoot did not find project root")
	}
	if got != root {
		t.Fatalf("unexpected project root: got=%q want=%q", got, root)
	}
}

func TestDefaultTokenDir_UsesProjectRoot(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "repo")
	sub := filepath.Join(root, "cmd", "krsec")

	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatalf("write go.mod failed: %v", err)
	}

	got, err := defaultTokenDir(sub)
	if err != nil {
		t.Fatalf("defaultTokenDir returned error: %v", err)
	}

	want := filepath.Join(root, ".krsec", "tokens")
	if got != want {
		t.Fatalf("unexpected token dir: got=%q want=%q", got, want)
	}
}
