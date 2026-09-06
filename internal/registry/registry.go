// Package registry reads and writes the machine-local session registry:
// one JSON object per line at $CCR_REGISTRY (default ~/.local/share/ccr/sessions.jsonl).
package registry

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Entry is one starred session. The on-disk schema is exactly these five fields.
type Entry struct {
	Name      string `json:"name"`
	SessionID string `json:"session_id"`
	Cwd       string `json:"cwd"`
	Note      string `json:"note"`
	Updated   string `json:"updated"`
}

// TimeLayout matches the format the existing registry rows use
// (`date +%Y-%m-%dT%H:%M:%S%z`), e.g. 2026-09-06T15:36:51+0200.
const TimeLayout = "2006-01-02T15:04:05-0700"

// Now returns the current time formatted for the Updated field.
func Now() string { return time.Now().Format(TimeLayout) }

// Path is the registry location: $CCR_REGISTRY, or ~/.local/share/ccr/sessions.jsonl.
func Path() (string, error) {
	if p := os.Getenv("CCR_REGISTRY"); p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "ccr", "sessions.jsonl"), nil
}

// Load parses the registry. A missing file is not an error (returns an empty slice).
// A malformed line IS an error, identified by line number, so data is never silently dropped.
func Load(path string) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Entry{}, nil
		}
		return nil, err
	}
	defer f.Close()

	var entries []Entry
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for line := 1; sc.Scan(); line++ {
		raw := bytes.TrimSpace(sc.Bytes())
		if len(raw) == 0 {
			continue
		}
		var e Entry
		if err := json.Unmarshal(raw, &e); err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, line, err)
		}
		entries = append(entries, e)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

// Save writes entries atomically: a temp file in the same directory, fsync, then rename.
// The registry is never truncated in place.
func Save(path string, entries []Entry) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".sessions-*.jsonl")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op after a successful rename

	w := bufio.NewWriter(tmp)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	for _, e := range entries {
		if err := enc.Encode(e); err != nil {
			tmp.Close()
			return err
		}
	}
	if err := w.Flush(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

// Upsert replaces the entry with the same SessionID, or appends e if none matches.
func Upsert(entries []Entry, e Entry) []Entry {
	for i := range entries {
		if entries[i].SessionID == e.SessionID {
			entries[i] = e
			return entries
		}
	}
	return append(entries, e)
}

// Backup copies path to path.<unix>.bak and returns the backup path.
// A missing source file is not an error (returns "", nil).
func Backup(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	bak := fmt.Sprintf("%s.%d.bak", path, time.Now().Unix())
	if err := os.WriteFile(bak, data, 0o644); err != nil {
		return "", err
	}
	return bak, nil
}
