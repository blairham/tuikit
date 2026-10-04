# tuikit

[![CI](https://github.com/blairham/tuikit/actions/workflows/ci.yml/badge.svg)](https://github.com/blairham/tuikit/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/tag/blairham/tuikit?sort=semver&label=release)](https://github.com/blairham/tuikit/tags)
[![CodeQL](https://github.com/blairham/tuikit/actions/workflows/codeql.yml/badge.svg)](https://github.com/blairham/tuikit/actions/workflows/codeql.yml)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/blairham/tuikit/badge)](https://scorecard.dev/viewer/?uri=github.com/blairham/tuikit)
[![Go version](https://img.shields.io/github/go-mod/go-version/blairham/tuikit)](go.mod)
[![Go Reference](https://pkg.go.dev/badge/github.com/blairham/tuikit.svg)](https://pkg.go.dev/github.com/blairham/tuikit)
[![License: Apache-2.0](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)

A k9s-style TUI chrome toolkit for [Bubble Tea](https://github.com/charmbracelet/bubbletea) apps. Build interactive terminal UIs that look and behave like `k9s` — info panel, shortcut grid, bordered content with injected titles, breadcrumb footer, live filter and command modes, help overlay — without reimplementing the chrome every time.

> **Status:** v0.0.x — pre-stable. The API is still being shaped by the applications migrating onto it. Expect breaking changes through the v0.x line; `v0.1.0` tags once the surface settles and the API stabilizes.

## What's in the box

Seven composable packages, each usable in isolation:

- **`theme`** — palette + lipgloss style helpers + a `Theme` struct apps inject at construction. Swap themes; the chrome reads from `Theme`.
- **`chrome`** — the visual frame: top section (info panel + shortcut grid + ASCII logo slot), bordered content with `InjectBorderTitle`, breadcrumb footer, status bar, filter / command / confirm bars (the filter and command bars recall earlier entries with ↑/↓, and expose `History`/`SetHistory` so apps can persist them), `Modal` (centered popup), `Confirm` (inline y/n), `Prompt` (one-shot prefilled text input), the `VersionLine` info row, the help overlay, and `SaveDump`, which writes a view to a timestamped plain-text file (ANSI stripped) for k9s-style ctrl+s screen dumps.
- **`table`** — themed styles + keymap for `bubbles/v2/table`, the `FixSelectedRow` ANSI post-processor that solves the table selection-paint bug, a `RowFilter` with k9s's three filter-bar modes — a regex with `!` negation and a literal fallback, `-f term` for a fuzzy (in-order, case-insensitive) match, and `-l app=web,!tier` for a label selector over a row's labels (`Match(fields, labels)`; `Kind()` tells an app a label filter is active), a `Sorter` (shift+←/→ sort column, stable sort, `↑`/`↓` header indicator), and `Marks` (space / ctrl+space / ctrl+\ marks for bulk actions, keyed so they follow their rows through re-sorts and refreshes), and `PlainText` (every row as aligned plain text, for saving).
- **`viewfsm`** — a `Router` over `ViewID` constants: `Active`/`Stack`/`IsAtRoot`, `Push`/`Pop`/`Replace`/`JumpTo`, `ResolveHotkey` (digit hotkeys), `Spec`, and `Breadcrumb` emission. **There is no `View` interface** — apps keep their own concrete view models and consult the router for what's on top. The only key helper is `TranslateNavKey`, which rewrites `j`/`k`/`g`/`G`/`ctrl+d`/`ctrl+u` into the arrow/page keys `bubbles` tables understand; the app owns `/ : ? esc q enter r` dispatch itself.
- **`tail`** — a viewport + follow-mode + live-filter wrapper for log / event streams. `AppendLine`/`AppendLines` add live lines at the bottom (the batch form repaints once instead of per line); `PrependLines` inserts older history at the top while preserving the user's scroll position, with `AtTop` as the cue to fetch more; `SetBackground` re-asserts the theme background after the reset codes inside styled log lines, so the painted area stays continuous. `SetSearch` is k9s's describe-view search beside the filter: it hides nothing, highlights matches inside already-colored lines, and `NextMatch`/`PrevMatch` jump between matching lines (`MatchIndex`/`Matches` give the "3/17").
- **`tree`** — a navigable, collapsible tree in the style of k9s's xray: a resource and everything under it. Hand it `tree.Node{ID, Label, Children}` values with `SetRoots` as often as you poll; expansion state and the cursor are kept by ID, so the user's place survives a refresh, and new nodes open to `SetDefaultDepth`. `HandleKey` moves over the visible rows (↑/↓, pgup/pgdown, home/end and the vim letters), `→`/`l` expands or steps into the first child, `←`/`h` collapses or steps to the parent, and space toggles; enter is left to the app. `SetFilter` takes a `table.ParseFilter` expression and keeps a node when it or any descendant matches, with the path to each match opened. `Selected`/`SelectedPath` read the cursor; `ExpandAll`/`CollapseAll`. Labels may carry color; rows are cut to the width ANSI-aware. See `examples/tree`.
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

To recolor, adjust the color fields and call `Rebuild` before handing the
theme out — the pre-built styles are computed from the colors at construction
and do not follow a later assignment on their own:

```go
t := theme.Default()
t.Logo = lipgloss.Color("#2496ED")
t.Accent = lipgloss.Color("#2496ED")
t.Rebuild()
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
footer is 2, or 0 while `Chrome.CrumbsHidden` is set — see `Chrome.ToggleCrumbs`
and `chrome.KeyToggleCrumbs`, ctrl+g. Likewise the top section's
`TopSectionRows()` drops to 0 while `Chrome.HeaderHidden` is set — see
`Chrome.ToggleHeader` and `chrome.KeyToggleHeader`, ctrl+e.)

## Header layout

The top section is laid out the way k9s lays out its header:

```
 info panel | gap | views | actions | actions | ...      fill      | logo |
```

The info panel shrinks to its widest `InfoLines` row (capped at
`Config.InfoLabelWidth`), and the shortcut columns start two cells after it,
left-aligned, whether or not a logo is shown. The logo is pinned as a block
against the right edge. `ShortcutGrid` wraps views and actions into columns of
`Config.ShortcutRows` entries (6 by default, so `<0>`..`<5>` then `<6>`..),
never taller than `TopSectionRows()`, and the header always occupies exactly
`TopSectionRows()` rows.

## Releases

A release *is* a signed git tag, and consumers pick it up with
`go get github.com/blairham/tuikit@vX.Y.Z`. Newest tag: **v0.0.20**. Each tag
also gets a GitHub release with the tagged source, a cosign-signed
`checksums.txt` and build provenance — see [SECURITY.md](SECURITY.md) for
verifying them.
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

Apache-2.0. See [LICENSE](./LICENSE) and [NOTICE](./NOTICE). Contributions are accepted under the [CLA](./CLA.md); see [CONTRIBUTING.md](./CONTRIBUTING.md).
