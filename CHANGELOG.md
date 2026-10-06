# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and the project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
Pre-stable releases (`v0.x.y`) make no API-stability promise — breaking changes can land in any `v0.x` bump.

## [Unreleased]

## [0.0.25] - 2026-10-06

### Fixed

- `chrome.Modal`'s buttons are drawn as k9s draws them: the focused one
  black on dodgerblue (it was white), the other in the dialog's cadetblue
  (it was gray), with two cells either side of each label.

### Added

- `theme`: `DialogText`, `DialogButtonFocus`, `DialogButtonFocusText`,
  `DialogLabel` and `DialogField`, set from a skin's `dialog:` block, with
  k9s's defaults in `Default()`; read through `DialogTextColor()` and its
  siblings, which fall back to the colors the dialog drew before.

## [0.0.24] - 2026-10-05

### Changed

- `chrome.Modal` is now the k9s dialog: it draws over `Content` (which
  stays visible around it) instead of replacing it, its buttons are one
  row, and **focus starts on Cancel** — Enter presses the focused button,
  so it no longer means yes by default. `ModalOpts.FocusOK` starts on OK.
  `y`/`n`/esc answer as before. A dispatch error is now shown inside the
  dialog (`Modal.Err`).
- `theme`: `Theme.WithSkin` returns only the theme. A color that does not
  parse is drawn in the terminal's own color, as k9s draws it, instead of
  failing the skin — k9s's own `stock` and `red` skins name `linegreen`
  and were refused.

### Added

- `chrome.SelectField` / `chrome.CheckboxField` form rows for `Modal`
  (k9s's "Propagation" and "Force"), read with `Modal.Value` /
  `Modal.Checked` in the dispatch; tab/arrows move focus, space toggles.
- `theme`: `Skin` carries every key in k9s's skin schema — `dialog`,
  `views.xray`, `views.yaml`, `views.picker`, `views.logs.indicator`, the
  logo's message colors, `info`'s CPU, memory and revision colors, the
  chart dial, background, focus and per-resource colors, and
  `selectedSortColumnColor`, xray's `showIcons` — so every key k9s reads
  has a field, and an app can decode a stock skin without losing any.
- `theme`: `Skin.Check` names every color in a skin that does not parse.
- `theme`/`tree`: a skin's `views.xray` colors draw the tree — label,
  cursor row, cursor text and guides — through the new optional
  `XrayText`, `XrayCursor`, `XrayCursorText` and `XrayGraphic` fields,
  each falling back to the color the tree drew before.
- `theme`: `ParseColor` reads `-` as the terminal's own color and
  `rebeccapurple`, as k9s does.

## [0.0.23] - 2026-10-04

### Fixed

- `chrome`: the help overlay lays sections out in as many columns as fit
  (at least 25 cells each) and stacks the rest under the shortest columns,
  instead of widening the row past the screen and cutting every
  description; and it is cut to the content area's height rather than
  pushing the frame past the terminal.

## [0.0.22] - 2026-10-04

### Added

- `theme`: a k9s skin's `views.charts.defaultChartColors` now sets
  `ChartPrimary` and `ChartSecondary` from its first two entries, so a skin
  recolors charts too. A list with one entry sets only `ChartPrimary`. (#99)

## [0.0.21] - 2026-10-04

### Added

- `chart`: small live charts for a dashboard in the style of k9s's pulses.
  `Series` is a bounded ring buffer of samples; `Sparkline` draws one or
  more series in ▁▂▃▄▅▆▇█ blocks, stacked on taller charts (3 rows give 24
  levels), scaled to `SetMax` or to the data on screen, newest on the
  right, with an optional title and current-value header; `Gauge` draws a
  value against a total as a bar with a label and high- or low-is-bad
  warn/critical thresholds in the status colors; `Grid` lays widgets out in
  rows and columns. Every widget draws exactly its `Resize` size; NaN,
  negative and -Inf samples draw as zero and +Inf as full. See
  `examples/chart`. (#97)
- `theme`: `ChartPrimary` and `ChartSecondary` (k9s's palegreen and
  orangered chart colors) and `Theme.ChartColor(i)`, falling back to
  `Accent` / `AccentAlt` when unset. (#97)

## [0.0.20] - 2026-10-04

### Added

- `tree`: a navigable, collapsible tree view in the style of k9s's xray.
  `SetRoots` takes `Node{ID, Label, Children}` as often as an app polls,
  keeping expansion state and the cursor by ID (a removed selection moves
  to its nearest surviving ancestor); `SetDefaultDepth` sets how deep new
  nodes start open. `HandleKey` moves over visible rows (↑/↓ j/k,
  pgup/pgdown ctrl+b/ctrl+f, home/end g/G), → / l expands or steps into
  the first child, ← / h collapses or steps to the parent, space toggles;
  enter is left to the app. `ExpandAll`, `CollapseAll`, `Selected`,
  `SelectedPath`, `Count`, `VisibleCount`. `SetFilter` takes a
  `table.ParseFilter` expression and keeps a node when it or a descendant
  matches, with the path to every match opened. `View` draws box-drawing
  guides with ▸/▾ markers, the selection in the theme's selection colors,
  and ANSI-aware truncation. See `examples/tree`. (#94)

## [0.0.19] - 2026-10-03

### Added

- `tail`: search beside the filter, as in k9s's describe view. `SetSearch`
  highlights matches in the lines the filter shows without hiding any — it
  is matched against the ANSI-stripped text and woven into lines that
  already carry SGR color without disturbing it. `NextMatch` / `PrevMatch`
  step between matching lines with wraparound, scrolling them into view
  (wrapped rows included) and pausing follow. `Matches` / `MatchIndex` give
  a "3/17" counter, and the current match stays on its line through
  appends, prepends, `ReplaceLines`, trims and filter changes.
  `SetSearchStyles` takes the highlight styles. (#85)
- `theme`: `SearchMatch` (reverse video) and `SearchCurrent` (the prompt
  pill's colors, bold) styles for the search highlight. (#85)
- `table.RowFilter` understands k9s's other two filter-bar modes. `-f term`
  is a fuzzy match: the term's characters must appear in a field in order,
  case-insensitively. `-l selector` is a label selector over a row's
  labels: comma-separated `k=v`, `k==v`, `k!=v` (differs or absent), `k`
  (exists) and `!k` (absent) terms, ANDed; a malformed term makes the
  filter match nothing. The new `Match(fields, labels)` covers all three
  modes, and `Kind()` reports which one is active. `MatchesAny` does fuzzy
  too, and matches nothing for a label filter, which needs labels — so in
  `tail.Model` a label filter shows only markers. The plain regex form,
  with `!` negation and its literal fallback, is unchanged. (#84)
- `chrome`: `CommandBar` and `FilterBar` keep an input history. Up/down
  recall earlier submitted values (enter, not esc), and down past the
  newest restores what was being typed; consecutive duplicates and empty
  values are skipped, and the newest `chrome.HistoryLimit` (50) are kept.
  `History()` / `SetHistory()` let apps persist and seed it. The filter bar
  re-runs `OnFilter` on each recall. (#82)

### Changed

- `chrome.CommandBar`: up/down recall history instead of cycling
  suggestions; suggestions cycle on ctrl+n / ctrl+p, and tab and → still
  accept one. (#82)
- `table.ParseFilter`: an expression that starts with `-f ` or `-l ` (or is
  exactly `-f` or `-l`) now selects the fuzzy or label mode instead of being
  a regex. `-foo` and other expressions are unchanged. (#84)

## [0.0.18] - 2026-10-03

### Fixed

- `table.ParseFilter` (and so `tail.Model.SetFilter`) no longer panics on
  an expression that is not valid UTF-8; the stray bytes become U+FFFD,
  which matches the same bytes in a field. (#74)

### Changed

- Dependencies raised, so a consumer's build moves to at least these:
  bubbles v2.2.1, bubbletea v2.0.8, lipgloss v2.0.5, x/ansi v0.11.8,
  go-colorful v1.4.1, yaml v3.0.5. (#76, #77, #78)

### Added

- Each tag now also gets a GitHub release with the tagged source, a
  cosign-signed `checksums.txt` and SLSA build provenance; SECURITY.md has
  the commands to verify them. (#81)

## [0.0.17] - 2026-10-02

### Added

- `tail.Model.ReplaceLines` swaps the whole buffer in place, keeping the
  filter, the follow state and the scroll offset (clamped when the content
  shrinks) — for a view that refetches a document, so a refresh keeps the
  reader's place. (#71)

## [0.0.16] - 2026-10-02

### Added

- `tail.Model.SetMaxLines` / `MaxLines` cap the buffer: past n lines the
  oldest are dropped, markers included, as k9s caps a log view at
  `logger.buffer`. Scrolled back, the view stays on the lines it shows; a
  full buffer keeps its newest lines over prepended history. 0, the
  default, keeps every line. (#68)

## [0.0.15] - 2026-10-02

### Added

- `table.FixRows(view, theme)`: `FixSelectedRow` with the theme's colors.
  `FixSelectedRow` assumed the default light-sky-blue selection and black
  canvas, so under a skin or `Theme.Inverted` the selected row went
  unrecognized and the padding after every styled cell showed the
  terminal's own background. Apps that skin should call `FixRows`. (#65)
- `Theme.SelectionText` (k9s's `views.table.cursorFgColor`, black by
  default) for the selected row's text, so an inverted theme turns it white
  along with the selection. (#65)

## [0.0.14] - 2026-10-02

### Added

- k9s skins. `theme.Skin` is k9s's skin schema with yaml tags, so an app
  unmarshals a stock k9s skin file into it; `Theme.WithSkin` applies its
  colors and rebuilds the styles, naming the key of any unreadable color.
  `theme.ParseColor` reads a skin color: a CSS name, `#rrggbb` / `#rgb`, or
  `default`, which for the body background stops the theme painting it. (#62)
- `Theme.Inverted` and `theme.InvertColor`: k9s's `--invert`, lightness
  flipped in OkLch with the hue and as much chroma as the sRGB gamut
  allows kept. (#62)
- Optional theme colors a skin sets — `MenuKey`, `MenuNumKey`, `MenuText`,
  `TableText`, `TableHeader`, `LogText` — each falling back to the color
  used before it existed, so a theme that sets none draws as before. (#62)

## [0.0.13] - 2026-10-02

### Changed

- A shortcut key column is as wide as its widest key plus one space, as k9s
  pads keys (`<0> all`), instead of at least 9 cells, which left a wide gap
  before the descriptions of a column of short keys. `Config.ShortcutKeyWidth`
  is now an optional minimum (0, the default, means none). (#59)

## [0.0.12] - 2026-10-02

### Added

- `chrome.SaveDump` writes a view to `<dir>/<name>-<timestamp>.txt` and returns
  the path for a status flash. ANSI is stripped, `dir` is created on demand,
  the name is made filename-safe, and a second save in the same second gets a
  `-2` suffix instead of overwriting. `chrome.KeySave` (ctrl+s) is the
  conventional binding. `table.PlainText` renders a table's header and every
  row as aligned plain text to feed it; a tail's `VisibleLines` serves the same
  purpose for log views. (#26)

- `table.Marks` marks rows for bulk actions as k9s does: `HandleKey` takes
  space (toggle the cursor row), ctrl+space (mark from the last mark to the
  cursor) and ctrl+\ (clear), as `KeyMarkToggle` / `KeyMarkRange` /
  `KeyMarkClear`. Marks are keyed by a row key (the first cell by default),
  so they follow their rows through re-sorts and refreshes; `Prune` drops
  marks whose rows went away. `Selected` gives the rows an action applies
  to (the marked rows, or the cursor row when none are marked), and `Style`
  draws marked rows in the new `theme.Theme.MarkStyle` (`Mark` color,
  palegreen as in k9s). (#24)
- `table.Sorter` gives every table k9s's column sort: `HandleKey` moves the
  sort column on shift+←/→ (`KeySortPrev` / `KeySortNext`), `SortBy` selects
  a column and reverses the direction when it is already active, `Sort` is a
  stable sort over the rows (ANSI ignored, numbers compared as numbers via
  `NaturalCompare`, replaceable with `SetCompare`), and `Columns` appends
  `↑`/`↓` to the active header. Call `Sort` and `Columns` on every refresh. (#25)

### Changed

- **Breaking:** `CommandBar.Error` and `Prompt.Error` are now `ErrMsg`. An
  `Error() string` method made both types satisfy `error` by accident, so
  `fmt` printed a bar as its message and `errors.As` could match one. `SetError`
  is unchanged. (#1)

## [0.0.11] - 2026-10-01

### Added

- `tail.Model.AppendMarker` appends a line that every filter lets through —
  a separator such as a log mark, which has to survive filtering for the
  lines after it. It keeps its place in the buffer, counts in
  `VisibleLines` while shown, and `Clear` drops it. (#51)

## [0.0.10] - 2026-10-01

### Changed

- Breadcrumbs are drawn as k9s draws them: a bold `<name>` pill per level,
  lowercased with spaces removed, black on aqua for the trail and black on
  orange for the current view, one space apart. The theme gains
  `BreadcrumbFg` and `BreadcrumbActive`; `BreadcrumbBg` now defaults to aqua.
  `chrome.CrumbText` gives a label as the footer shows it. (#48)

## [0.0.9] - 2026-10-01

### Added

- `tail.Model` can soft-wrap long lines (`SetWrap` / `Wrap`, which holds
  across the lazy viewport creation in `Resize`), empty its buffer while a
  stream keeps appending (`Clear`), and hand back the filter-passing lines
  for saving or copying (`VisibleLines`, a copy). (#45)

## [0.0.8] - 2026-10-01

### Fixed

- Help overlay entries no longer wrap: each section's key column is as wide
  as its widest key, as in k9s, and a description that still does not fit
  ends in "…" on its own row. At 120 columns a fixed 14-column key left too
  little room, so descriptions spilled into the row below. (#42)

## [0.0.7] - 2026-10-01

### Changed

- Help overlay section headers are plain green (`Theme.HelpSection`,
  `#008000`), with no bold or underline, as in k9s. `HelpSection.TitleColor`
  still overrides. (#39)

## [0.0.6] - 2026-10-01

### Added

- In the command bar, `→` at the end of the input accepts the inline
  suggestion, like `tab` — as in k9s. Mid-text it still moves the cursor,
  and with no suggestion it does nothing new. (#36)

## [0.0.5] - 2026-10-01

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
