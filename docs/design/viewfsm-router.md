# Design: viewfsm.Router

**Status:** living document — describes current behavior
**Code:** `viewfsm/` (`viewfsm.Router`, `viewfsm.ViewID`, `viewfsm.Spec`, `viewfsm.TranslateNavKey`)

## Purpose

`viewfsm.Router` tracks navigation state for a k9s-style app — which view
is active, the drill stack to it, and the digit-hotkey bindings — **without
owning the view objects themselves**. This doc explains that separation
and why it lets one router serve both shallow drill stacks and deep
detail/form stacks.

## The router does not own views

The router tracks three things:

- **which `ViewID` is active** (the top of the drill stack),
- **the drill stack** (root first, active last), and
- **digit-hotkey resolution** (`1..N` → `ViewID`).

Apps keep their concrete view models in their own state and consult the
router for "what's on top?" (`Active`) and "what's the breadcrumb chain?"
(`Breadcrumb`). The router never holds a reference to a view's model, its
columns, or its data — that all stays app-side.

## ViewID and Spec

Routes are plain `ViewID int` constants. The app typedefs its own enum and
registers a `Spec{Name, Hotkey}` per ID, then translates the router's
active ID back to a concrete view in `Update`/`View`:

```go
type ViewID viewfsm.ViewID
const (
    ViewLogGroups ViewID = iota
    ViewStreams
    ViewTail
)
```

`Spec.Name` is the breadcrumb / display label; `Spec.Hotkey` is the digit
key (`"1"`, `"2"`, …) or empty for no hotkey. `NewRouter` builds the
hotkey index from the specs and pushes the initial view as the stack root.
IDs are just ints — tuikit doesn't care about their values, only that
they're unique within a `Router`.

## Stack operations

The router exposes the full drill-stack vocabulary so both navigation
shapes are expressible:

| Method | Effect |
|---|---|
| `Push(id)` | Drill into `id` (no-op if unknown to the router). |
| `Pop()` | Return to the previous view; returns `false` at the root. |
| `Replace(id)` | Swap the active view without changing depth (sibling jump). |
| `JumpTo(id)` | Reset the stack to a single entry (digit hotkey *and* reset depth). |
| `ResolveHotkey(key)` | Look up the `ViewID` bound to a digit key. |
| `Active` / `Stack` / `IsAtRoot` | Read current state. |
| `Breadcrumb()` | Emit one `chrome.Crumb` per stack level, leaf last. |

The two consumer shapes both stay working off this surface:

- **Shallow:** digit hotkeys call `ResolveHotkey` then
  `JumpTo`, so switching the active resource view also resets drill depth.
- **Deep:** detail/form views `Push` onto the stack and `Pop` back;
  `Replace` handles sibling jumps at the current depth.

`Breadcrumb()` returns `chrome.Crumb` values directly, so the result feeds
straight into `chrome.Frame.Breadcrumb` — the router is the bridge between
navigation state and the chrome footer.

## Navigation keys

`TranslateNavKey` rewrites the `j/k/g/G/ctrl+d/ctrl+u` keymap onto the
arrow/page keys that `bubbles` tables understand natively. Apps that strip
`j/k` from the table keymap (via `tuikit/table.KeyMap`) call this before
feeding a key message into the active view's `table.Update`, so vim-style
navigation works without the table double-handling the literal letters.
