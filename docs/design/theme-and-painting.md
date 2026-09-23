# Design: Theme dependency injection and background painting

**Status:** living document — describes current behavior
**Code:** `theme/` (`theme.Theme`, `theme.Default`, `Theme.On`); painting honored in `chrome/`, `table/`, `tail/`, `loading/`

## Purpose

`theme.Theme` is the dependency-injection seed for the whole toolkit, and
`Theme.PaintBackground` is the single knob that decides how the chrome
deals with the terminal background. This doc captures why the theme is
passed by value into every package (no globals) and why `PaintBackground`
exists as a first-class toggle rather than a fixed choice.

## Theme is the dependency-injection seed

`theme.Theme` is constructed once by the app and passed into
`chrome.Config` (and, transitively, into every other package). Each
package reads colors and pre-built lipgloss styles from the `Theme` it is
handed — there are **no package-level globals**. To recolor the chrome,
an app builds a different `Theme`; nothing else changes.

`theme.Default()` returns the canonical k9s-style palette with
`PaintBackground` enabled. Apps that want a different look construct their
own `Theme` and inject it; the chrome, tables, tail viewports, and loading
screens all follow.

## `PaintBackground` — the load-bearing knob

`Theme.PaintBackground` distinguishes the two terminal aesthetics tuikit
supports:

- **`true`** — every styled span explicitly paints `Background(Theme.Bg)`.
  This is required on terminals (notably macOS Terminal) where the
  default background bleeds through unstyled gaps: lipgloss does not
  restore a surrounding background after an *inner* reset, so any raw
  span emitted after a styled one falls back to the terminal default and
  shows as a gray sliver against the chrome's black. `Default()` returns
  this mode.
- **`false`** — only the outer screen wrapper paints `Bg`. This lets a
  `bubbles/v2/table.Styles.Selected.Foreground` win without a background
  fight, so per-cell foreground overrides render as intended.

Both modes are first-class — neither is a fallback.

### The trailing-fill problem

The top section makes the bleed-through concrete. When the chrome aligns
shortcut/logo rows, the trailing fill on each row is rendered through a
background-painting style (`bgSpaces` in `chrome/chrome.go`). Without it,
the inner styled content (the logo style or the shortcut padder) emits a
reset escape before any raw trailing space, and the outer block's
background is not re-applied to those raw cells. The same pattern recurs
wherever a styled block is followed by padding: every bar (filter,
command, confirm, status), the footer, and the help overlay gate their
`Background(...)` calls on `Theme.PaintBackground`.

`tail` has the same hazard for log content: `SetBackground` re-asserts the
theme background after the reset codes embedded in styled log lines so the
content stays continuous.

## Rule for new styled output

When adding any new styled output, route it through `Theme.On(...)` (or
one of the pre-built styles on `Theme`) so it honors `PaintBackground`
automatically. **Do not** call `Background(...)` directly without gating
it on `Theme.PaintBackground` — an ungated background breaks the
`false` mode and an absent one breaks the `true` mode.

## Field ordering

`theme.Theme` (like `chrome.Frame`) carries a `//nolint:govet` for
fieldalignment: readability of a public API struct wins over byte
alignment. Do not reorder these structs to satisfy `fieldalignment`.
