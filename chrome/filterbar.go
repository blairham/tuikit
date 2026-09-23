package chrome

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/blairham/tuikit/theme"
)

// FilterBar is the live `/filter` input — keystrokes update a view
// filter on every press, Enter commits and closes (filter stays
// applied), Esc cancels and clears.
//
// Apps own the filter target (`view.SetFilter(value)`) and pass a
// callback to [FilterBar.Update]; the bar handles the rest.
//
// Apps wire [FilterBar.Input] into [Frame.Filter] when the bar is
// active.
type FilterBar struct {
	theme  theme.Theme
	input  textinput.Model
	active bool
}

// FilterBarOpts customizes a FilterBar at construction. The zero
// value is fine — a bare textinput with no prompt.
type FilterBarOpts struct {
	// Prompt is rendered before the cursor. Empty means no prompt
	// (the default look). To get a leading "/", set
	// Prompt to "/" — a trailing space is appended automatically.
	Prompt string
	// Placeholder text shown when the value is empty.
	Placeholder string
	// Width is the textinput width in cells. 0 leaves the default.
	Width int
	// CharLimit caps the input length. 0 leaves the default.
	CharLimit int
}

// OnFilter is invoked on every value change (live filter) and again
// on Esc with an empty string (to clear the filter from the view).
//
// The callback returns no error — filter inputs render whatever the
// user typed; views are expected to fall back to "no rows match"
// rather than reject the input. Validation belongs in the view's
// SetFilter implementation, not in the bar.
type OnFilter func(value string)

// NewFilterBar returns a FilterBar ready for use. The textinput is
// pre-styled to match the theme; callers can override via
// FilterBar.Input().SetStyles.
func NewFilterBar(t theme.Theme, opts FilterBarOpts) *FilterBar {
	in := textinput.New()
	if opts.Prompt != "" {
		in.Prompt = opts.Prompt + " "
	} else {
		in.Prompt = ""
	}
	in.Placeholder = opts.Placeholder
	if opts.Width > 0 {
		in.SetWidth(opts.Width)
	}
	if opts.CharLimit > 0 {
		in.CharLimit = opts.CharLimit
	}
	in.SetStyles(filterBarStyles(in.Styles(), t))
	return &FilterBar{theme: t, input: in}
}

// Active reports whether the bar is currently open.
func (f *FilterBar) Active() bool { return f.active }

// Open focuses the bar without clearing the existing value, so users
// can reopen mid-filter and continue editing. Apps that want a fresh
// filter on each open should call [FilterBar.OpenWith]("") instead.
func (f *FilterBar) Open() tea.Cmd {
	f.active = true
	return f.input.Focus()
}

// OpenWith focuses the bar and replaces its value. Cursor is placed
// at the end.
func (f *FilterBar) OpenWith(value string) tea.Cmd {
	f.active = true
	f.input.SetValue(value)
	f.input.CursorEnd()
	return f.input.Focus()
}

// Close blurs the bar. The value is preserved so apps can keep the
// filter applied; call [FilterBar.Clear] to drop it.
func (f *FilterBar) Close() {
	f.active = false
	f.input.Blur()
}

// Clear blurs the bar AND empties the value. The callback (if any)
// is NOT invoked — apps that want the filter dropped on the view
// should call onFilter("") themselves.
func (f *FilterBar) Clear() {
	f.active = false
	f.input.Blur()
	f.input.SetValue("")
}

// Value returns the current textinput value.
func (f *FilterBar) Value() string { return f.input.Value() }

// SetValue replaces the textinput value without touching the active
// flag. The cursor moves to the end. Useful for restoring a filter
// from saved state at startup.
func (f *FilterBar) SetValue(v string) {
	f.input.SetValue(v)
	f.input.CursorEnd()
}

// Input exposes the underlying textinput.Model so apps can wire it
// into [Frame.Filter] for rendering.
func (f *FilterBar) Input() *textinput.Model { return &f.input }

// Update forwards msg to the textinput while the bar is active.
//
// Behavior on tea.KeyMsg when active:
//
//   - "esc": clears the value, closes the bar, invokes
//     onFilter("") so the view drops the filter.
//   - "enter": closes the bar but keeps the value (filter stays
//     applied). onFilter is NOT invoked again — it's already been
//     called for the latest value on the previous keystroke.
//   - any other key: forwarded to the textinput; onFilter is invoked
//     with the new value so the view re-filters live.
//
// Non-KeyMsg messages are forwarded to the textinput but reported as
// handled=false so the app's outer router still sees them.
//
// When inactive, Update is a no-op and returns handled=false.
func (f *FilterBar) Update(msg tea.Msg, onFilter OnFilter) (handled bool, cmd tea.Cmd) {
	if !f.active {
		return false, nil
	}

	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case keyStrEsc:
			f.Clear()
			if onFilter != nil {
				onFilter("")
			}
			return true, nil
		case keyStrEnter:
			f.Close()
			return true, nil
		}
		f.input, cmd = f.input.Update(msg)
		if onFilter != nil {
			onFilter(f.input.Value())
		}
		return true, cmd
	}

	f.input, cmd = f.input.Update(msg)
	return false, cmd
}

// filterBarStyles paints the textinput against the theme's filter
// palette. Apps that want a different look can override via
// Input().SetStyles after construction.
func filterBarStyles(s textinput.Styles, t theme.Theme) textinput.Styles {
	s.Focused.Prompt = t.On(t.Filter).Bold(true)
	s.Focused.Text = t.On(t.Value)
	s.Focused.Placeholder = t.On(t.Muted)
	s.Focused.Suggestion = t.On(t.Muted)
	return s
}
