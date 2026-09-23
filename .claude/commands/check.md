---
description: Run the tuikit gate (make check) or a single package/test, and summarize.
argument-hint: "[package or -run pattern, e.g. ./chrome/... or -run TestRouter ./viewfsm]"
allowed-tools: Bash(make:*), Bash(go test:*), Bash(go vet:*), Bash(go build:*), Bash(gofmt:*), Read, Edit, Glob, Grep
---

Run this repo's gate.

- If `$ARGUMENTS` is empty: run `make check` — `go build ./...`, `go vet ./...`,
  then `go test -race ./...`. It is seconds on a library this size.
- If `$ARGUMENTS` is provided: run a focused check instead, e.g.
  `go test -race $ARGUMENTS` (so `/check ./chrome/...` or
  `/check -run TestRouter ./viewfsm` work). Prefer this while iterating.

**Never run `golangci-lint`** — not directly, not through `pre-commit run`, not
"just once to check". It runs as the pre-commit hook when you commit and in CI
as the merge gate; a hand-started run reports nothing the hook will not report
moments later, and it costs gigabytes of RSS at `NumCPU` concurrency. When the
hook fails, fix what it reports and commit again.

Then:
1. If `make fmt` reformatted files, note which, confirm the diffs are cosmetic,
   and re-stage them with `git add`.
2. Remember `chrome.Frame` and `theme.Theme` are intentionally `//nolint:govet`
   for fieldalignment — never reorder their fields.
3. Report a tight pass/fail summary: the exact failing check, file:line, and the
   minimal fix. Do not edit beyond making the gate green.
