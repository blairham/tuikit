# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and the project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
Pre-stable releases (`v0.x.y`) make no API-stability promise — breaking changes can land in any `v0.x` bump.

## [Unreleased]

### Fixed

- The text typed into the command and filter bars is bold, as in k9s's
  prompt (`[::b]`); the suggestion stays regular weight. (#33)

## [0.0.4] - 2026-10-01

### Changed

- The command and filter bars use k9s's prompt colors: the command bar's
  border is the new `Theme.CommandBorder` (aqua) instead of `Accent`, typed
  text is `Theme.InputText` (cadetblue) instead of `Value`, and the inline
  suggestion is `Theme.Suggestion` (dodgerblue) instead of `Muted`. An app
  that recolors its accent no longer recolors the command bar. (#30)

## [0.0.3] - 2026-10-01

### Added

- `viewfsm.TranslateNavKey` maps `h`/`l` to `←`/`→` and `ctrl+f`/`ctrl+b` to
  `pgdown`/`pgup` (k9s's page keys), beside the existing `j`/`k`/`g`/`G`/
  `ctrl+d`/`ctrl+u`. The arrow and page keys keep working unchanged. (#22)
- Conventional key constants in `chrome`: `KeyBack` (`q`), `KeyReload`
  (`ctrl+r`), `KeyHistoryBack`/`KeyHistoryForward` (`[`/`]`), `KeyLastView`
  (`-`), `KeyFieldNext`/`KeyFieldPrev` (`tab`/`shift+tab`). (#22)
- `chrome.GeneralHelp()` and `chrome.NavigationHelp()`: the help overlay's
  GENERAL and NAVIGATION columns in k9s's wording, listing the vim keys and
  not the arrows they translate to. Apps append their own entries. (#22)
- `viewfsm.History`: the visited-view trail behind k9s's `[` back, `]`
  forward and `-` last view, with browser semantics (a visit after going back
  drops the forward entries, repeats collapse, bounded by
  `DefaultHistorySize`). Standalone, so apps with their own view stack can
  use it. `Router.JumpTo` records into the router's own history, and
  `Router.HistoryBack`, `HistoryForward` and `LastView` jump through it
  without re-recording. (#23)

## [0.0.2] - 2026-10-01

### Added

- `theme.Theme.BorderFocus` (default `#87CEFA` LightSkyBlue, k9s's frame
  focus color) and `Theme.FocusBorder()`, which falls back to `Border` when
  `BorderFocus` is nil. `chrome.InjectBorderTitleColor` repaints a box's title
  row in a given color, for boxes not drawn in the focus color. (#18)

### Changed

- The main content box (`Theme.TableBorder`, `Chrome.BorderedContent`) and the
  top line `InjectBorderTitle` repaints now use the focus border color instead
  of `Border`, matching k9s, which draws the focused frame in light sky blue.
  Modals keep `Border`. (#18)
- The header lays out like k9s's whether or not a logo is shown: the shortcut
  columns start two cells after the info panel, left-aligned, and the logo is
  pinned as a block against the right edge with flexible fill in between.
  Logo mode used to right-align every shortcut row against the logo, leaving
  a wide gap after the info panel, and right-justified ragged logo lines one
  by one. A shortcut row too wide for the space before the logo is truncated
  rather than wrapped. (#16)
- The info panel shrinks to its widest `InfoLines` row plus the gap;
  `InfoLabelWidth` (still 56 by default) is now the cap past which a row
  wraps, not a fixed width. (#16)
- `ShortcutGrid` wraps views and actions into columns of `ShortcutRows`
  entries (capped at `TopSectionRows()`) instead of `TopSectionRows()`, so a
  tall info panel or logo no longer makes every shortcut column taller. (#17)
- The default `ShortcutRows` is 6 (was 5), k9s's menu height. This raises the
  minimum `TopSectionRows()`, and so the header reservation
  `ContentInnerSize` subtracts, from 5 to 6 rows. (#17)

### Deprecated

- `Config.ShortcutColumnWidth` / `Chrome.ShortcutColumnWidth` no longer affect
  layout; the shortcut block no longer needs padding to keep the logo aligned.
  The field is kept so existing configs compile. (#16)

### Fixed

- The help overlay's title row was repainted in `Border` while its sides use
  `HelpBorder`, leaving a color seam on the top line; it now uses `HelpBorder`
  throughout. (#18)
- The top section always renders exactly `TopSectionRows()` rows. A header
  with fewer rows of content was drawn short, which moved the content box up a
  row and left a blank row above the footer.

## [0.0.1] - 2026-09-30

### Added

- `theme.Theme.Rebuild()` recomputes the pre-built styles from the current
  color fields, so a color changed after `Default()` / `NoPaintBackground()`
  actually reaches `LogoStyle`, `Title` and the rest. It keeps the theme's
  `PaintBackground` mode. (#7)
- `chrome.Config.ShortcutKeyWidth` / `ShortcutDescWidth` (and the matching
  `Chrome` fields) set the minimum key and description column widths used by
  `Shortcut`, `ShortcutPair` and `ShortcutGrid`. Zero keeps the old 9 / 10.
  (#5)
- `theme.ReassertBackground(s, bgSeq)` and `theme.BackgroundSeq(c)` — the
  shared pass that re-asserts a painted canvas after every SGR that leaves the
  background at default, and the raw escape for a color.
- `chrome.Chrome.HeaderHidden` / `CrumbsHidden` hide the top section and the
  breadcrumb footer (k9s's headless / crumbsless modes). `Render` omits the
  hidden part and `ContentInnerSize` releases its reservation, so the bordered
  content grows into the freed rows. `ToggleHeader` / `ToggleCrumbs` flip them;
  `chrome.KeyToggleHeader` (`ctrl+e`) and `chrome.KeyToggleCrumbs` (`ctrl+g`)
  are the conventional bindings, matched by the app in its own `Update`. The
  `commandbar` example wires both up and now sizes its content through
  `ContentInnerSize`. (#6)

### Changed

- `ShortcutGrid` wraps a views or actions list longer than `TopSectionRows()`
  into additional columns (k9s behavior), so a grid is never taller than the
  header's reservation, and sizes each column from its widest key and
  description. (#4, #5)

### Fixed

- `table.Truncate` now truncates by display width (terminal cells) via
  `ansi.Truncate` instead of slicing bytes. A cut can no longer tear a
  multi-byte rune into invalid UTF-8, multi-byte runes no longer eat three
  cells of budget for one, wide runes (CJK, emoji) no longer overrun
  `maxLen`, and ANSI SGR sequences survive the cut. A `maxLen` below 1 still
  means "no limit". (#3)
- `table.KeyMap` help text for line up/down now reads `↑`/`↓` instead of
  bubbles' `↑/k`/`↓/j`; `k`/`j` are deliberately unbound there (apps route
  them through `viewfsm.TranslateNavKey`), so the help was advertising keys
  that do nothing. The doc comment now says so explicitly. (#9)
- The top section never renders more than `TopSectionRows()` rows: extra
  `Frame.Shortcuts` rows and overflowing `InfoLines` are clipped, so the frame
  no longer outgrows the terminal and pushes the content's bottom border and
  the footer off-screen. (#4)
- In logo mode, shortcut rows past the end of the logo are padded with a blank
  logo-width segment and align with the rows above instead of right-aligning
  under the logo. (#4)
- Shortcut keys and descriptions always keep at least one space before the next
  column: a ten-character description no longer renders as `Containers<a>`,
  and a key wider than its column no longer wraps onto a second line. (#5)
- With `PaintBackground` on, the canvas no longer drops out behind spans that
  end in a combined reset such as systemd's `\x1b[0;32m`, or in `\x1b[49m`.
  `tail.SetBackground` used to re-assert only after bare `\x1b[0m` / `\x1b[m`;
  it now parses each SGR's parameters, skipping the arguments of extended
  colors so a black `38;2;0;0;0` is not read as a reset (#8).
- `Chrome.Render` now applies the same pass to the whole frame, so cells after
  an app's fg-only styled span (e.g. an info-line value) stay on the canvas
  instead of the terminal's own background. Line endings are left clean (#8).
- `table.FixSelectedRow`'s non-selected repaint uses the same pass, so it also
  catches combined resets.

## [0.0.0] - 2026-09-22

Initial release: six composable packages under one module, all rooted on
`theme`. Apps import whichever subset they need; nothing here owns the
`tea.Program` loop.

### Added

**`theme`** — the palette every other package reads. `Default()` is the
canonical k9s-style theme (deep-black canvas, dodger-blue borders,
aqua/fuchsia accents, full background painting). `NoPaintBackground()` keeps
the palette but lets the terminal background show through unstyled gaps, for
tables that override per-cell foregrounds without a background fight.

**`chrome`** — the frame around an app: info panel, shortcuts, logo, bordered
content, breadcrumb and status bar, assembled per tick from a `Frame`.

- `Chrome.VersionLine(label, current, latest)` — k9s-style version row for the
  info panel, padded to the conventional 10-column label width, appending
  ` ⚡ <latest>` in the highlight style when a newer version differs.
- `chrome.Confirm` — inline single-key y/n confirmation in the error palette,
  since most confirms gate destructive actions. Other keys are swallowed.
- `chrome.Modal` — the same keymap rendered as a centered bordered box with a
  title pill and Cancel/OK buttons, for when a visual interrupt is wanted.
- `chrome.Prompt` — one-shot prefilled text input for the save-as / edit-config
  prompts that do not fit command-parse semantics. `SelectAll` mutes the
  prefill and drops it on the first keystroke.
- `chrome.FilterBar` — live filter input; the callback fires on every
  keystroke, Enter closes and keeps the value, Esc closes and clears.
- `chrome.CommandBar` — the `:` command palette, owning the input, active flag,
  error string and Enter/Esc dispatch. Static or per-keystroke suggestions feed
  textinput's ghost-text.
- `chrome.InjectBorderTitle` — rewrites a box's top border so a title sits
  centered in it.

**`table`** — `Styles` derives bubbles-table styles from a theme.
`FixSelectedRow` repairs bubbles-table's per-cell `\x1b[m` resets, which
otherwise clear the row-level Selected background after the first column and
drop cell foregrounds back to their pre-Selected color — leaving the highlight
on one column and the rest of the row often invisible against it.

**`tail`** — a follow-mode viewport. `AppendLines` batches a backfill into a
single render; `PrependLines` inserts older lines while preserving the user's
view; `AtTop` is the cue to fetch more history; `SetBackground` re-asserts the
background after each reset code so formatted log lines do not punch gaps in it.

**`loading`** — an animated spinner beside a rotating one-line tip, for initial
load and between-view transitions. Renders as a single line or theme-painted
and centered.

**`viewfsm`** — `Router`: view registration, drill stack, global key dispatch
and breadcrumb emission. Serves both fixed shallow drill stacks and
arbitrarily-deep stacks with detail/form views.

**`examples/commandbar`** — runnable demo wiring a CommandBar into a minimal
frame.
