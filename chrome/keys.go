// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package chrome

// Key-string constants shared by the chrome bar widgets
// (CommandBar / Prompt / Confirm / FilterBar). These match
// [bubbletea/v2.KeyPressMsg.String]'s output for the corresponding
// keys.
const (
	keyStrEsc   = "esc"
	keyStrEnter = "enter"
)

// KeyToggleCrumbs is the conventional binding for [Chrome.ToggleCrumbs]
// (k9s uses ctrl+g). The chrome does not own the key loop — apps match
// it in their own Update and call ToggleCrumbs.
const KeyToggleCrumbs = "ctrl+g"

// KeyToggleHeader is the conventional binding for [Chrome.ToggleHeader]
// (k9s uses ctrl+e). Like [KeyToggleCrumbs], apps match it in their own
// Update and call ToggleHeader.
const KeyToggleHeader = "ctrl+e"

// Conventional bindings for the general shortcuts every k9s-style app
// shares. Like [KeyToggleCrumbs], these are names, not behavior: the chrome
// does not own the key loop, so apps match them in their own Update. They
// are what [GeneralHelp] and [NavigationHelp] advertise.
const (
	// KeyBack leaves the current view, like esc without clearing a filter.
	KeyBack = "q"
	// KeyReload refetches the active view.
	KeyReload = "ctrl+r"
	// KeyHistoryBack and KeyHistoryForward walk the visited-view history.
	KeyHistoryBack    = "["
	KeyHistoryForward = "]"
	// KeyLastView returns to the previously used view; pressed again, it
	// toggles back.
	KeyLastView = "-"
	// KeyFieldNext and KeyFieldPrev move between fields of a form.
	KeyFieldNext = "tab"
	KeyFieldPrev = "shift+tab"
)
