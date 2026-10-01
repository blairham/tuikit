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

### Recoloring a constructed theme

The pre-built styles on `Theme` (`LogoStyle`, `Title`, `ShortcutKey`,
`TableBorder`, …) are computed from the color fields when `Default()` or
`NoPaintBackground()` builds the theme. Assigning a color field afterwards
does **not** reach them on its own; call `Theme.Rebuild()` once the fields
are set:

```go
t := theme.Default()
t.Logo = lipgloss.Color("#2496ED")
t.Accent = lipgloss.Color("#2496ED")
t.Rebuild() // LogoStyle and Title now render in the new colors
```

`Rebuild` keeps the theme's `PaintBackground` setting, so a
`NoPaintBackground()` theme stays unpainted. It overwrites any style
assigned by hand, so override individual styles after calling it.

### Border colors: focused vs unfocused

k9s draws its focused frame in light sky blue (`#87CEFA`, its frame
`focusColor`) and every other border in dodger blue. tuikit mirrors that
with two fields:

- **`Theme.BorderFocus`** (`#87CEFA` in `Default()`) — the main content
  box: `Theme.TableBorder`, so `Chrome.BorderedContent`, and the top line
  `chrome.InjectBorderTitle` repaints around the title.
- **`Theme.Border`** (`#1E90FF`) — unfocused borders: modals, secondary
  panes, and the action-key shortcut color.

Read the focused color through `Theme.FocusBorder()`, which falls back to
`Border` when `BorderFocus` is nil (a `Theme` built by hand rather than
from `Default()`), so such a theme never renders with a nil color.

`InjectBorderTitle` repaints the box's whole top line, so the color it
uses must be the one the side borders use or the title row shows a seam.
It assumes the focused content box. Any box drawn in another color goes
through `InjectBorderTitleColor(box, title, color, theme)` with that
color: the modal passes `Border`, the help overlay passes `HelpBorder`.

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

`tail` has the same hazard for log content, and `Chrome.Render` has it for the
whole frame: any span an app styles with a foreground but no background ends
in a reset. Both run `theme.ReassertBackground`, which parses every SGR and
re-asserts the background after any that leaves it at default — a bare reset,
a combined one such as `\x1b[0;32m`, or `49` — skipping the arguments of
`38`/`48`/`58` extended colors so the zeros of a black color are not read as a
reset. A reset that ends a line is left alone so the background never bleeds
onto the next line. `table.FixSelectedRow`'s non-selected repaint uses the same
pass. Use it rather than a `strings.ReplaceAll` on `\x1b[0m`, which misses
combined resets.

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
