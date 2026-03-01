package orderctxstore

import (
	"path/filepath"
	"testing"
	"time"
)

type testOrder struct {
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

func updatedAtOf(v testOrder) time.Time {
	return v.UpdatedAt
}

func TestLoadMissingFile(t *testing.T) {
	t.Parallel()

	orders := map[string]testOrder{}
	err := Load(filepath.Join(t.TempDir(), "missing.json"), orders, 10, updatedAtOf)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(orders) != 0 {
		t.Fatalf("expected empty map, got %d", len(orders))
	}
}

func TestCompactKeepsNewest(t *testing.T) {
	t.Parallel()

	now := time.Now()
	orders := map[string]testOrder{
		"o1": {Status: "a", UpdatedAt: now.Add(-3 * time.Hour)},
		"o2": {Status: "b", UpdatedAt: now.Add(-2 * time.Hour)},
		"o3": {Status: "c", UpdatedAt: now.Add(-1 * time.Hour)},
	}

	Compact(orders, 2, updatedAtOf)

	if len(orders) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(orders))
	}
	if _, ok := orders["o3"]; !ok {
		t.Fatalf("newest order must remain")
	}
	if _, ok := orders["o2"]; !ok {
		t.Fatalf("second newest order must remain")
	}
	if _, ok := orders["o1"]; ok {
		t.Fatalf("oldest order must be removed")
	}
}

func TestPersistThenLoad(t *testing.T) {
	t.Parallel()

	now := time.Now()
	source := map[string]testOrder{
		"o1": {Status: "pending", UpdatedAt: now.Add(-1 * time.Minute)},
		"o2": {Status: "filled", UpdatedAt: now},
	}
	path := filepath.Join(t.TempDir(), "orders.json")

	if err := Persist(path, source); err != nil {
		t.Fatalf("Persist() error = %v", err)
	}

	loaded := map[string]testOrder{}
	if err := Load(path, loaded, 10, updatedAtOf); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("expected 2 loaded orders, got %d", len(loaded))
	}
	if loaded["o1"].Status != "pending" || loaded["o2"].Status != "filled" {
		t.Fatalf("loaded values mismatch: %+v", loaded)
	}
}
