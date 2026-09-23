# tuikit docs

Documentation for tuikit, a k9s-style TUI chrome toolkit for
[Bubble Tea v2](https://github.com/charmbracelet/bubbletea) apps. The
library is six composable packages under one module — `theme`, `chrome`,
`table`, `tail`, `loading`, `viewfsm` — that apps import à la carte;
nothing here owns the `tea.Program` loop. Status is pre-stable `v0.0.x`
while the applications that use it migrate onto it.

## How the docs are organized

| Folder | Purpose | When to read |
|---|---|---|
| [`design/`](design/) | Per-subsystem living docs (one per subsystem). Each has a `Status:` and `Code:` header pointing at the implementation. | Before changing a specific subsystem |

The package map and the per-tick rendering flow live in the repo
[`AGENTS.md`](../AGENTS.md) under **Project Structure** and
**Architecture**; the design docs below go a level deeper on the
load-bearing decisions.

## Design docs by subsystem

| Subsystem | Doc |
|---|---|
| Theme dependency injection + the `PaintBackground` knob | [`design/theme-and-painting.md`](design/theme-and-painting.md) |
| `viewfsm.Router` — drill stack, ViewID enums, digit hotkeys | [`design/viewfsm-router.md`](design/viewfsm-router.md) |
| The per-tick `chrome.Frame` contract + bar lifecycle | [`design/chrome-frame-contract.md`](design/chrome-frame-contract.md) |

Each design doc has a `Status:` line (`living document`, `draft`, etc.)
and a `Code:` line pointing at the implementing package(s). If you're
changing the code, update the doc in the same commit.

## Conventions

- **Status / Code header** at the top of every design doc — see
  [`design/theme-and-painting.md`](design/theme-and-painting.md) for the
  canonical shape: `# Design: <Title>`, then `**Status:**`, `**Code:**`,
  then a `## Purpose` section.
- **Update the doc in the same commit that changes the code.** Stale
  design docs cost more than missing ones.
- **en-US spelling** everywhere (misspell locale is US; required by the
  parent `CLAUDE.md`).
- Design docs are descriptive ("the system works like this"), not
  imperative how-tos. Filenames are kebab-case.
