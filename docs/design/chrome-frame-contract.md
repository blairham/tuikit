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
footer (breadcrumb + bottom gap): 2 rows, 0 when Chrome.CrumbsHidden
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

## The top section never outgrows its reservation

`ContentInnerSize` runs before the frame exists, so it cannot see how many
rows the app will put in `Frame.Shortcuts` or `Frame.InfoLines`; it trusts
`TopSectionRows()`. `Render` keeps that trust honest: the top section is
clipped to `TopSectionRows()` rows, and shortcut rows past it are dropped.
Before this guard, a view with twelve actions on a six-row header drew a
46-line frame on a 40-line terminal, and the content's bottom border and the
footer were the rows that fell off (#4).

The clip works in the other direction too: a header with fewer rows of
content than `TopSectionRows()` is padded to exactly that many, so the content
box starts directly below it rather than one row up, which left a blank row
above the footer.

Apps should not rely on the clip. `ShortcutGrid` wraps a views or actions
list into additional columns of `Config.ShortcutRows` entries (6 by default,
k9s's menu height; views first, then actions), capped at `TopSectionRows()`,
so the header stays within its reservation without losing shortcuts. The
column height is deliberately *not* the header height: a seven-row info panel
makes the header seven rows tall, but k9s still stops the namespace column at
`<5>` and starts `<6>` in the next one (#17). To give a tall action list more
rows instead of more columns, raise `Config.ShortcutRows`; that grows the
reservation `ContentInnerSize` subtracts too. Because 6 is the floor of the
right column, the minimum header is 6 rows.

Grid columns are sized from their widest key and description plus a
one-space gap, floored at `Config.ShortcutKeyWidth` / `ShortcutDescWidth`
(9 / 10 by default). `Shortcut` and `ShortcutPair` apply the same floor and
gap per call, so adjacent columns never run together (#5); only
`ShortcutGrid` can align a column across rows whose cells overflow the floor.

## Header layout

Each header row is laid out the same way with or without a logo (#16):

```
inset | info | gap | shortcut columns (left-aligned) | fill | logo | inset
```

- The info block is the 1-cell left inset, the widest `InfoLines` row, and a
  2-cell gap, capped at `InfoLabelWidth` (56 by default; a row wider than
  that wraps). So the first shortcut column starts right after the info
  panel, where k9s puts it, rather than at a fixed column 56.
- Shortcut rows are left-aligned. A row too wide to leave one blank cell
  before the logo is truncated, never wrapped, so it can neither run into the
  logo nor push the rows below it down.
- The logo is pinned as a block: every line is padded to the widest, and
  that width ends against the 1-cell right inset. Rows past the end of the
  logo get a blank segment as wide as it. The flexible fill between the
  shortcuts and the logo absorbs all the slack.

Before #16 the logo branch right-aligned each whole row (shortcuts padded to
`ShortcutColumnWidth`, then the logo line) against the right edge, which on a
220-column terminal parked the first shortcut at column 103 while the info
panel ended near column 45, and right-justified ragged logo lines one by one.
`ShortcutColumnWidth` no longer affects layout and is kept only so existing
configs compile.

## Hiding the header and footer

The footer can be hidden: `Chrome.CrumbsHidden` (flipped by
`Chrome.ToggleCrumbs`, conventionally bound to `chrome.KeyToggleCrumbs`,
ctrl+g as in k9s) makes `Render` skip the breadcrumb footer and
`ContentInnerSize` release its 2 rows, so the bordered content grows to the
bottom of the terminal. The chrome does not own the key loop — the app
matches the key and calls `ToggleCrumbs` on the `Chrome` it keeps in its
model. Because the reservation lives on `Chrome`, an app that sizes its
content through `ContentInnerSize` (rather than a hardcoded height) picks
up the resize for free.

The header can be hidden the same way: `Chrome.HeaderHidden` (flipped by
`Chrome.ToggleHeader`, conventionally `chrome.KeyToggleHeader`, ctrl+e as in
k9s) makes `Render` skip the top section and `ContentInnerSize` release its
`TopSectionRows()` rows, so the content's top border becomes the first row.

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
