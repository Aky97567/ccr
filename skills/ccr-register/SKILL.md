---
name: ccr-register
description: Star the current Claude Code session in ccr so Akshay can resume it later by name. Use when he says "star this session", "register this session", "save this session to ccr", "add this to my resume list", or anything equivalent.
---

# Register this session with ccr

Run one command:

```
ccr star --name "<kebab-case-name>" --note "<one-line status>"
```

- **name**: a short, kebab-case slug that identifies what this session is about
  (e.g. `crosslex-backend-build`, `cv-job-search-tailoring`). Derive it from the work
  in this conversation. Keep it under ~40 characters.
- **note**: one sentence on where things stand and what is next. No line breaks.

`ccr` reads the session id from `$CLAUDE_CODE_SESSION_ID` and the working directory
from the current shell, so nothing else needs to be passed. It writes one line to
`~/.local/share/ccr/sessions.jsonl` (or `$CCR_REGISTRY`), replacing any previous entry
for this same session.

After it runs, show Akshay the JSON line it printed, then carry on with the task.

If `ccr` is not installed, tell him to run
`brew install aky97567/ccr/ccr` (or point him at the ccr repo), and fall back to the
manual instructions in the ccr repo's `docs/manual-register.md`.
