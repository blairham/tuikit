package chrome

import (
	tea "charm.land/bubbletea/v2"

	"github.com/blairham/tuikit/theme"
)

// Confirm is a single-key y/n confirmation widget. It does NOT echo
// keystrokes; it consumes one key and dispatches the bool result.
//
//   - y / Y / Enter → yes
//   - n / N / Esc   → no
//   - any other key → swallowed (the bar stays open)
//
// Apps construct one Confirm at startup and reopen it from any
// destructive callsite with a per-Open prompt — the question is the
// only thing that changes between deletions.
//
// Unlike CommandBar / Prompt, Confirm has no textinput. Apps render
// it by reading [Confirm.Prompt] when [Confirm.Active] is true and
// passing it to [Frame.Confirm].
type Confirm struct {
	theme  theme.Theme
	prompt string
	active bool
}

// ConfirmDispatch is the application callback invoked with the y/n
// result.
//
//   - Returning a non-empty errMsg keeps the bar open (the prompt is
//     preserved so the user can re-answer). Apps that want to display
//     the error string should surface it via Frame.StatusBar.
//   - Returning errMsg=="" closes the bar.
//
// The returned tea.Cmd is forwarded regardless of branch.
type ConfirmDispatch func(yes bool) (errMsg string, cmd tea.Cmd)

// NewConfirm returns a Confirm ready for use. The rendered prompt
// is colored by chrome with the theme's error palette since most
// confirms gate destructive actions.
func NewConfirm(t theme.Theme) *Confirm {
	return &Confirm{theme: t}
}

// Active reports whether the bar is currently open.
func (c *Confirm) Active() bool { return c.active }

// Open activates the bar with the given question. The y/n suffix is
// appended automatically — pass just the question ("Delete topic
// foo?"), not the suffix.
func (c *Confirm) Open(prompt string) {
	c.active = true
	c.prompt = prompt
}

// Close deactivates the bar and clears the stored prompt.
func (c *Confirm) Close() {
	c.active = false
	c.prompt = ""
}

// Prompt returns the active question text ("" when inactive). Apps
// pass this to [Frame.Confirm] when [Active] is true.
func (c *Confirm) Prompt() string { return c.prompt }

// Update consumes a single tea.KeyMsg.
//
//   - y / Y / enter: dispatch(true).
//   - n / N / esc:   dispatch(false).
//   - any other key: swallowed, returns handled=true, no dispatch.
//
// Non-KeyMsg messages are not consumed (handled=false) so the app's
// outer router still sees them.
//
// When inactive, Update is a no-op and returns handled=false.
func (c *Confirm) Update(msg tea.Msg, dispatch ConfirmDispatch) (handled bool, cmd tea.Cmd) {
	if !c.active {
		return false, nil
	}
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return false, nil
	}
	switch key.String() {
	case "y", "Y", keyStrEnter:
		return true, c.fire(dispatch, true)
	case "n", "N", keyStrEsc:
		return true, c.fire(dispatch, false)
	}
	// Swallow other keys — protects against y-fat-fingering some
	// adjacent key and accidentally taking the no path.
	return true, nil
}

func (c *Confirm) fire(dispatch ConfirmDispatch, yes bool) tea.Cmd {
	if dispatch == nil {
		c.Close()
		return nil
	}
	errMsg, cmd := dispatch(yes)
	if errMsg != "" {
		// Stay open so the user can re-answer; app surfaces errMsg
		// via Frame.StatusBar.
		return cmd
	}
	c.Close()
	return cmd
}
