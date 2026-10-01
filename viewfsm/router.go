package viewfsm

import (
	tea "charm.land/bubbletea/v2"

	"github.com/blairham/tuikit/chrome"
)

// ViewID is an app-defined identifier for one view in the app. Apps
// typically define a typed enum:
//
//	type ViewID viewfsm.ViewID
//	const (
//		ViewLogGroups ViewID = iota
//		ViewStreams
//		ViewTail
//	)
//
// IDs are just ints; tuikit doesn't care about their values, only that
// they're unique within a [Router].
type ViewID int

// Spec describes one view to the [Router]: its breadcrumb label and
// optional digit hotkey ("1", "2", "3", ...). The actual view content
// is owned by the app — viewfsm just tracks which view is active and
// the drill stack to it.
type Spec struct {
	Name   string // breadcrumb / display label
	Hotkey string // "1", "2", ...; empty for no hotkey
}

// Router manages the active view, the drill stack, and digit-hotkey
// resolution. It does NOT own the view objects themselves — apps keep
// concrete view models and consult the router for "which one is
// active now?" and "what's the breadcrumb chain?".
type Router struct {
	specs   map[ViewID]Spec
	hotkeys map[string]ViewID
	history *History
	stack   []ViewID
}

// NewRouter constructs a router with a registered set of views and an
// initial active view. The active view is pushed onto the stack as the
// root; calling [Router.Pop] when only the root is present is a no-op.
func NewRouter(specs map[ViewID]Spec, initial ViewID) *Router {
	hotkeys := make(map[string]ViewID, len(specs))
	for id, s := range specs {
		if s.Hotkey != "" {
			hotkeys[s.Hotkey] = id
		}
	}
	history := NewHistory(0)
	history.Visit(initial)
	return &Router{
		specs:   specs,
		hotkeys: hotkeys,
		history: history,
		stack:   []ViewID{initial},
	}
}

// Active returns the currently-active view ID (top of the stack).
func (r *Router) Active() ViewID {
	if len(r.stack) == 0 {
		var zero ViewID
		return zero
	}
	return r.stack[len(r.stack)-1]
}

// Stack returns a copy of the current drill stack (root first, active
// last).
func (r *Router) Stack() []ViewID {
	out := make([]ViewID, len(r.stack))
	copy(out, r.stack)
	return out
}

// IsAtRoot reports whether the stack contains exactly the root view.
// Useful for "esc was pressed but we're already at the top" branches.
func (r *Router) IsAtRoot() bool {
	return len(r.stack) <= 1
}

// Push drills into the given view, adding it to the stack. If the view
// is unknown to the router, Push is a no-op.
func (r *Router) Push(id ViewID) {
	if _, ok := r.specs[id]; !ok {
		return
	}
	r.stack = append(r.stack, id)
}

// Pop returns to the previous view. Returns true if the stack was
// modified, false if only the root remained.
func (r *Router) Pop() bool {
	if r.IsAtRoot() {
		return false
	}
	r.stack = r.stack[:len(r.stack)-1]
	return true
}

// Replace swaps the active view without changing depth. Use this when
// the user picks a sibling (e.g. jumping via digit hotkey at the
// current depth).
func (r *Router) Replace(id ViewID) {
	if _, ok := r.specs[id]; !ok {
		return
	}
	if len(r.stack) == 0 {
		r.stack = []ViewID{id}
		return
	}
	r.stack[len(r.stack)-1] = id
}

// JumpTo resets the stack to a single entry — the given view — and records
// it in the router's [History]. Use this when a digit hotkey should also
// reset the drill depth — the usual shape for a shallow stack.
func (r *Router) JumpTo(id ViewID) {
	if _, ok := r.specs[id]; !ok {
		return
	}
	r.stack = []ViewID{id}
	r.history.Visit(id)
}

// HistoryBack jumps to the previously visited view without recording the
// move, for k9s's `[`. It reports false, and changes
// nothing, at the oldest entry.
func (r *Router) HistoryBack() (ViewID, bool) { return r.jumpUnrecorded(r.history.Back()) }

// HistoryForward undoes a [Router.HistoryBack], for k9s's `]`.
func (r *Router) HistoryForward() (ViewID, bool) { return r.jumpUnrecorded(r.history.Forward()) }

// LastView returns to the view before the current one, for k9s's `-`;
// called twice it toggles between the two.
func (r *Router) LastView() (ViewID, bool) { return r.jumpUnrecorded(r.history.Last()) }

func (r *Router) jumpUnrecorded(id ViewID, ok bool) (ViewID, bool) {
	if !ok {
		return 0, false
	}
	r.stack = []ViewID{id}
	return id, true
}

// ResolveHotkey returns the view ID bound to the given digit key, if
// any. Apps call this from their global key handler — typically chain
// it with [Router.JumpTo] or [Router.Replace] depending on whether
// digit-keys should reset depth.
func (r *Router) ResolveHotkey(key string) (ViewID, bool) {
	id, ok := r.hotkeys[key]
	return id, ok
}

// Spec returns the registered spec for a view ID, or the zero Spec if
// the ID is unknown.
func (r *Router) Spec(id ViewID) Spec {
	return r.specs[id]
}

// Breadcrumb returns one chrome.Crumb per stack level, in drill order
// (root first, active last with Leaf=true). Apps pass this directly
// to [chrome.Frame.Breadcrumb].
func (r *Router) Breadcrumb() []chrome.Crumb {
	out := make([]chrome.Crumb, 0, len(r.stack))
	for i, id := range r.stack {
		out = append(out, chrome.Crumb{
			Label: r.specs[id].Name,
			Leaf:  i == len(r.stack)-1,
		})
	}
	return out
}

// The modified keys TranslateNavKey maps, as tea.KeyMsg.String spells them.
const (
	navPageDownVim = "ctrl+d"
	navPageUpVim   = "ctrl+u"
	navPageDownK9s = "ctrl+f"
	navPageUpK9s   = "ctrl+b"
)

// TranslateNavKey rewrites the vim-style navigation keys onto the
// arrow/page keys that bubbles tables and viewports understand natively:
//
//	j / k            ↓ / ↑
//	h / l            ← / →
//	g / G            home / end
//	ctrl+f / ctrl+b  pgdown / pgup  (k9s's page keys)
//	ctrl+d / ctrl+u  pgdown / pgup
//
// The arrow, page and home/end keys themselves pass through untouched, so
// both spellings work; [chrome.NavigationHelp] lists only the vim ones. Apps that
// strip j/k from the table keymap (via [tuikit/table.KeyMap]) call this
// before feeding the message into the active view's table.Update — and
// after their own view keys, so a view that binds one of these letters
// (`l` for logs, say) still receives it.
func TranslateNavKey(msg tea.KeyMsg) tea.Msg {
	switch msg.String() {
	case "j":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "k":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "g":
		return tea.KeyPressMsg{Code: tea.KeyHome}
	case "G":
		return tea.KeyPressMsg{Code: tea.KeyEnd}
	case "h":
		return tea.KeyPressMsg{Code: tea.KeyLeft}
	case "l":
		return tea.KeyPressMsg{Code: tea.KeyRight}
	case navPageDownVim, navPageDownK9s:
		return tea.KeyPressMsg{Code: tea.KeyPgDown}
	case navPageUpVim, navPageUpK9s:
		return tea.KeyPressMsg{Code: tea.KeyPgUp}
	}
	return msg
}
