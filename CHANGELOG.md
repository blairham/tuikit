# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and the project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
Pre-stable releases (`v0.x.y`) make no API-stability promise — breaking changes can land in any `v0.x` bump.

## [Unreleased]

### Added

- `theme.Theme.Rebuild()` recomputes the pre-built styles from the current
  color fields, so a color changed after `Default()` / `NoPaintBackground()`
  actually reaches `LogoStyle`, `Title` and the rest. It keeps the theme's
  `PaintBackground` mode. (#7)
- `chrome.Config.ShortcutKeyWidth` / `ShortcutDescWidth` (and the matching
  `Chrome` fields) set the minimum key and description column widths used by
  `Shortcut`, `ShortcutPair` and `ShortcutGrid`. Zero keeps the old 9 / 10.
  (#5)

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
