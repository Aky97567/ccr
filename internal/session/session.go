// Package session resolves the identifier of the Claude Code session that ccr is
// being run from.
package session

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// EnvVar is exported to Bash subprocesses by Claude Code >= 2.1.132.
const EnvVar = "CLAUDE_CODE_SESSION_ID"

// CurrentID returns the current session's UUID.
//
// Primary: the $CLAUDE_CODE_SESSION_ID environment variable.
// Fallback: the most recently modified non-"agent-" transcript under
// ~/.claude/projects/ whose recorded cwd equals the given directory.
func CurrentID(cwd string) (string, error) {
	if id := strings.TrimSpace(os.Getenv(EnvVar)); id != "" {
		return id, nil
	}
	root, err := ProjectsDir()
	if err != nil {
		return "", err
	}
	id, err := findByCwd(root, cwd)
	if err != nil {
		return "", err
	}
	if id == "" {
		return "", fmt.Errorf(
			"could not determine the current session id: %s is unset and no transcript under ~/.claude/projects/ records cwd=%q\n"+
				"run this from the session's project directory, or pass --session-id",
			EnvVar, cwd)
	}
	return id, nil
}

// ProjectsDir is ~/.claude/projects.
func ProjectsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "projects"), nil
}

func findByCwd(root, cwd string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(root, "*", "*.jsonl"))
	if err != nil {
		return "", err
	}

	type cand struct {
		id  string
		mod int64
	}
	var cands []cand
	for _, m := range matches {
		base := filepath.Base(m)
		if strings.HasPrefix(base, "agent-") {
			continue
		}
		fi, err := os.Stat(m)
		if err != nil {
			continue
		}
		if transcriptCwd(m) != cwd {
			continue
		}
		id := strings.TrimSuffix(base, ".jsonl")
		cands = append(cands, cand{id: id, mod: fi.ModTime().UnixNano()})
	}
	if len(cands) == 0 {
		return "", nil
	}
	sort.Slice(cands, func(i, j int) bool { return cands[i].mod > cands[j].mod })
	return cands[0].id, nil
}

// transcriptCwd returns the first non-empty "cwd" value in a transcript, or "".
func transcriptCwd(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for n := 0; n < 200 && sc.Scan(); n++ {
		var row struct {
			Cwd string `json:"cwd"`
		}
		if json.Unmarshal(sc.Bytes(), &row) == nil && row.Cwd != "" {
			return row.Cwd
		}
	}
	return ""
}
