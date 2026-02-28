package adapter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const maxPersistedOrderContexts = 300

type orderContextState struct {
	Version   int                     `json:"version"`
	UpdatedAt time.Time               `json:"updated_at"`
	Orders    map[string]orderContext `json:"orders"`
}

func (a *Adapter) loadOrderContexts() error {
	path, err := a.orderContextFilePath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var state orderContextState
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}
	if len(state.Orders) == 0 {
		return nil
	}

	a.mu.Lock()
	for orderID, meta := range state.Orders {
		a.orders[orderID] = meta
	}
	a.compactOrderContextsLocked(maxPersistedOrderContexts)
	a.mu.Unlock()
	return nil
}

func (a *Adapter) persistOrderContexts() error {
	path, err := a.orderContextFilePath()
	if err != nil {
		return err
	}

	a.mu.RLock()
	snapshot := make(map[string]orderContext, len(a.orders))
	for orderID, meta := range a.orders {
		snapshot[orderID] = meta
	}
	a.mu.RUnlock()

	state := orderContextState{
		Version:   1,
		UpdatedAt: time.Now(),
		Orders:    snapshot,
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (a *Adapter) compactOrderContextsLocked(limit int) {
	if limit <= 0 || len(a.orders) <= limit {
		return
	}

	type row struct {
		orderID   string
		updatedAt time.Time
	}
	rows := make([]row, 0, len(a.orders))
	for orderID, meta := range a.orders {
		rows = append(rows, row{orderID: orderID, updatedAt: meta.UpdatedAt})
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].updatedAt.After(rows[j].updatedAt)
	})

	keep := make(map[string]struct{}, limit)
	for i := 0; i < limit && i < len(rows); i++ {
		keep[rows[i].orderID] = struct{}{}
	}
	for orderID := range a.orders {
		if _, ok := keep[orderID]; !ok {
			delete(a.orders, orderID)
		}
	}
}

func (a *Adapter) orderContextFilePath() (string, error) {
	baseDir := a.orderDir
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		baseDir = filepath.Join(home, ".kr-broker", "orders")
	}

	env := "real"
	if a.sandbox {
		env = "sandbox"
	}
	file := fmt.Sprintf("%s-%s-%s.json", env, a.accountID, a.accountPrdtCD)
	return filepath.Join(baseDir, file), nil
}
