package session

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCurrentIDPrefersEnvVar(t *testing.T) {
	t.Setenv(EnvVar, "  env-session-id  ")
	got, err := CurrentID("/anywhere")
	if err != nil {
		t.Fatal(err)
	}
	if got != "env-session-id" {
		t.Fatalf("got %q, want trimmed env value", got)
	}
}

func TestFindByCwdPicksNewestMatchingTranscript(t *testing.T) {
	root := t.TempDir()
	cwd := "/Users/akshay/Code/personal/crosslex"

	// A project dir whose encoded name is lossy on purpose; cwd is read from content.
	proj := filepath.Join(root, "-Users-akshay-Code-personal-crosslex")
	if err := os.MkdirAll(proj, 0o755); err != nil {
		t.Fatal(err)
	}

	older := filepath.Join(proj, "11111111-1111-1111-1111-111111111111.jsonl")
	newer := filepath.Join(proj, "22222222-2222-2222-2222-222222222222.jsonl")
	other := filepath.Join(proj, "33333333-3333-3333-3333-333333333333.jsonl")
	agent := filepath.Join(proj, "agent-99999999999999999.jsonl")

	write(t, older, `{"type":"summary","summary":"s"}`+"\n"+`{"cwd":"`+cwd+`"}`+"\n")
	write(t, newer, `{"cwd":"`+cwd+`"}`+"\n")
	write(t, other, `{"cwd":"/some/other/place"}`+"\n")
	write(t, agent, `{"cwd":"`+cwd+`"}`+"\n")

	past := time.Now().Add(-time.Hour)
	os.Chtimes(older, past, past)
	os.Chtimes(agent, time.Now().Add(time.Hour), time.Now().Add(time.Hour)) // newest, but ignored

	got, err := findByCwd(root, cwd)
	if err != nil {
		t.Fatal(err)
	}
	if got != "22222222-2222-2222-2222-222222222222" {
		t.Fatalf("got %q, want the newer matching transcript", got)
	}
}

func TestFindByCwdNoMatch(t *testing.T) {
	root := t.TempDir()
	proj := filepath.Join(root, "-x")
	os.MkdirAll(proj, 0o755)
	write(t, filepath.Join(proj, "44444444-4444-4444-4444-444444444444.jsonl"), `{"cwd":"/elsewhere"}`+"\n")

	got, err := findByCwd(root, "/not/here")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
