package chrome

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/blairham/tuikit/theme"
)

// Prompt is a one-shot text input — a transient bar that asks the user
// for a value, then dispatches and closes. Unlike [CommandBar], there
// is no command-vs-args parsing: the dispatcher receives the typed
// string verbatim.
//
// Apps construct one Prompt at startup and reopen it from many call
// sites with different prefills, placeholders, and prompt labels —
// e.g. one bar serving "rename to:" / "save to file:" / "purge after:"
// — by passing per-call options to [Prompt.Open].
//
// Apps wire [Prompt.Input] into [Frame.Command]; the Prompt and
// [CommandBar] share the same chrome slot and are mutually exclusive
// at the app level.
type Prompt struct {
	theme  theme.Theme
	err    string
	input  textinput.Model
	active bool
}

// PromptOpts customizes a Prompt at construction. The zero value is
// fine — a `"> "` prompt with no charlimit.
type PromptOpts struct {
	// Prompt is the prefix shown before the cursor. Defaults to ">".
	// A trailing space is appended automatically. Apps that want the
	// prompt to change per call should set it on [OpenOpts] instead.
	Prompt string
	// Width is the textinput width in cells. 0 leaves the default.
	Width int
	// CharLimit caps the input length. 0 leaves the default.
	CharLimit int
}

// OpenOpts customizes a single [Prompt.Open] call. The zero value is
// fine — empty prefill, no placeholder, no select-all.
//
//nolint:govet // public option struct; readability beats fieldalignment
type OpenOpts struct {
	// Prompt overrides the constructor's prompt label for this open
	// only. Empty leaves the existing label.
	Prompt string
	// Placeholder text shown when the value is empty.
	Placeholder string
	// SelectAll mutes the prefill in the placeholder style and clears
	// it on the first keystroke. Used for "edit this value or type a
	// fresh one" prompts where the prefill is a hint, not the answer.
	SelectAll bool
}

// NewPrompt returns a Prompt ready for use. The textinput is
// pre-styled to match the theme; callers can override via
// Prompt.Input().SetStyles.
func NewPrompt(t theme.Theme, opts PromptOpts) *Prompt {
	prompt := opts.Prompt
	if prompt == "" {
		prompt = ">"
	}
	in := textinput.New()
	in.Prompt = prompt + " "
	if opts.Width > 0 {
		in.SetWidth(opts.Width)
	}
	if opts.CharLimit > 0 {
		in.CharLimit = opts.CharLimit
	}
	in.SetStyles(commandBarStyles(in.Styles(), t))
	return &Prompt{theme: t, input: in}
}

// Active reports whether the prompt is currently open.
func (p *Prompt) Active() bool { return p.active }

// Open focuses the prompt with the given prefill value. Per-call
// options override the constructor's defaults.
//
// When opts.SelectAll is true, the prefill renders in the muted
// placeholder style and is cleared on the first user keystroke — the
// callsite gets one "edit the suggestion or replace it" flow without
// having to track a "first key seen" flag itself.
func (p *Prompt) Open(value string, opts OpenOpts) tea.Cmd {
	p.active = true
	p.err = ""
	if opts.Prompt != "" {
		p.input.Prompt = opts.Prompt + " "
	}
	p.input.Placeholder = opts.Placeholder
	if opts.SelectAll && value != "" {
		// Show the prefill as a placeholder so it renders in the muted
		// style; the value itself stays empty so the first keystroke
		// replaces it instead of appending.
		p.input.Placeholder = value
		p.input.SetValue("")
	} else {
		p.input.SetValue(value)
		p.input.CursorEnd()
	}
	return p.input.Focus()
}

// Close blurs the prompt, clears its value, and resets the
// placeholder. The error string is preserved so apps can surface the
// last error in the status bar after close; call SetError("") to
// clear.
func (p *Prompt) Close() {
	p.active = false
	p.input.Blur()
	p.input.SetValue("")
	p.input.Placeholder = ""
}

// Value returns the current textinput value.
func (p *Prompt) Value() string { return p.input.Value() }

// Error returns the current error message ("" if none).
func (p *Prompt) Error() string { return p.err }

// SetError replaces the current error string. Pass "" to clear.
func (p *Prompt) SetError(s string) { p.err = s }

// Input exposes the underlying textinput.Model so apps can wire it
// into [Frame.Command] for rendering.
func (p *Prompt) Input() *textinput.Model { return &p.input }

// Update forwards msg to the textinput while the prompt is active.
//
// Behavior on tea.KeyMsg when active:
//
//   - "esc": closes the prompt, clears the error, returns
//     handled=true.
//   - "enter": invokes dispatch with the trimmed value. Empty value
//     closes silently. Non-empty errMsg from dispatch keeps the
//     prompt open with the cursor at end; empty errMsg closes.
//   - any other key: forwarded to the textinput.
//
// Non-KeyMsg messages are forwarded to the textinput but reported as
// handled=false so the app's outer router still sees them.
//
// When inactive, Update is a no-op and returns handled=false.
func (p *Prompt) Update(msg tea.Msg, dispatch Dispatch) (handled bool, cmd tea.Cmd) {
	if !p.active {
		return false, nil
	}

	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case keyStrEsc:
			p.err = ""
			p.Close()
			return true, nil
		case keyStrEnter:
			value := strings.TrimSpace(p.input.Value())
			if value == "" {
				p.Close()
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
				p.err = errMsg
				p.input.CursorEnd()
				return true, dCmd
			}
			p.err = ""
			p.Close()
			return true, dCmd
		}
		// Once a real keystroke arrives, drop the SelectAll placeholder
		// so the textinput renders the typed value normally.
		p.input.Placeholder = ""
		p.input, cmd = p.input.Update(msg)
		return true, cmd
	}

	p.input, cmd = p.input.Update(msg)
	return false, cmd
}
