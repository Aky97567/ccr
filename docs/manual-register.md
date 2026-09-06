# Registering a session without `ccr`

If `ccr` is installed, ignore this file — run `ccr star --name <name> --note "<note>"`
from inside the session, or trigger the `/ccr-register` skill.

This is the fallback for a machine where `ccr` isn't available yet. Paste it into
the running Claude Code session you want to save:

---

Add this session to my ccr resume list.

1. Session id: run `echo "$CLAUDE_CODE_SESSION_ID"`. If that's empty (Claude Code
   older than 2.1.132), find the transcript instead — the most recently modified
   file under `~/.claude/projects/` whose recorded `cwd` matches `pwd`:
   `grep -rlm1 "\"cwd\":\"$(pwd)\"" ~/.claude/projects/*/ | xargs ls -t | head -1`
   and take the basename without `.jsonl`.
2. Working directory: `pwd`.
3. Choose a short kebab-case `name` for this session and a one-sentence `note` on
   where things stand.
4. Ensure the file exists:
   `mkdir -p ~/.local/share/ccr && touch ~/.local/share/ccr/sessions.jsonl`
5. If a line with this `session_id` already exists there, delete that old line.
6. Append exactly one compact JSON line:
   `{"name":"<name>","session_id":"<id>","cwd":"<pwd>","note":"<note>","updated":"<`date +%Y-%m-%dT%H:%M:%S%z`>"}`
7. Print the line you added.

Then carry on with what we were doing.
