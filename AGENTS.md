# AGENTS.md

Guidance for AI coding agents (Claude Code, Cursor, Copilot, Codex, OpenCode, …) working in this repository. This is the **cross-tool single source of truth** — `CLAUDE.md` imports it.

## Project Overview

A k9s-style TUI chrome toolkit for [Bubble Tea v2](https://github.com/charmbracelet/bubbletea) apps. Status is **pre-stable `v0.0.x`** — the API is intentionally fluid while the applications that use it migrate onto it. `v0.1.0` ships once the surface settles. Expect breaking changes; favor small additive PRs.

## Releases and consumers

**No goreleaser** — a release *is* an annotated git tag, and consumers pick it up with `go get github.com/blairham/tuikit@vX.Y.Z`. Newest tag: **v0.0.0**. Read [`.claude/commands/release-tag.md`](.claude/commands/release-tag.md) before cutting `v0.0.1`; it is the procedure (clean tree, green `make check`, monotonic tag check, CHANGELOG move out of `[Unreleased]`, confirm-before-push).

Assume any given consumer may be several tags behind when reasoning about whether a fix has reached it — being on `main` here is not evidence that it has.

## Quick Reference

The Makefile is the entry point. Use `make help` to list targets. Common flows:

```bash
make check        # build + vet + race tests — this is what CI runs
make build        # go build ./...
make test         # go test -race ./...
make vet          # go vet ./...
make fmt          # gofumpt + fieldalignment
make tidy         # go mod tidy
make sync         # rewrite .tool-versions to match go.mod's Go version

# Single-package or single-test runs (no Make target — invoke go directly):
go test -race ./chrome/...
go test -race -run TestRouter ./viewfsm

# Run the example app:
go run ./examples/commandbar
```

Go-based tooling (`gofumpt`, `fieldalignment`, `golangci-lint`) is declared in `go.mod`'s `tool` block and invoked via `go tool <name>` — do **not** add it to `.tool-versions`. The Go toolchain itself is pinned in `.tool-versions` (asdf-managed).

**There is no `lint` target, and you never run `golangci-lint` by hand.** It runs as a pre-commit hook at commit time and in CI as the merge gate. A hand-started run reports nothing the hook would not report moments later, and it defaults its concurrency to `NumCPU` at gigabytes of RSS — `run.concurrency` in `.golangci.yml` is the lever that bounds it. When the hook reports a failure, fix it and commit again; that is the only lint output worth reading.

## Project Structure

Six composable packages under one module, all rooted on `theme`. Apps import whichever subset they need; nothing here owns the `tea.Program` loop.

| Package | Role |
|---|---|
| `theme/` | Color palette + lipgloss styles + the `Theme` struct apps inject at construction (the DI seed every other package reads from). |
| `chrome/` | The visual frame: top section, bordered content, footer, filter / command / confirm bars, `Modal` (centered popup), `Prompt` (one-shot prefilled text input), the `VersionLine` info row, and the help overlay. |
| `table/` | Themed `bubbles/v2/table` styles, the `FixSelectedRow` ANSI fix, and a `RowFilter`. |
| `tail/` | Viewport + follow-mode + `RowFilter` for log streams. |
| `loading/` | Spinner + rotating-tip loading / transition screen. |
| `viewfsm/` | The `Router` over `ViewID` constants: drill stack, digit hotkeys, breadcrumbs, plus the `TranslateNavKey` helper. **No `View` interface** — apps own their view models. |

`examples/` holds runnable demos (e.g. `examples/commandbar`). See [`docs/design/`](docs/design/) for per-subsystem design notes.

## Architecture

Six composable packages under one module. Apps import whichever subset they need; nothing here owns the `tea.Program` loop.

```
theme  ── color palette + lipgloss styles + Theme struct (DI seed)
   │
   ├──► chrome    (visual frame: top section, bordered content, footer, bars, help overlay)
   ├──► table     (themed bubbles/v2/table styles, FixSelectedRow ANSI fix, RowFilter)
   ├──► tail      (viewport + follow-mode + RowFilter for log streams)
   ├──► loading   (spinner + rotating-tip loading/transition screen)
   └──► viewfsm   (Router over ViewID consts: drill stack, digit hotkeys, breadcrumbs)
```

`tail.Model` feeds via `AppendLine`/`AppendLines` (live, at bottom) and `PrependLines` (older history, at top, preserving scroll position); `AtTop` is the cue to fetch more history. `SetBackground` re-asserts the theme background after the reset codes in styled log lines so the content stays continuous (lipgloss does not restore a surrounding background after an inner reset).

`loading.Model` is a spinner beside a rotating one-line tip for loading screens and between-views transitions: `Tick` to start, forward each `TickMsg` to `Update` while loading, render `View`/`Centered`.

### Data flow per tick

Apps drive the loop. Each `View()` call:

1. App builds a `chrome.Frame{}` for this tick: `InfoLines`, `Shortcuts`, rendered `Content`, `Breadcrumb`, optional `Filter`/`Command` textinputs, optional `StatusBar`/`StatusBarLevel`, `HelpVisible`.
2. App calls `chrome.Chrome.ContentInnerSize(w, h, filtering, commanding, confirming, statusBar)` **before** rendering `Content` to know the inner table/viewport dimensions — note the **six** parameters: `confirming` sits between `commanding` and `statusBar` (it was added in 0.0.0 as a breaking change, so a five-argument call is a stale one). The top-section reservation is derived from `Chrome.TopSectionRows()`, not hardcoded, so logo/info-panel/shortcut-row counts compose correctly. Reservations: filter / command / confirm bars are **3 rows each** when active, the status bar is **1 row**, the footer is 2. `renderStatusBar` renders `Height(1)` to match — the 0.0.0 CHANGELOG's "3 rows to match the 3-row reservation" entry describes the shape at that time, not today's.
3. App wraps content via `Chrome.BorderedContent(...)` then `chrome.InjectBorderTitle(...)` for the centered title pill.
4. App returns `Chrome.Render(frame)`.

`chrome.Frame` is a per-tick assembly contract — its `//nolint:govet` for fieldalignment is intentional (readability of a public API struct wins over alignment). Same convention applies to `theme.Theme`. Do not reorder these structs to satisfy `fieldalignment`.

### Theme is the dependency-injection seed

`theme.Theme` is constructed once by the app and passed into `chrome.Config`. Every other package reads colors and pre-built styles from it — there are no package-level globals. To recolor the chrome, apps build a different `Theme`; nothing else changes.

`Theme.PaintBackground` is the load-bearing knob:

- **`true`** — every styled span explicitly paints `Bg`. Required on terminals (notably macOS Terminal) where the default background bleeds through unstyled gaps. `Default()` returns this.
- **`false`** — only the outer screen wrapper paints `Bg`. Lets `bubbles/v2/table.Styles.Selected.Foreground` win without a background fight.

Both modes are first-class. When adding any new styled output, route it through `Theme.On(...)` (or one of the pre-built styles) so it honors `PaintBackground` automatically. Don't apply `Background(...)` directly without gating on `Theme.PaintBackground`.

### viewfsm.Router does not own views

The router tracks **which view ID is active**, the drill stack, and digit hotkeys (`1..N`). Apps keep their concrete view models in their own state and consult the router for "what's on top?" and "what's the breadcrumb chain?". Routes are plain `ViewID int` constants — typedef in your app, register `Spec{Name, Hotkey}` per ID, then translate the router's active ID to your concrete view in `Update`/`View`.

**There is no `View` interface**, despite what the `viewfsm` package comment says — the exported surface is `NewRouter`, `Active`, `Stack`, `IsAtRoot`, `Push`, `Pop`, `Replace`, `JumpTo`, `ResolveHotkey`, `Spec`, `Breadcrumb`, and the free function `TranslateNavKey`. Nor does the router dispatch global keys: `TranslateNavKey` handles **only** `j`/`k`/`g`/`G`/`ctrl+d`/`ctrl+u`, rewriting them into the arrow/page keys `bubbles` tables understand. `/`, `:`, `?`, `esc`, `q`, `enter` and `r` are the app's own key path (digit hotkeys reach the router via `ResolveHotkey`).

The router supports both fixed shallow drill stacks and deep stacks with detail/form views — both shapes must keep working.

### chrome.CommandBar / filter bar lifecycle

`chrome.CommandBar` owns the `:` palette: `Active()`, `Update(msg, dispatch)`, render. The app forwards key messages to it when `Active()` and uses its returned `handled` bool to decide whether to fall through to other keys. See `examples/commandbar/main.go` for the canonical pattern.

## Code Conventions

- **Dependency tree is intentionally tight** — Charm (`bubbletea/v2`, `bubbles/v2`, `lipgloss/v2`) plus `image/color`. Do not add new deps without an issue.
- **goimports local-prefix is `github.com/blairham/tuikit`** (see `.golangci.yml` / `gci` settings) — internal imports go in their own group.
- **`golines` max-len 120, tab-len 1** — long lines get wrapped by `make fmt`.
- **misspell locale: US** — en-US spelling everywhere (already required by parent `CLAUDE.md`).
- **`v0.0.x` API rule** — no major surface additions without a prior issue; domain-specific features stay in the consuming applications, not here.
- **Examples must lint** — `examples/` is excluded only from `funlen`/`gocyclo`/`gocognit`; everything else still applies.

## Documentation

See [`docs/README.md`](docs/README.md) for the docs index. In-depth design notes live in [`docs/design/`](docs/design/) — one living document per subsystem, each with a `Status:` and `Code:` header pointing at the implementation.

Read the relevant doc before changing the area it covers — they capture trade-offs, prior incidents, and constraints that aren't visible from the source. When a change makes a doc stale, update it in the same commit; don't defer.
