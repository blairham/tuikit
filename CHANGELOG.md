# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and the project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
Pre-stable releases (`v0.x.y`) make no API-stability promise — breaking changes can land in any `v0.x` bump.

## [Unreleased]

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
