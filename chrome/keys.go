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
