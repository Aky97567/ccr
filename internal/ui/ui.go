// Package ui presents the starred sessions and returns the chosen one.
package ui

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Aky97567/ccr/internal/registry"
)

// ErrAborted is returned when the user dismisses the selector without choosing.
var ErrAborted = fmt.Errorf("aborted")

// Select shows the entries and returns the chosen one. It uses fzf when it is on
// PATH and the session is interactive; otherwise a numbered prompt.
func Select(entries []registry.Entry) (registry.Entry, error) {
	if len(entries) == 0 {
		return registry.Entry{}, fmt.Errorf("no starred sessions yet — run `ccr star` inside a Claude Code session first")
	}
	lines := make([]string, len(entries))
	for i, e := range entries {
		name := e.Name
		if name == "" {
			name = "(unnamed)"
		}
		lines[i] = fmt.Sprintf("%-28s  %s", name, prettyDir(e.Cwd))
	}

	if _, err := exec.LookPath("fzf"); err == nil && isTTY(os.Stdin) && isTTY(os.Stdout) {
		return selectFzf(entries, lines)
	}
	return selectNumbered(entries, lines)
}

func selectFzf(entries []registry.Entry, lines []string) (registry.Entry, error) {
	indexed := make([]string, len(lines))
	for i, l := range lines {
		indexed[i] = fmt.Sprintf("%d\t%s", i, l)
	}
	cmd := exec.Command("fzf",
		"--with-nth=2..", "--delimiter=\t",
		"--no-sort", "--reverse", "--height=40%",
		"--prompt=resume ▸ ", "--header=pick a session to resume")
	cmd.Stdin = strings.NewReader(strings.Join(indexed, "\n"))
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return registry.Entry{}, ErrAborted
	}
	sel := strings.TrimSpace(string(out))
	if sel == "" {
		return registry.Entry{}, ErrAborted
	}
	idx, err := strconv.Atoi(strings.SplitN(sel, "\t", 2)[0])
	if err != nil || idx < 0 || idx >= len(entries) {
		return registry.Entry{}, fmt.Errorf("unexpected fzf output: %q", sel)
	}
	return entries[idx], nil
}

func selectNumbered(entries []registry.Entry, lines []string) (registry.Entry, error) {
	for i, l := range lines {
		fmt.Fprintf(os.Stderr, "%3d  %s\n", i+1, l)
	}
	fmt.Fprintf(os.Stderr, "\nResume which? [1-%d, q to quit]: ", len(entries))

	sc := bufio.NewScanner(os.Stdin)
	if !sc.Scan() {
		return registry.Entry{}, ErrAborted
	}
	choice := strings.TrimSpace(sc.Text())
	if choice == "" || choice == "q" || choice == "Q" {
		return registry.Entry{}, ErrAborted
	}
	n, err := strconv.Atoi(choice)
	if err != nil || n < 1 || n > len(entries) {
		return registry.Entry{}, fmt.Errorf("invalid choice: %q", choice)
	}
	return entries[n-1], nil
}

// prettyDir shortens the user's home directory to ~.
func prettyDir(dir string) string {
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		if dir == home {
			return "~"
		}
		if strings.HasPrefix(dir, home+string(filepath.Separator)) {
			return "~" + dir[len(home):]
		}
	}
	return dir
}

func isTTY(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}
