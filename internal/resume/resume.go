// Package resume launches `claude --resume` for a stored session.
//
// This is the only package that encodes the current constraint that
// `claude --resume <id>` only searches ~/.claude/projects/<encoded-$PWD>/, so
// the session's directory must be entered first. If Anthropic decouples resume
// from the working directory (claude-code issues #58591 / #74953), only this
// package changes.
package resume

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// ShellCommand renders a copy-pasteable, shell-safe resume command.
func ShellCommand(cwd, id string) string {
	return fmt.Sprintf("cd %s && claude --resume %s", shellQuote(cwd), shellQuote(id))
}

// Exec replaces the current process with `claude --resume <id>`, run from cwd.
// It only returns if the exec fails.
func Exec(cwd, id string) error {
	if fi, err := os.Stat(cwd); err != nil {
		return fmt.Errorf("project directory %q: %w", cwd, err)
	} else if !fi.IsDir() {
		return fmt.Errorf("project path %q is not a directory", cwd)
	}
	bin, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("`claude` not found on PATH: %w", err)
	}
	if err := os.Chdir(cwd); err != nil {
		return err
	}
	return syscall.Exec(bin, []string{bin, "--resume", id}, os.Environ())
}

// shellQuote wraps s in single quotes, escaping any embedded single quotes.
func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
