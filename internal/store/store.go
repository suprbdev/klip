package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Entry struct {
	Value     string     `json:"value"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"` // nil = never expires
	Temp      bool       `json:"temp"`                 // informational only
}

type Store struct {
	Version int              `json:"version"`
	Entries map[string]Entry `json:"entries"`
}

// Load reads the JSON store from path. If file does not exist, returns an empty store.
func Load(path string) (*Store, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("cannot create config dir: %w", err)
	}
	b, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("read error: %w", err)
	}
	s := &Store{
		Version: 1,
		Entries: map[string]Entry{},
	}
	if len(b) == 0 || os.IsNotExist(err) {
		return s, nil
	}
	if err := json.Unmarshal(b, s); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w", err)
	}
	if s.Version == 0 {
		s.Version = 1
	}
	if s.Entries == nil {
		s.Entries = map[string]Entry{}
	}
	return s, nil
}

// Save writes the store atomically (temp file + rename).
func Save(path string, s *Store) error {
	tmpPath := path + ".tmp"
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err := os.WriteFile(tmpPath, b, 0o600); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}
	// rename is atomic on POSIX and Windows (replace semantics)
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}

// Get returns the entry for key if present and not expired.
func (s *Store) Get(key string) (Entry, bool) {
	e, ok := s.Entries[key]
	if !ok || IsExpired(e, time.Now()) {
		return Entry{}, false
	}
	return e, true
}

// Set stores/overwrites an entry.
func (s *Store) Set(key string, e Entry) {
	s.Entries[key] = e
}

// Delete removes a key; returns true if something was deleted.
func (s *Store) Delete(key string) bool {
	if _, ok := s.Entries[key]; ok {
		delete(s.Entries, key)
		return true
	}
	return false
}

// List returns a map of entries; includeExpired controls whether expired ones are kept.
func (s *Store) List(includeExpired bool) map[string]Entry {
	out := make(map[string]Entry, len(s.Entries))
	now := time.Now()
	for k, e := range s.Entries {
		if !includeExpired && IsExpired(e, now) {
			continue
		}
		out[k] = e
	}
	return out
}

// PurgeExpired removes all expired entries and returns the number removed.
func (s *Store) PurgeExpired() int {
	now := time.Now()
	cnt := 0
	for k, e := range s.Entries {
		if IsExpired(e, now) {
			delete(s.Entries, k)
			cnt++
		}
	}
	return cnt
}

// IsExpired reports whether an entry is expired.
func IsExpired(e Entry, now time.Time) bool {
	if e.ExpiresAt == nil {
		return false
	}
	return now.After(*e.ExpiresAt)
}
