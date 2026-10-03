// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

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
// The bar remembers submitted commands: up recalls older ones and down
// newer ones, k9s-style (see [CommandBar.Update] and
// [CommandBar.History]). Because up/down mean history, they no longer
// cycle suggestions as textinput's default keymap would:
// [CommandBar.Update] keeps them from the textinput, so suggestions
// cycle on ctrl+n / ctrl+p only. Tab and → (at the end of the value)
// accept one.
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
	history   inputHistory
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
//     the error via [CommandBar.ErrMsg]. The cursor stays at the end
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
		history:   newInputHistory(),
	}
}

// Active reports whether the bar is currently open.
func (c *CommandBar) Active() bool { return c.active }

// Open focuses the bar with an empty value and clears any prior
// error. Returns the textinput's focus cmd (cursor blink).
func (c *CommandBar) Open() tea.Cmd {
	c.active = true
	c.err = ""
	c.history.reset()
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
	c.history.reset()
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

// ErrMsg returns the current error message ("" if none).
func (c *CommandBar) ErrMsg() string { return c.err }

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

// History returns a copy of the submitted commands, oldest first, so
// apps can persist them.
func (c *CommandBar) History() []string { return c.history.snapshot() }

// SetHistory replaces the remembered commands, oldest first — for
// seeding from saved state at startup. The recording rules apply:
// empty values are skipped, a value equal to the one before it is
// dropped, and only the newest [HistoryLimit] are kept. The slice is
// copied. Any recall in progress ends.
func (c *CommandBar) SetHistory(entries []string) { c.history.set(entries) }

// Update forwards msg to the textinput while the bar is active.
//
// Behavior on tea.KeyMsg when active:
//
//   - "esc": closes the bar, clears the error, returns handled=true.
//   - "enter": invokes dispatch with the trimmed value. Empty value
//     closes silently. Otherwise the value is recorded in the history
//     — even when dispatch returns an error, so a mistyped command can
//     be recalled and fixed — and the error/close behavior follows
//     dispatch's return.
//   - "up" / "down": recall older / newer history entries into the
//     input, cursor at the end. Down past the newest restores what was
//     typed before recall began. Recalled text edits like typed text.
//     With nothing to recall the key is consumed and does nothing; it
//     never cycles suggestions (ctrl+n / ctrl+p do that).
//   - any other key: forwarded to the textinput. If a SuggestFn was
//     registered, suggestions are refreshed against the new value.
//     → at the end of the value accepts the current suggestion, as tab
//     does.
//
// Opening, submitting and canceling all end any recall in progress.
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
		case keyStrEsc:
			c.err = ""
			c.Close()
			return true, nil
		case keyStrEnter:
			value := strings.TrimSpace(c.input.Value())
			if value == "" {
				c.Close()
				return true, nil
			}
			var (
				errMsg string
				dCmd   tea.Cmd
			)
			// Record before dispatching, so a dispatch that persists
			// History() sees the command it is running.
			c.history.add(value)
			c.history.reset()
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
		case keyStrUp, keyStrDown:
			if changed := c.history.recall(&c.input, key.String()); changed && c.suggestFn != nil {
				c.input.SetSuggestions(c.suggestFn(c.input.Value()))
			}
			return true, nil
		case "right":
			// → at the end of the input accepts the suggestion, as tab does
			// and as in k9s. Mid-text it still moves the cursor.
			if c.input.Position() == len([]rune(c.input.Value())) && c.input.CurrentSuggestion() != "" {
				msg = tea.KeyPressMsg{Code: tea.KeyTab}
			}
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
	// Bold typed text and a plain suggestion, as k9s writes its prompt.
	s.Focused.Text = bg.Foreground(t.InputText).Bold(true)
	s.Focused.Placeholder = bg.Foreground(t.Muted)
	s.Focused.Suggestion = bg.Foreground(t.Suggestion)
	return s
}
