package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type SenderLimiter interface {
	Allow(ctx context.Context, account string, senderID int64) (bool, error)
	Close() error
}

type BaseDB struct {
	path string
	mu   sync.Mutex
	Seen map[string]int64 `json:"seen"`
}

func OpenBaseDB(path string) (*BaseDB, error) {
	if path == "" {
		path = "data/base.json"
	}
	db := &BaseDB{
		path: path,
		Seen: make(map[string]int64),
	}

	data, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(data, &db.Seen)
	}

	db.cleanup()
	return db, nil
}

func (b *BaseDB) cleanup() {
	cutoff := time.Now().Add(-24 * time.Hour).Unix()
	changed := false
	for k, ts := range b.Seen {
		if ts < cutoff {
			delete(b.Seen, k)
			changed = true
		}
	}
	if changed {
		_ = b.save()
	}
}

func (b *BaseDB) save() error {
	data, err := json.MarshalIndent(b.Seen, "", "  ")
	if err != nil {
		return err
	}
	_ = os.MkdirAll(filepath.Dir(b.path), 0o700)
	return os.WriteFile(b.path, data, 0o600)
}

func (b *BaseDB) Allow(ctx context.Context, account string, senderID int64) (bool, error) {
	if senderID == 0 {
		return true, nil
	}

	key := fmt.Sprintf("%s:%d", account, senderID)

	b.mu.Lock()
	defer b.mu.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour).Unix()
	if ts, ok := b.Seen[key]; ok && ts >= cutoff {
		return false, nil
	}

	b.Seen[key] = time.Now().Unix()
	_ = b.save()
	
	return true, nil
}

func (b *BaseDB) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.save()
}
