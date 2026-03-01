package filetoken

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestManagerSetAndGetToken(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	m := New(Options{
		Dir:             dir,
		AuthLimiterName: "test-auth",
		ValidityBuffer:  5 * time.Minute,
	})

	exp := time.Now().Add(30 * time.Minute)
	if err := m.SetToken("app", "tok", exp); err != nil {
		t.Fatalf("SetToken() error = %v", err)
	}

	token, expiresAt, ok := m.GetToken("app")
	if !ok {
		t.Fatalf("GetToken() expected token")
	}
	if token != "tok" {
		t.Fatalf("token = %q, want tok", token)
	}
	if expiresAt.IsZero() {
		t.Fatalf("expiresAt must be set")
	}
}

func TestManagerLoadFiltersAndPrunesExpired(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	valid := Entry{
		AccessToken: "ok",
		ExpiresAt:   time.Now().Add(20 * time.Minute),
		AppKey:      "app-ok",
	}
	expired := Entry{
		AccessToken: "old",
		ExpiresAt:   time.Now().Add(-1 * time.Hour),
		AppKey:      "app-old",
	}
	emptyKey := Entry{
		AccessToken: "nokey",
		ExpiresAt:   time.Now().Add(20 * time.Minute),
		AppKey:      "",
	}

	writeEntry := func(name string, e Entry) {
		t.Helper()
		data, err := json.Marshal(e)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	writeEntry("kiwoom-"+HashAppKey(valid.AppKey)+".json", valid)
	writeEntry("kiwoom-"+HashAppKey(expired.AppKey)+".json", expired)
	writeEntry("kiwoom-"+HashAppKey("empty")+".json", emptyKey)
	writeEntry(HashAppKey("other")+".json", valid) // non-prefixed file

	m := New(Options{
		Dir:                 dir,
		AuthLimiterName:     "kiwoom-auth",
		ValidityBuffer:      3 * time.Minute,
		AllowFileName:       PrefixedJSONFileOnly("kiwoom-"),
		BuildFileName:       PrefixedHashedFileName("kiwoom-"),
		RequireAppKeyOnLoad: true,
	})

	if _, _, ok := m.GetToken("app-ok"); !ok {
		t.Fatalf("expected prefixed valid token to load")
	}
	if _, _, ok := m.GetToken("app-old"); ok {
		t.Fatalf("expired token must not load")
	}

	// expired and empty-appkey files should be removed.
	for _, name := range []string{
		"kiwoom-" + HashAppKey(expired.AppKey) + ".json",
		"kiwoom-" + HashAppKey("empty") + ".json",
	} {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Fatalf("expected %s to be removed", name)
		}
	}
}

func TestFindProjectRoot(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	root := filepath.Join(base, "repo")
	sub := filepath.Join(root, "internal", "pkg")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	got, ok := FindProjectRoot(sub)
	if !ok || got != root {
		t.Fatalf("FindProjectRoot() = (%q,%v), want (%q,true)", got, ok, root)
	}
}
