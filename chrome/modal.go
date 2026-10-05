// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package chrome

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/blairham/tuikit/theme"
)

// Modal is a centered popup dialog drawn over the content area — the
// shape k9s uses for its delete / restart / scale prompts. Where
// [Confirm] reserves a strip at the top of the screen, Modal floats a
// bordered box over the table, which stays visible around it.
//
// A dialog has a message, optional form fields ([SelectField],
// [CheckboxField]) and a Cancel / OK button pair. As in k9s, focus
// starts on Cancel, so a stray Enter never confirms a destructive
// action; set [ModalOpts.FocusOK] for prompts where OK is the safe
// default.
//
// Keys:
//
//   - tab / down, shift+tab / up: move focus through fields and buttons.
//   - left / right: on the buttons, move between them; on a select
//     field, step through its options.
//   - space: toggle a checkbox, step a select forward.
//   - enter: press the focused button; on a field, act like space.
//   - esc: cancel. y / Y answer yes and n / N answer no from anywhere,
//     the same accelerators [Confirm] takes.
//
// Every other key is swallowed while the modal is open.
type Modal struct {
	theme       theme.Theme
	title       string
	prompt      string
	okLabel     string
	cancelLabel string
	errMsg      string
	fields      []ModalField
	focus       int
	active      bool
}

// ModalDispatch is the callback invoked with the dialog's answer. Read
// the field values with [Modal.Value] and [Modal.Checked] inside it —
// they are cleared when the modal closes. A non-empty errMsg keeps the
// modal open and shows the message inside it, so the user can change a
// field or re-answer; "" closes it.
type ModalDispatch func(yes bool) (errMsg string, cmd tea.Cmd)

// ModalOpts customizes one Open. Zero values fall back to "Confirm" /
// "OK" / "Cancel", no fields, and focus on Cancel.
type ModalOpts struct {
	Title       string
	OkLabel     string
	CancelLabel string
	// Fields are the form rows drawn between the message and the
	// buttons, in focus order.
	Fields []ModalField
	// FocusOK starts focus on OK instead of Cancel.
	FocusOK bool
}

// ModalField is one form row in a [Modal]: a select (one of Options)
// or a checkbox. Build them with [SelectField] and [CheckboxField].
type ModalField struct {
	Key      string
	Label    string
	Options  []string
	selected int
	checkbox bool
	checked  bool
}

// SelectField returns a field that cycles through options, starting on
// the first — k9s's "Propagation: Background" row. Read the choice
// with [Modal.Value](key).
func SelectField(key, label string, options ...string) ModalField {
	return ModalField{Key: key, Label: label, Options: options}
}

// CheckboxField returns an on/off field — k9s's "Force:" row. Read it
// with [Modal.Checked](key).
func CheckboxField(key, label string, checked bool) ModalField {
	return ModalField{Key: key, Label: label, checkbox: true, checked: checked}
}

// NewModal returns a Modal ready for use.
func NewModal(t theme.Theme) *Modal {
	return &Modal{theme: t}
}

// Active reports whether the modal is currently open.
func (m *Modal) Active() bool { return m.active }

// Open activates the modal with the given message and options. Pass
// just the question text — the chrome draws the border, title, fields
// and buttons.
func (m *Modal) Open(prompt string, opts ModalOpts) {
	m.active = true
	m.prompt = prompt
	m.errMsg = ""
	m.title = or(opts.Title, "Confirm")
	m.okLabel = or(opts.OkLabel, "OK")
	m.cancelLabel = or(opts.CancelLabel, "Cancel")
	m.fields = append([]ModalField(nil), opts.Fields...)
	m.focus = m.cancelIndex()
	if opts.FocusOK {
		m.focus = m.okIndex()
	}
}

func or(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

// Close deactivates the modal and clears its state.
func (m *Modal) Close() {
	*m = Modal{theme: m.theme}
}

// Prompt returns the active question text ("" when inactive).
func (m *Modal) Prompt() string { return m.prompt }

// Title returns the active modal title ("" when inactive).
func (m *Modal) Title() string { return m.title }

// Err returns the message the last dispatch kept the modal open with.
func (m *Modal) Err() string { return m.errMsg }

// Value returns the selected option of the select field key, or "" if
// there is no such select field.
func (m *Modal) Value(key string) string {
	for _, f := range m.fields {
		if f.Key == key && !f.checkbox && len(f.Options) > 0 {
			return f.Options[f.selected]
		}
	}
	return ""
}

// Checked reports whether the checkbox field key is on.
func (m *Modal) Checked(key string) bool {
	for _, f := range m.fields {
		if f.Key == key && f.checkbox {
			return f.checked
		}
	}
	return false
}

// Focus order is the fields, then Cancel, then OK.
func (m *Modal) cancelIndex() int { return len(m.fields) }
func (m *Modal) okIndex() int     { return len(m.fields) + 1 }
func (m *Modal) focusables() int  { return len(m.fields) + 2 }

// OKFocused reports whether focus is on the OK button.
func (m *Modal) OKFocused() bool { return m.active && m.focus == m.okIndex() }

// Update consumes a single tea.KeyMsg; see [Modal] for the keys.
// Non-KeyMsg messages are not consumed (handled=false). When inactive,
// Update is a no-op and returns handled=false.
func (m *Modal) Update(msg tea.Msg, dispatch ModalDispatch) (handled bool, cmd tea.Cmd) {
	if !m.active {
		return false, nil
	}
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return false, nil
	}
	onField := m.focus < len(m.fields)
	switch key.String() {
	case "y", "Y":
		return true, m.fire(dispatch, true)
	case "n", "N", keyStrEsc:
		return true, m.fire(dispatch, false)
	case "tab", "down":
		m.focus = (m.focus + 1) % m.focusables()
	case "shift+tab", "up":
		m.focus = (m.focus + m.focusables() - 1) % m.focusables()
	case "left", "h":
		if onField {
			m.step(-1)
		} else {
			m.focus = m.cancelIndex()
		}
	case "right", "l":
		if onField {
			m.step(1)
		} else {
			m.focus = m.okIndex()
		}
	case "space":
		if onField {
			m.step(1)
		}
	case keyStrEnter:
		if onField {
			m.step(1)
			return true, nil
		}
		return true, m.fire(dispatch, m.focus == m.okIndex())
	}
	return true, nil
}

// step toggles the focused checkbox or moves the focused select by d.
func (m *Modal) step(d int) {
	f := &m.fields[m.focus]
	if f.checkbox {
		f.checked = !f.checked
		return
	}
	if n := len(f.Options); n > 0 {
		f.selected = (f.selected + d + n) % n
	}
}

func (m *Modal) fire(dispatch ModalDispatch, yes bool) tea.Cmd {
	if dispatch == nil {
		m.Close()
		return nil
	}
	errMsg, cmd := dispatch(yes)
	if errMsg != "" {
		m.errMsg = errMsg
		return cmd
	}
	m.Close()
	return cmd
}

// overlayModal draws the modal box over content, centered in it, and
// returns content with the box composited on top. The result keeps
// content's exact dimensions: the box is clipped rather than allowed to
// grow the frame.
func (c Chrome) overlayModal(content string, m *Modal) string {
	w, h := lipgloss.Width(content), lipgloss.Height(content)
	if w == 0 || h == 0 {
		return content
	}
	box := c.renderModalBox(m, w-4)
	x := max((w-lipgloss.Width(box))/2, 0)
	y := max((h-lipgloss.Height(box))/2, 0)
	comp := lipgloss.NewCompositor(
		lipgloss.NewLayer(content),
		lipgloss.NewLayer(box).X(x).Y(y).Z(1),
	)
	return lipgloss.NewCanvas(w, h).Compose(comp).Render()
}

// renderModalBox renders the bordered dialog at most maxW cells wide.
func (c Chrome) renderModalBox(m *Modal, maxW int) string {
	t := c.Theme
	// ~72 cells is the width k9s gives its delete dialog; shrink to fit.
	boxW := min(72, maxW)
	boxW = max(boxW, 20)
	innerW := boxW - 4 // -2 border, -2 horizontal padding

	line := func(s lipgloss.Style) lipgloss.Style {
		s = s.Width(innerW)
		if t.PaintBackground {
			s = s.Background(t.Bg)
		}
		return s
	}
	blank := line(lipgloss.NewStyle()).Render("")

	rows := []string{
		blank,
		line(lipgloss.NewStyle().Foreground(t.HelpDesc).Align(lipgloss.Center)).
			Render(WrapText(m.prompt, innerW)),
		blank,
	}
	for i, f := range m.fields {
		rows = append(rows, c.renderModalField(f, i == m.focus, innerW))
	}
	if len(m.fields) > 0 {
		rows = append(rows, blank)
	}
	if m.errMsg != "" {
		rows = append(rows,
			line(lipgloss.NewStyle().Foreground(t.Status.Error).Align(lipgloss.Center)).
				Render(WrapText(m.errMsg, innerW)),
			blank)
	}
	rows = append(rows, c.renderModalButtons(m, innerW))

	box := lipgloss.NewStyle().
		Width(boxW).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(t.Border).
		Padding(0, 1)
	if t.PaintBackground {
		box = box.Background(t.Bg).BorderBackground(t.Bg)
	}
	title := t.On(t.HelpTitle).Render("<" + m.title + ">")
	return InjectBorderTitleColor(box.Render(strings.Join(rows, "\n")), title, t.Border, t)
}

// renderModalField renders one "Label: value" row. The focused field's
// value is drawn in the selection colors, the way k9s marks the form
// item that has focus.
func (c Chrome) renderModalField(f ModalField, focused bool, innerW int) string {
	t := c.Theme
	var val string
	switch {
	case f.checkbox && f.checked:
		val = "[x]"
	case f.checkbox:
		val = "[ ]"
	case len(f.Options) > 0:
		val = f.Options[f.selected]
	}
	valStyle := t.On(t.Value)
	if focused {
		valStyle = lipgloss.NewStyle().Foreground(t.SelectionText).Background(t.Selection)
	}
	row := t.On(t.Value).Render(f.Label+": ") + valStyle.Render(val)
	pad := lipgloss.NewStyle().Width(innerW)
	if t.PaintBackground {
		pad = pad.Background(t.Bg)
	}
	return pad.Render(row)
}

// renderModalButtons renders the one-row Cancel / OK pair, the focused
// one filled in the border color like a k9s dialog button.
func (c Chrome) renderModalButtons(m *Modal, innerW int) string {
	t := c.Theme
	button := func(label string, focused bool) string {
		if focused {
			return lipgloss.NewStyle().Padding(0, 1).
				Foreground(t.Value).Background(t.Border).Render(label)
		}
		return t.On(t.Muted).Padding(0, 1).Render(label)
	}
	row := button(m.cancelLabel, m.focus == m.cancelIndex()) +
		t.On(t.Muted).Render("  ") +
		button(m.okLabel, m.focus == m.okIndex())
	s := lipgloss.NewStyle().Width(innerW).Align(lipgloss.Center)
	if t.PaintBackground {
		s = s.Background(t.Bg)
	}
	return s.Render(row)
}
