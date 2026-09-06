// Command ccr stars Claude Code sessions and resumes them by name.
package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/Aky97567/ccr/internal/registry"
	"github.com/Aky97567/ccr/internal/resume"
	"github.com/Aky97567/ccr/internal/session"
	"github.com/Aky97567/ccr/internal/ui"
)

// version is overridden at release time via -ldflags "-X main.version=...".
var version = "dev"

//go:embed skills/ccr-register/SKILL.md
var skillMD string

const usage = `ccr — star Claude Code sessions and resume them by name

usage:
  ccr                     pick a starred session and resume it
  ccr star [flags]        star (or update) the current session
  ccr list                list starred sessions
  ccr rm <name>           remove a starred session
  ccr install-skill       install the /ccr-register Claude Code skill
  ccr version             print version

ccr star flags:
  --name   short name for the session (default: current directory name)
  --note   one-line status note
  --session-id  override session id (default: $CLAUDE_CODE_SESSION_ID or cwd match)

registry: $CCR_REGISTRY, or ~/.local/share/ccr/sessions.jsonl
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "ccr: "+err.Error())
		os.Exit(1)
	}
}

func run(args []string) error {
	cmd := ""
	if len(args) > 0 {
		cmd = args[0]
	}
	switch cmd {
	case "", "select":
		return cmdSelect()
	case "-h", "--help", "help":
		fmt.Print(usage)
		return nil
	case "version", "-v", "--version":
		fmt.Println("ccr " + version)
		return nil
	case "star":
		return cmdStar(args[1:])
	case "list", "ls":
		return cmdList()
	case "rm", "remove":
		return cmdRm(args[1:])
	case "install-skill":
		return cmdInstallSkill()
	}
	if strings.HasPrefix(cmd, "-") {
		fmt.Print(usage)
		return nil
	}
	return fmt.Errorf("unknown command %q (try `ccr --help`)", cmd)
}

func cmdSelect() error {
	path, err := registry.Path()
	if err != nil {
		return err
	}
	entries, err := registry.Load(path)
	if err != nil {
		return err
	}
	e, err := ui.Select(entries)
	if err != nil {
		if err == ui.ErrAborted {
			return nil
		}
		return err
	}
	fmt.Fprintln(os.Stderr, "→ "+resume.ShellCommand(e.Cwd, e.SessionID))
	return resume.Exec(e.Cwd, e.SessionID)
}

func cmdStar(args []string) error {
	fs := flag.NewFlagSet("star", flag.ContinueOnError)
	name := fs.String("name", "", "short name for the session")
	note := fs.String("note", "", "one-line status note")
	sid := fs.String("session-id", "", "override session id")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	id := strings.TrimSpace(*sid)
	if id == "" {
		if id, err = session.CurrentID(cwd); err != nil {
			return err
		}
	}
	n := strings.TrimSpace(*name)
	if n == "" {
		n = filepath.Base(cwd)
	}

	path, err := registry.Path()
	if err != nil {
		return err
	}
	entries, err := registry.Load(path)
	if err != nil {
		return err
	}
	e := registry.Entry{
		Name:      n,
		SessionID: id,
		Cwd:       cwd,
		Note:      strings.TrimSpace(*note),
		Updated:   registry.Now(),
	}
	entries = registry.Upsert(entries, e)
	if err := registry.Save(path, entries); err != nil {
		return err
	}
	out, _ := json.Marshal(e)
	fmt.Println(string(out))
	return nil
}

func cmdList() error {
	path, err := registry.Path()
	if err != nil {
		return err
	}
	entries, err := registry.Load(path)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "no starred sessions yet — run `ccr star` inside a Claude Code session")
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tCWD\tID\tNOTE")
	for _, e := range entries {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.Name, e.Cwd, short(e.SessionID), truncate(e.Note, 60))
	}
	return w.Flush()
}

func cmdRm(args []string) error {
	if len(args) != 1 || args[0] == "" {
		return fmt.Errorf("usage: ccr rm <name>")
	}
	target := args[0]
	path, err := registry.Path()
	if err != nil {
		return err
	}
	entries, err := registry.Load(path)
	if err != nil {
		return err
	}
	kept := make([]registry.Entry, 0, len(entries))
	removed := 0
	for _, e := range entries {
		if e.Name == target {
			removed++
			continue
		}
		kept = append(kept, e)
	}
	if removed == 0 {
		return fmt.Errorf("no starred session named %q", target)
	}
	bak, err := registry.Backup(path)
	if err != nil {
		return err
	}
	if err := registry.Save(path, kept); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "removed %d entr%s named %q (backup: %s)\n",
		removed, plural(removed), target, bak)
	return nil
}

func cmdInstallSkill() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".claude", "skills", "ccr-register")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	dst := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(dst, []byte(skillMD), 0o644); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "installed skill: "+dst)
	return nil
}

func short(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func plural(n int) string {
	if n == 1 {
		return "y"
	}
	return "ies"
}
