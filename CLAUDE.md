# CLAUDE.md

@AGENTS.md

<!-- AGENTS.md (imported above) is the cross-tool single source of truth, read by every AI coding tool. Put durable project context THERE. This file holds only Claude Code-specific extras. -->

## Claude Code-specific notes
- **Subagents** (`.claude/agents/`):
  - `tuikit-component-author` — authoring/extending a chrome/table/tail/loading widget the theme-DI + PaintBackground way (no globals, no raw `Background(...)`).
  - `viewfsm-router-helper` — wiring a consumer app onto `viewfsm.Router`: ViewID enums, `Spec` registration, drill/pop, digit hotkeys, breadcrumb chain.
  - `tuikit-check-runner` — runs `make check` (build + vet + race tests), interprets failures, and re-stages reformatted files. Never runs `golangci-lint` by hand — that is the pre-commit hook's job.
- **Slash commands** (`.claude/commands/`):
  - `/check` — run the `make check` gate (or a single package/test via args) and summarize results.
  - `/release-tag` — cut a `v0.0.x` git tag (this repo releases by tag, not goreleaser): verify clean tree + green check, bump CHANGELOG, tag, push.
