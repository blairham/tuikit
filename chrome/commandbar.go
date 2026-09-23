package chrome

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/blairham/tuikit/theme"
)

// CommandBar owns the textinput, active flag, error, and dispatch
// lifecycle for the k9s-style `:` command palette. Apps construct one
// at startup, call [CommandBar.Open] on the `:` keystroke, forward
// messages via [CommandBar.Update], and pass [CommandBar.Input] to
// [Frame.Command] when [CommandBar.Active] is true.
//
// The bar handles a single mode — a `:` palette with optional
// suggestions. Modal variants (y/n confirms, save-as prompts,
// in-place value edits) are intentionally out of scope; build them as
// sibling widgets if needed.
type CommandBar struct {
	suggestFn func(value string) []string
	theme     theme.Theme
	err       string
	input     textinput.Model
	active    bool
}

// CommandBarOpts customizes a CommandBar. The zero value is fine —
// the bar uses a `:` prompt and no suggestions.
//
//nolint:govet // public option struct; readability beats fieldalignment
type CommandBarOpts struct {
	SuggestFn   func(value string) []string
	Prompt      string
	Placeholder string
	Suggestions []string
	Width       int
	CharLimit   int
}

// Dispatch is the application callback invoked when the user presses
// Enter on a non-empty value.
//
//   - Returning a non-empty errMsg keeps the bar open and surfaces
//     the error via [CommandBar.Error]. The cursor stays at the end
//     of the value so the user can correct and retry.
//   - Returning errMsg=="" closes the bar.
//
// The returned tea.Cmd is forwarded by [CommandBar.Update] regardless
// of which branch fires, so apps can return a Cmd alongside an error
// message (e.g. a logging command).
type Dispatch func(value string) (errMsg string, cmd tea.Cmd)

// NewCommandBar returns a CommandBar ready for use. The textinput is
// pre-styled to match the theme; callers can override via
// CommandBar.Input().SetStyles.
func NewCommandBar(t theme.Theme, opts CommandBarOpts) *CommandBar {
	prompt := opts.Prompt
	if prompt == "" {
		prompt = ":"
	}
	in := textinput.New()
	in.Prompt = prompt + " "
	in.Placeholder = opts.Placeholder
	if opts.Width > 0 {
		in.SetWidth(opts.Width)
	}
	if opts.CharLimit > 0 {
		in.CharLimit = opts.CharLimit
	}
	if len(opts.Suggestions) > 0 || opts.SuggestFn != nil {
		in.ShowSuggestions = true
		if len(opts.Suggestions) > 0 {
			in.SetSuggestions(opts.Suggestions)
		}
	}
	in.SetStyles(commandBarStyles(in.Styles(), t))
	return &CommandBar{
		theme:     t,
		input:     in,
		suggestFn: opts.SuggestFn,
	}
}

// Active reports whether the bar is currently open.
func (c *CommandBar) Active() bool { return c.active }

// Open focuses the bar with an empty value and clears any prior
// error. Returns the textinput's focus cmd (cursor blink).
func (c *CommandBar) Open() tea.Cmd {
	c.active = true
	c.err = ""
	c.input.SetValue("")
	if c.suggestFn != nil {
		c.input.SetSuggestions(c.suggestFn(""))
	}
	return c.input.Focus()
}

// OpenWith focuses the bar pre-filled with value, cursor placed at
// the end. Clears any prior error.
func (c *CommandBar) OpenWith(value string) tea.Cmd {
	c.active = true
	c.err = ""
	c.input.SetValue(value)
	c.input.CursorEnd()
	if c.suggestFn != nil {
		c.input.SetSuggestions(c.suggestFn(value))
	}
	return c.input.Focus()
}

// Close blurs the bar and clears its value. The error string is
// preserved so apps can keep surfacing the last error in the status
// bar after the bar itself has closed; call [CommandBar.SetError]("")
// to clear it explicitly.
func (c *CommandBar) Close() {
	c.active = false
	c.input.Blur()
	c.input.SetValue("")
}

// Value returns the current textinput value.
func (c *CommandBar) Value() string { return c.input.Value() }

// Error returns the current error message ("" if none).
func (c *CommandBar) Error() string { return c.err }

// SetError replaces the current error string. Pass "" to clear.
func (c *CommandBar) SetError(s string) { c.err = s }

// SetSuggestions replaces the static suggestion list. Apps that
// registered a SuggestFn should not need this — the function is
// invoked on each keystroke.
func (c *CommandBar) SetSuggestions(suggestions []string) {
	c.input.ShowSuggestions = len(suggestions) > 0
	c.input.SetSuggestions(suggestions)
}

// Input exposes the underlying textinput.Model so apps can wire it
// into [Frame.Command] for rendering.
func (c *CommandBar) Input() *textinput.Model { return &c.input }

// Update forwards msg to the textinput while the bar is active.
//
// Behavior on tea.KeyMsg when active:
//
//   - "esc": closes the bar, clears the error, returns handled=true.
//   - "enter": invokes dispatch with the trimmed value. Empty value
//     closes silently. Otherwise the error/close behavior follows
//     dispatch's return.
//   - any other key: forwarded to the textinput. If a SuggestFn was
//     registered, suggestions are refreshed against the new value.
//
// For non-KeyMsg messages (cursor blink, window resize, etc.) the
// message is forwarded to the textinput so it stays animated, but
// handled is reported as false so the app's outer router still sees
// them.
//
// When the bar is inactive, Update is a no-op and returns
// handled=false.
func (c *CommandBar) Update(msg tea.Msg, dispatch Dispatch) (handled bool, cmd tea.Cmd) {
	if !c.active {
		return false, nil
	}

	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc":
			c.err = ""
			c.Close()
			return true, nil
		case "enter":
			value := strings.TrimSpace(c.input.Value())
			if value == "" {
				c.Close()
				return true, nil
			}
			var (
				errMsg string
				dCmd   tea.Cmd
			)
			if dispatch != nil {
				errMsg, dCmd = dispatch(value)
			}
			if errMsg != "" {
				c.err = errMsg
				c.input.CursorEnd()
				return true, dCmd
			}
			c.err = ""
			c.Close()
			return true, dCmd
		}
		c.input, cmd = c.input.Update(msg)
		if c.suggestFn != nil {
			c.input.SetSuggestions(c.suggestFn(c.input.Value()))
		}
		return true, cmd
	}

	c.input, cmd = c.input.Update(msg)
	return false, cmd
}

// commandBarStyles paints the textinput against the theme. Apps that
// want a different look can override via Input().SetStyles after
// construction.
func commandBarStyles(s textinput.Styles, t theme.Theme) textinput.Styles {
	bg := lipgloss.NewStyle()
	if t.PaintBackground {
		bg = bg.Background(t.Bg)
	}
	s.Focused.Prompt = bg.Foreground(t.Logo).Bold(true)
	s.Focused.Text = bg.Foreground(t.Value)
	s.Focused.Placeholder = bg.Foreground(t.Muted)
	s.Focused.Suggestion = bg.Foreground(t.Muted)
	return s
}
