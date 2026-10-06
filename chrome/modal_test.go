// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package chrome

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"

	"github.com/blairham/tuikit/theme"
)

func TestModal_InitialState(t *testing.T) {
	t.Parallel()
	m := NewModal(theme.Default())
	if m.Active() {
		t.Error("new Modal should be inactive")
	}
	if m.Prompt() != "" || m.Title() != "" {
		t.Errorf("new Modal should be empty; prompt=%q title=%q", m.Prompt(), m.Title())
	}
}

func TestModal_OpenDefaultsAndCustom(t *testing.T) {
	t.Parallel()
	m := NewModal(theme.Default())

	m.Open("Delete topic foo?", ModalOpts{})
	if !m.Active() {
		t.Error("Open should activate")
	}
	if m.Prompt() != "Delete topic foo?" {
		t.Errorf("Prompt() = %q", m.Prompt())
	}
	if m.Title() != "Confirm" {
		t.Errorf("default Title() = %q; want %q", m.Title(), "Confirm")
	}
	if m.okLabel != "OK" || m.cancelLabel != "Cancel" {
		t.Errorf("default button labels: ok=%q cancel=%q", m.okLabel, m.cancelLabel)
	}

	m.Open("Apply changes?", ModalOpts{
		Title:       "Re-authenticate",
		OkLabel:     "Login",
		CancelLabel: "Skip",
	})
	if m.Title() != "Re-authenticate" || m.okLabel != "Login" || m.cancelLabel != "Skip" {
		t.Errorf("custom opts not honored: title=%q ok=%q cancel=%q", m.Title(), m.okLabel, m.cancelLabel)
	}
}

func TestModal_CloseClears(t *testing.T) {
	t.Parallel()
	m := NewModal(theme.Default())
	m.Open("question?", ModalOpts{Title: "X"})
	m.Close()
	if m.Active() || m.Prompt() != "" || m.Title() != "" {
		t.Errorf("after Close: active=%v prompt=%q title=%q", m.Active(), m.Prompt(), m.Title())
	}
}

func TestModal_UpdateInactiveIsNoop(t *testing.T) {
	t.Parallel()
	m := NewModal(theme.Default())
	handled, cmd := m.Update(keyChar('y'), nil)
	if handled || cmd != nil {
		t.Errorf("inactive Update should be no-op; handled=%v cmd=%v", handled, cmd)
	}
}

func TestModal_KeysFireDispatchAndClose(t *testing.T) {
	t.Parallel()
	cases := []struct {
		key  tea.KeyMsg
		name string
		opts ModalOpts
		want bool
	}{
		{name: "lowercase y", key: keyChar('y'), want: true},
		{name: "uppercase Y", key: keyChar('Y'), want: true},
		{name: "enter on default Cancel", key: keyEnter(), want: false},
		{name: "enter with FocusOK", key: keyEnter(), opts: ModalOpts{FocusOK: true}, want: true},
		{name: "lowercase n", key: keyChar('n'), want: false},
		{name: "uppercase N", key: keyChar('N'), want: false},
		{name: "esc", key: keyEsc(), want: false},
		{name: "esc with FocusOK", key: keyEsc(), opts: ModalOpts{FocusOK: true}, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := NewModal(theme.Default())
			m.Open("Drop the table?", tc.opts)
			got := !tc.want
			dispatch := func(yes bool) (string, tea.Cmd) {
				got = yes
				return "", nil
			}
			handled, _ := m.Update(tc.key, dispatch)
			if !handled {
				t.Errorf("%s: should be handled", tc.name)
			}
			if got != tc.want {
				t.Errorf("%s: dispatch received %v; want %v", tc.name, got, tc.want)
			}
			if m.Active() {
				t.Errorf("%s: should close after answer", tc.name)
			}
		})
	}
}

// k9s starts its delete dialog on Cancel: moving to OK takes a
// deliberate keystroke, and only then does Enter confirm.
func TestModal_FocusMovesBetweenButtons(t *testing.T) {
	t.Parallel()
	for _, move := range []tea.KeyPressMsg{keyRight(), keyTab(), keyChar('l')} {
		m := NewModal(theme.Default())
		m.Open("Delete?", ModalOpts{})
		if m.OKFocused() {
			t.Fatal("focus should start on Cancel")
		}
		m.Update(move, nil)
		if !m.OKFocused() {
			t.Fatalf("%q should move focus to OK", move.String())
		}
		var got bool
		m.Update(keyEnter(), func(yes bool) (string, tea.Cmd) { got = yes; return "", nil })
		if !got {
			t.Errorf("enter on OK after %q should answer yes", move.String())
		}
	}
	m := NewModal(theme.Default())
	m.Open("Delete?", ModalOpts{FocusOK: true})
	m.Update(keyLeft(), nil)
	if m.OKFocused() {
		t.Error("left should move focus back to Cancel")
	}
}

func TestModal_Fields(t *testing.T) {
	t.Parallel()
	m := NewModal(theme.Default())
	m.Open("Delete pod?", ModalOpts{Fields: []ModalField{
		SelectField("propagation", "Propagation", "Background", "Foreground", "Orphan"),
		CheckboxField("force", "Force", false),
	}})
	if got := m.Value("propagation"); got != "Background" {
		t.Errorf("select should start on its first option, got %q", got)
	}
	// Focus starts on Cancel; shift+tab twice reaches the select, then
	// the checkbox via tab.
	m.Update(keyShiftTab(), nil)
	m.Update(keyShiftTab(), nil)
	m.Update(keyRight(), nil)
	if got := m.Value("propagation"); got != "Foreground" {
		t.Errorf("right on the select = %q; want Foreground", got)
	}
	m.Update(keyLeft(), nil)
	m.Update(keyLeft(), nil)
	if got := m.Value("propagation"); got != "Orphan" {
		t.Errorf("left should wrap to the last option, got %q", got)
	}
	m.Update(keyTab(), nil)
	m.Update(keyChar(' '), nil)
	if !m.Checked("force") {
		t.Error("space on the checkbox should check it")
	}
	m.Update(keyEnter(), nil)
	if m.Checked("force") || !m.Active() {
		t.Error("enter on a field should toggle it, not answer the dialog")
	}
	m.Update(keyEnter(), nil)

	var force bool
	var prop string
	m.Update(keyChar('y'), func(bool) (string, tea.Cmd) {
		force, prop = m.Checked("force"), m.Value("propagation")
		return "", nil
	})
	if !force || prop != "Orphan" {
		t.Errorf("dispatch should see the field values: force=%v propagation=%q", force, prop)
	}
	if m.Checked("force") || m.Value("propagation") != "" {
		t.Error("Close should clear the fields")
	}
}

func TestModal_OpenDoesNotShareFields(t *testing.T) {
	t.Parallel()
	fields := []ModalField{CheckboxField("force", "Force", false)}
	m := NewModal(theme.Default())
	m.Open("?", ModalOpts{Fields: fields})
	m.Update(keyShiftTab(), nil)
	m.Update(keyChar(' '), nil)
	if fields[0].checked {
		t.Error("toggling a field must not write through to the caller's slice")
	}
}

func TestModal_OtherKeysAreSwallowed(t *testing.T) {
	t.Parallel()
	m := NewModal(theme.Default())
	m.Open("are you sure?", ModalOpts{})
	called := false
	dispatch := func(bool) (string, tea.Cmd) {
		called = true
		return "", nil
	}
	for _, r := range "abcdefghijklmopqrstuvwxz" {
		handled, _ := m.Update(keyChar(r), dispatch)
		if !handled {
			t.Errorf("non y/n key %q should be swallowed (handled=true)", r)
		}
	}
	if called {
		t.Error("dispatch must not fire for non-y/n keys")
	}
	if !m.Active() {
		t.Error("modal should stay open after non-y/n keys")
	}
}

func TestModal_ErrMsgKeepsOpen(t *testing.T) {
	t.Parallel()
	m := NewModal(theme.Default())
	m.Open("Delete?", ModalOpts{Title: "Delete"})
	dispatch := func(bool) (string, tea.Cmd) { return "permission denied", nil }
	m.Update(keyChar('y'), dispatch)
	if !m.Active() {
		t.Error("non-empty errMsg should keep the modal open")
	}
	if m.Prompt() != "Delete?" || m.Title() != "Delete" {
		t.Error("prompt + title should be preserved while modal stays open")
	}
}

func TestModal_NonKeyMsgPassesThrough(t *testing.T) {
	t.Parallel()
	m := NewModal(theme.Default())
	m.Open("?", ModalOpts{})
	type dummy struct{}
	handled, _ := m.Update(dummy{}, nil)
	if handled {
		t.Error("non-KeyMsg should not be consumed when modal is active")
	}
}

func TestModal_NilDispatchClosesSafely(t *testing.T) {
	t.Parallel()
	m := NewModal(theme.Default())
	m.Open("ok?", ModalOpts{})
	m.Update(keyChar('y'), nil)
	if m.Active() {
		t.Error("nil dispatch should still close on y")
	}
}

func TestModal_ErrMsgShownInBox(t *testing.T) {
	t.Parallel()
	c := New(Config{Theme: theme.Default()})
	m := NewModal(theme.Default())
	m.Open("Delete?", ModalOpts{})
	m.Update(keyChar('y'), func(bool) (string, tea.Cmd) { return "permission denied", nil })
	if m.Err() != "permission denied" {
		t.Errorf("Err() = %q", m.Err())
	}
	if !strings.Contains(c.renderModalBox(m, 100), "permission denied") {
		t.Error("the dispatch error should be drawn inside the modal")
	}
}

func TestChrome_RenderModalOverContent(t *testing.T) {
	t.Parallel()
	c := New(Config{Theme: theme.Default()})
	m := NewModal(theme.Default())
	m.Open("Delete pods distribution/trader-tools?", ModalOpts{
		Title:  "Delete",
		Fields: []ModalField{SelectField("p", "Propagation", "Background"), CheckboxField("f", "Force", false)},
	})
	rows := make([]string, 30)
	for i := range rows {
		rows[i] = fmt.Sprintf("row%02d %s", i, strings.Repeat("x", 110))
	}
	content := c.BorderedContent(strings.Join(rows, "\n"), 120, 30)
	out := c.overlayModal(content, m)

	if lipgloss.Width(out) != lipgloss.Width(content) || lipgloss.Height(out) != lipgloss.Height(content) {
		t.Fatalf("overlay changed the content size: %dx%d -> %dx%d",
			lipgloss.Width(content), lipgloss.Height(content), lipgloss.Width(out), lipgloss.Height(out))
	}
	plain := xansi.Strip(out)
	for _, want := range []string{"<Delete>", "Delete pods distribution/trader-tools?", "Propagation: Background", "Force: [ ]", "Cancel", "OK"} {
		if !strings.Contains(plain, want) {
			t.Errorf("overlay missing %q", want)
		}
	}
	// The table stays visible around the box: the first and last rows
	// are untouched, and the rows beside the box keep their edges.
	for _, want := range []string{"row00 ", "row29 ", "row15 "} {
		if !strings.Contains(plain, want) {
			t.Errorf("content row %q should show around the modal", want)
		}
	}
	lines := strings.Split(plain, "\n")
	mid := lines[len(lines)/2]
	if !strings.Contains(mid, "xxx│") && !strings.Contains(mid, "xxx ") {
		t.Errorf("content should show on both sides of the box: %q", mid)
	}
}

func TestChrome_RenderDrawsModalOverContent(t *testing.T) {
	t.Parallel()
	c := New(Config{Theme: theme.Default()})
	m := NewModal(theme.Default())
	m.Open("Proceed?", ModalOpts{Title: "Confirm"})
	content := c.BorderedContent("behind the modal", 100, 20)
	out := xansi.Strip(c.Render(Frame{Width: 100, Height: 40, Content: content, Modal: m}))
	if !strings.Contains(out, "behind the modal") || !strings.Contains(out, "Proceed?") {
		t.Errorf("Render should draw the modal over Content, keeping Content visible:\n%s", out)
	}
}

func keyLeft() tea.KeyPressMsg  { return tea.KeyPressMsg{Code: tea.KeyLeft} }
func keyRight() tea.KeyPressMsg { return tea.KeyPressMsg{Code: tea.KeyRight} }
func keyShiftTab() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
}

// The buttons are k9s's: the focused one black on dodgerblue, the other
// in the dialog's cadetblue — not white on blue and gray.
func TestModal_ButtonColors(t *testing.T) {
	t.Parallel()
	c := New(Config{Theme: theme.Default()})
	m := NewModal(theme.Default())
	m.Open("Delete?", ModalOpts{})
	out := c.renderModalButtons(m, 40)
	for name, want := range map[string]string{
		"focused Cancel": "\x1b[38;2;0;0;0;48;2;30;144;255mCancel",
		"unfocused OK":   "\x1b[38;2;95;158;160;48;2;0;0;0mOK",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("%s: want %q in %q", name, want, out)
		}
	}

	// A skin's dialog colors reach the buttons.
	th := theme.Default()
	th.DialogButtonFocus = lipgloss.Color("#FF00FF")
	th.DialogText = lipgloss.Color("#00FF00")
	c = New(Config{Theme: th})
	out = c.renderModalButtons(m, 40)
	if !strings.Contains(out, "48;2;255;0;255mCancel") || !strings.Contains(out, "38;2;0;255;0;48;2;0;0;0mOK") {
		t.Errorf("theme dialog colors not used: %q", out)
	}
}
