# ccr

Star Claude Code sessions and resume them by name.

`claude --resume <id>` only works from the directory the session started in, and
there is no way to bookmark or label a session. `ccr` keeps a small machine-local
list of sessions you care about — a name, a note, the working directory, the id —
and resumes the one you pick.

## Install

Homebrew:

```
brew install aky97567/ccr/ccr
```

Or grab a binary from the [releases](https://github.com/Aky97567/ccr/releases) and
drop it on your `PATH`, then:

```
ccr install-skill      # adds the /ccr-register skill for Claude Code
```

## Use

Inside a Claude Code session, star it:

```
ccr star --name crosslex-backend-build --note "phase 5 done; debugging the compound-word regex"
```

or just say "star this session" and let the `/ccr-register` skill run that for you.

Later, from any terminal:

```
ccr                    # pick a session -> resumes it in its own directory
ccr list               # show everything you've starred
ccr rm <name>          # drop one (writes a .bak first)
```

`ccr` uses [`fzf`](https://github.com/junegunn/fzf) for the picker when it is
installed, and a numbered prompt otherwise.

## Registry

One JSON object per line at `~/.local/share/ccr/sessions.jsonl`
(override with `$CCR_REGISTRY`):

```json
{"name":"crosslex-backend-build","session_id":"2f11dc41-…","cwd":"/Users/akshay/Code/personal/crosslex","note":"…","updated":"2026-09-06T15:58:58+0200"}
```

It is machine-local by design and never committed to a repo. Every write is a
temp-file-plus-rename; `rm` takes a timestamped backup first.

Without the skill, `docs/manual-register.md` has a prompt you can paste into a
session instead.

## How resume works

`claude --resume <id>` is currently filtered to
`~/.claude/projects/<encoded-$PWD>/`, so `ccr` enters the stored `cwd` before
exec-ing `claude`. If [claude-code#58591](https://github.com/anthropics/claude-code/issues/58591)
/ [#74953](https://github.com/anthropics/claude-code/issues/74953) land, only
`internal/resume` needs to change.
