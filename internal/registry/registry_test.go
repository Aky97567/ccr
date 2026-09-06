package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingFileIsEmpty(t *testing.T) {
	entries, err := Load(filepath.Join(t.TempDir(), "nope.jsonl"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("want 0 entries, got %d", len(entries))
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sessions.jsonl")
	in := []Entry{
		{Name: "a", SessionID: "id-a", Cwd: "/Users/akshay/Code/personal/cv repos/1c2p", Note: "quote's ok", Updated: "2026-09-06T15:36:51+0200"},
		{Name: "b", SessionID: "id-b", Cwd: "/tmp/b", Note: "", Updated: "2026-09-06T15:50:03+0200"},
	}
	if err := Save(path, in); err != nil {
		t.Fatalf("Save: %v", err)
	}
	out, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(out) != len(in) {
		t.Fatalf("want %d entries, got %d", len(in), len(out))
	}
	for i := range in {
		if out[i] != in[i] {
			t.Errorf("entry %d: got %+v, want %+v", i, out[i], in[i])
		}
	}
}

func TestSaveLeavesNoTempFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sessions.jsonl")
	if err := Save(path, []Entry{{Name: "a", SessionID: "id-a"}}); err != nil {
		t.Fatal(err)
	}
	names, _ := filepath.Glob(filepath.Join(dir, "*"))
	if len(names) != 1 || filepath.Base(names[0]) != "sessions.jsonl" {
		t.Fatalf("expected only sessions.jsonl, got %v", names)
	}
}

func TestLoadMalformedLineErrorsWithLineNumber(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sessions.jsonl")
	os.WriteFile(path, []byte(`{"name":"ok","session_id":"x"}`+"\n"+`{not json}`+"\n"), 0o644)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected an error for malformed line")
	}
	if got := err.Error(); !contains(got, ":2:") {
		t.Fatalf("error should point at line 2, got %q", got)
	}
}

func TestUpsertReplacesBySessionID(t *testing.T) {
	entries := []Entry{
		{Name: "old", SessionID: "id-1", Note: "before"},
		{Name: "keep", SessionID: "id-2"},
	}
	entries = Upsert(entries, Entry{Name: "new", SessionID: "id-1", Note: "after"})
	if len(entries) != 2 {
		t.Fatalf("want 2 entries, got %d", len(entries))
	}
	if entries[0].Name != "new" || entries[0].Note != "after" {
		t.Errorf("id-1 not replaced: %+v", entries[0])
	}
	if entries[1].Name != "keep" {
		t.Errorf("id-2 disturbed: %+v", entries[1])
	}
}

func TestUpsertAppendsWhenNew(t *testing.T) {
	entries := Upsert([]Entry{{SessionID: "id-1"}}, Entry{Name: "n", SessionID: "id-2"})
	if len(entries) != 2 || entries[1].SessionID != "id-2" {
		t.Fatalf("append failed: %+v", entries)
	}
}

func TestBackupCopiesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sessions.jsonl")
	os.WriteFile(path, []byte("hello\n"), 0o644)
	bak, err := Backup(path)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(bak)
	if string(got) != "hello\n" {
		t.Fatalf("backup content = %q", got)
	}
}

func TestBackupMissingFileIsNoop(t *testing.T) {
	bak, err := Backup(filepath.Join(t.TempDir(), "nope.jsonl"))
	if err != nil || bak != "" {
		t.Fatalf("want (\"\", nil), got (%q, %v)", bak, err)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
