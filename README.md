# tuikit

A k9s-style TUI chrome toolkit for [Bubble Tea](https://github.com/charmbracelet/bubbletea) apps. Build interactive terminal UIs that look and behave like `k9s` — info panel, shortcut grid, bordered content with injected titles, breadcrumb footer, live filter and command modes, help overlay — without reimplementing the chrome every time.

> **Status:** v0.0.x — pre-stable. The API is still being shaped by the applications migrating onto it. Expect breaking changes through the v0.x line; `v0.1.0` tags once the surface settles and the API stabilizes.

## What's in the box

Six composable packages, each usable in isolation:

- **`theme`** — palette + lipgloss style helpers + a `Theme` struct apps inject at construction. Swap themes; the chrome reads from `Theme`.
- **`chrome`** — the visual frame: top section (info panel + shortcut grid + ASCII logo slot), bordered content with `InjectBorderTitle`, breadcrumb footer, status bar, filter / command / confirm bars, `Modal` (centered popup), `Confirm` (inline y/n), `Prompt` (one-shot prefilled text input), the `VersionLine` info row, and the help overlay.
- **`table`** — themed styles + keymap for `bubbles/v2/table`, the `FixSelectedRow` ANSI post-processor that solves the table selection-paint bug, and a `RowFilter` (regex + `!` negation + literal fallback).
- **`viewfsm`** — a `Router` over `ViewID` constants: `Active`/`Stack`/`IsAtRoot`, `Push`/`Pop`/`Replace`/`JumpTo`, `ResolveHotkey` (digit hotkeys), `Spec`, and `Breadcrumb` emission. **There is no `View` interface** — apps keep their own concrete view models and consult the router for what's on top. The only key helper is `TranslateNavKey`, which rewrites `j`/`k`/`g`/`G`/`ctrl+d`/`ctrl+u` into the arrow/page keys `bubbles` tables understand; the app owns `/ : ? esc q enter r` dispatch itself.
- **`tail`** — a viewport + follow-mode + live-filter wrapper for log / event streams. `AppendLine`/`AppendLines` add live lines at the bottom (the batch form repaints once instead of per line); `PrependLines` inserts older history at the top while preserving the user's scroll position, with `AtTop` as the cue to fetch more; `SetBackground` re-asserts the theme background after the reset codes inside styled log lines, so the painted area stays continuous.
- **`loading`** — a spinner beside a rotating one-line tip for initial loading screens and between-view transitions. `Tick` to start, forward each `TickMsg` to `Update`, render `View` (single line) or `Centered(w, h)`.

## Design principles

- **Toolkit, not framework.** Apps own their `tea.Program` and assemble what they need. Like `bubbles`, not like a base class.
- **Theme-first.** Every visible style flows from `theme.Theme`. Build your own theme; reuse everyone's chrome.
- **No background assumptions.** A `Theme.PaintBackground` toggle covers both the "paint everything black" and "let the terminal background show" terminal aesthetics.
- **Apps own data + views.** What the views are, what columns they have, how they fetch data — all app-specific. Tuikit owns the visual scaffold.

## Quickstart

```go
import (
    tea "charm.land/bubbletea/v2"
    "github.com/blairham/tuikit/chrome"
    "github.com/blairham/tuikit/loading"
    "github.com/blairham/tuikit/theme"
    "github.com/blairham/tuikit/viewfsm"
)

const (
    vList viewfsm.ViewID = iota
    vDetail
)

func main() {
    t := theme.Default()
    app := myApp{
        theme:  t,
        chrome: chrome.New(chrome.Config{Theme: t, Logo: myLogoLines}),
        router: viewfsm.NewRouter(map[viewfsm.ViewID]viewfsm.Spec{
            vList:   {Name: "List", Hotkey: "1"},
            vDetail: {Name: "Detail", Hotkey: "2"},
        }, vList),
        loading: loading.New(t, myTips),
    }
    if _, err := tea.NewProgram(app).Run(); err != nil {
        log.Fatal(err)
    }
}
```

Size the content area **before** rendering it, so the chrome's reservations and
your table/viewport dimensions agree:

```go
innerW, innerH := app.chrome.ContentInnerSize(
    w, h,
    app.filter.Active(), app.command.Active(), app.confirm.Active(), app.statusBar != "",
)
```

See `examples/` for a runnable demo.

## Sizing note: the status bar is ONE row

`ContentInnerSize` reserves a single row for the status bar, and
`renderStatusBar` renders exactly `Height(1)` to fill it. The 0.0.0 CHANGELOG
has a "Fixed" entry describing a 3-row status bar matched to a 3-row
reservation — that was the shape at the time and has since changed. Reserve one
row. (The filter, command, and confirm bars are 3 rows each when active; the
footer is 2.)

## Releases

This repo has **no goreleaser** — a release *is* a git tag, and consumers pick
it up with `go get github.com/blairham/tuikit@vX.Y.Z`. Newest tag: **v0.0.0**.
Read [`.claude/commands/release-tag.md`](.claude/commands/release-tag.md) for
the procedure before cutting `v0.0.1` — it covers the clean-tree and green-check
gates, the CHANGELOG move out of `[Unreleased]`, and the confirm-before-push
rule.

Consuming applications may be several tags behind at any given moment, so do
not assume a fix has reached one just because it is on `main` here.

## Roadmap

- **v0.0.x** — extraction of shared chrome out of the applications that grew it. API in flux. See the CHANGELOG for the surface shipped in `v0.0.0`.
- **v0.1** — first batch of opt-in extras: row tick-flash, fuzzy-match command suggestions, splash screen.
- **v0.2** — community features. PR away.

## License

Apache-2.0. See [LICENSE](./LICENSE).
