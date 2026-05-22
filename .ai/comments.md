# Comments

Comment code as you write it, in the same edit — commenting is part of writing code.

## Tags

* Four tags, lowercase: `// todo:` `// fixme:` `// note:` `// why:`. Find with `rg -i "// (todo|fixme|note|why):"`.
* `// todo:` — deferred or low-priority work; states what is missing.
* `// fixme:` — a shipped gap to fix ASAP; states which invariant, edge case, or validation is skipped.
* `// note:` — a clarification or summary; states what the code does or what to watch for.
* `// why:` — the reason for a non-obvious decision; states why this, not the alternative.

```go
// todo: reconnect on recoverable errors
// fixme: no zone check — network calls can leak; caller must add Network
// note: this connection only keeps the query handler alive; serves no data
// why: text interface omitted to force base64 encoding in channels
```

## MUST

* Comment intent, not mechanics.
* Write each comment as a terse, declarative statement — one fact per line.
* Tag a non-obvious decision with `// why:`, never bury it in prose.
* Keep tags in sync with their code; remove a `todo`/`fixme` when resolved.
* Link to `.ai/...` for context held elsewhere.

## MUST NOT

* Comment, reformat, or re-tag code outside the current change.
* Restate the code (`i++ // increment i`) or add an empty tag (`// note: helper function`).
* Duplicate `.ai/` context inline.
