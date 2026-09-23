# Design: The chrome.Frame per-tick contract

**Status:** living document — describes current behavior
**Code:** `chrome/` (`chrome.Chrome`, `chrome.Frame`, `chrome.CommandBar`, `chrome/filterbar.go`); example at `examples/commandbar/main.go`

## Purpose

The chrome is stateless across ticks: the app assembles a fresh
`chrome.Frame` on every `View()` call and the chrome lays it out. This doc
describes that per-tick contract, the size-before-render ordering it
requires, and the lifecycle of the filter / command bars that ride above
the content.

## The per-tick assembly contract

Apps drive the `tea.Program` loop. Each `View()` call goes through four
steps:

1. **Build a `chrome.Frame{}` for this tick** — `InfoLines`,
   `Shortcuts`, the rendered `Content`, `Breadcrumb`, optional
   `Filter`/`Command` textinputs, optional `StatusBar` /
   `StatusBarLevel`, and `HelpVisible`.
2. **Call `Chrome.ContentInnerSize(w, h, filtering, commanding, confirming, statusBar)` *before* rendering `Content`** — this returns the inner table/viewport dimensions the app must size its view to. The top-section reservation is derived from `Chrome.TopSectionRows()`, **not** hardcoded, so logo / info-panel / shortcut-row counts compose correctly and apps with shorter (or absent) logos don't leave a gap above the footer.
3. **Wrap content** via `Chrome.BorderedContent(...)`, then
   `chrome.InjectBorderTitle(...)` for the centered title pill.
4. **Return `Chrome.Render(frame)`** from `tea.Model.View()`.

`chrome.Frame` is purely a per-tick assembly contract — it holds no state
between ticks. Its `//nolint:govet` for fieldalignment is intentional:
readability of a public API struct wins over byte alignment. The same
applies to `theme.Theme`. **Do not reorder these structs to satisfy
`fieldalignment`.**

## Size before render

The ordering in step 2 is load-bearing. `ContentInnerSize` subtracts a
reservation for every element the chrome will stack around the content:

```
top section (info + shortcuts):  TopSectionRows()
bordered content frame:           2 rows
footer (breadcrumb + bottom gap): 2 rows
filter bar:                       3 rows when active
command bar:                      3 rows when active
confirm bar:                      3 rows when active
status bar:                       1 row when active
```

Because the bars are conditionally present, the app must tell
`ContentInnerSize` which bars are active *this tick* so the content area
shrinks to match. Sizing the view to anything else pushes the footer off
the bottom of the terminal (content too tall) or leaves a gap (too
short). `Chrome.Render` then stacks the same elements in the same order,
gating each one's background on `Theme.PaintBackground` (see
[`theme-and-painting.md`](theme-and-painting.md)).

## Command / filter bar lifecycle

`chrome.CommandBar` owns the `:` palette as a standalone widget: it holds
the textinput, the active flag, the last error, and the dispatch
lifecycle. The app:

1. constructs one at startup (`NewCommandBar`),
2. calls `Open()` (or `OpenWith(value)`) on the `:` keystroke,
3. forwards key messages via `Update(msg, dispatch)` while `Active()`,
4. passes `Input()` to `Frame.Command` so the chrome renders it.

`Update` returns a `handled bool`: when the bar is active and consumes a
key (`esc` closes, `enter` dispatches, other keys edit the value),
`handled` is `true` and the app's outer key router falls through to
nothing else. For non-key messages (cursor blink, resize) the bar still
updates so it stays animated, but reports `handled=false` so the outer
router still sees them. The `Dispatch` callback decides the outcome of
`enter`: a non-empty `errMsg` keeps the bar open and surfaces the error
(apps typically route it into `Frame.StatusBar`); an empty `errMsg`
closes the bar. The filter bar follows the same active/forward/render
shape. `examples/commandbar/main.go` is the canonical wiring.

The bar handles a single mode — a `:` palette with optional suggestions.
Modal variants (y/n confirms via `chrome.Confirm`, save-as prompts,
in-place value edits) are intentionally separate sibling widgets, not
modes of the command bar.
