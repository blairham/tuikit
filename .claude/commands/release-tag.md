---
description: Cut a v0.0.x release tag (tuikit releases by git tag, not goreleaser).
argument-hint: "<version, e.g. v0.0.12>"
allowed-tools: Bash(make:*), Bash(go test:*), Bash(go vet:*), Bash(go build:*), Read, Edit, Glob, Grep
---

Cut a tuikit release. This repo has **no goreleaser** — a consumer picks up the
new version via a `go get github.com/blairham/tuikit@$1` tag bump, so the
release IS the git tag. Target version: `$1` (must be `v0.0.x`, pre-stable).

Steps:
1. Confirm the working tree is clean and we're on `main` (or the branch the user
   intends). If dirty, stop and report.
2. Run `make check` — the tag must point at a green commit. If it fails, stop.
3. Verify `$1` is a valid, monotonically-increasing `v0.0.x` tag — check
   `git tag --list 'v0.0.*'` and `git describe --tags --abbrev=0` so we don't
   reuse or skip backwards.
4. Update `CHANGELOG.md`: move the Unreleased entries under a new `## $1` heading
   dated today; leave a fresh Unreleased section.
5. STOP and show the user the proposed CHANGELOG diff and the exact commands
   (`git commit`, `git tag -a $1 -m ...`, `git push && git push origin $1`)
   before running anything that mutates the remote. Do not push tags without
   explicit confirmation.
