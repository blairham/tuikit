# Contributing to tuikit

Thanks for looking. Issues, bug reports and pull requests are all welcome.

tuikit is pre-stable `v0.0.x`: the API is still being shaped by the
applications migrating onto it, and `v0.1.0` tags once it settles. Breaking
changes can land in any release until then, so small, additive PRs are the
easiest to take.

## Before you open a PR

```sh
pre-commit install   # once per clone: formatting, golangci-lint, secrets, YAML, license headers
make test            # go test -race ./...
```

golangci-lint runs as the commit hook and in CI, not as a make target; a
failing hook fails the commit, and you fix it and commit again. `make fmt`
applies gofumpt and fieldalignment.

## What we want

- **Bug reports** with a minimal repro. The surface is small; reproductions
  should be tractable.
- **Cross-terminal screenshots** (macOS Terminal, iTerm2, Alacritty, kitty,
  WezTerm, Ghostty). `PaintBackground` in `theme` is load-bearing across
  terminals, and both modes have to stay correct.
- **Themes** — a new constructor in `theme/`.
- **Examples** in `examples/` showing one pattern each.

## What we don't want (yet)

- Large API additions while `v0.0.x` is unstable. Open an issue first.
- Features only one app needs — those belong in that app.
- New dependencies. Beyond the Charm stack there are two (go-colorful and
  yaml, for skins); a third needs a reason.

## The shape of a change

[`AGENTS.md`](AGENTS.md) is the architecture guide. The rules that most often
decide a review:

- **tuikit never owns the `tea.Program` loop.** Packages hand back models,
  styles and strings; the app drives them.
- **Every package reads its colors from `theme.Theme`**, which the app
  injects at construction.
- Struct literals use **named fields**; the linter's `fieldalignment` fix
  reorders fields.
- Every exported change gets a line under `[Unreleased]` in
  [`CHANGELOG.md`](CHANGELOG.md).

## Tests

**New functionality comes with tests in the same pull request, and a bug fix
comes with a test that fails without the fix.** A PR that adds an exported
function, option or key binding without a test that exercises it is not
ready to merge.

- Render and assert. Strip ANSI when the text is the point; assert on the
  raw sequences when the bytes are.
- Tests must never touch real user state — use `t.TempDir()` and
  `t.Setenv`.
- Code that takes untrusted text has a fuzz target (`go test -fuzz`) that
  checks a property, not just the absence of a panic.

## The Contributor License Agreement

Contributions require a signed CLA; the text is in [`CLA.md`](CLA.md).

**Why.** The project may need to offer different licensing terms in future.
That is only possible if one party can license the whole work, and copyright
in a contribution stays with its author unless licensed onward.

The CLA does **not** take your copyright. You keep it; you grant a license
broad enough to include sublicensing, and you affirm the work is your own —
including that no employer holds rights to it.

## Commits and PRs

- Prefix the subject with the package it touches (`table:`, `tail:`,
  `chrome:`, `docs:`, `ci:`).
- Explain the **why** in the commit message. The diff already says what.
- One change per PR, and put `Closes #N` in the PR body.
- Commits must be signed.
- Every `.go` file carries the two-line SPDX header (`Apache-2.0`); the
  pre-commit hook fails without it.

## Releasing

Maintainers only. A release is a signed, annotated `vX.Y.Z` tag on `main`,
on the commit that moves the CHANGELOG's `[Unreleased]` section under
`## [X.Y.Z] - <date>`. Pushing the tag runs the release workflow, which
publishes that section as the notes, the source archive, a cosign-signed
`checksums.txt` and build provenance. Nothing is compiled.
