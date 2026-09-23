package chrome

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

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
		want bool
	}{
		{name: "lowercase y", key: keyChar('y'), want: true},
		{name: "uppercase Y", key: keyChar('Y'), want: true},
		{name: "enter", key: keyEnter(), want: true},
		{name: "lowercase n", key: keyChar('n'), want: false},
		{name: "uppercase N", key: keyChar('N'), want: false},
		{name: "esc", key: keyEsc(), want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := NewModal(theme.Default())
			m.Open("Drop the table?", ModalOpts{})
			var got bool
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

func TestChrome_RenderModalContent(t *testing.T) {
	t.Parallel()
	c := New(Config{Theme: theme.Default()})
	m := NewModal(theme.Default())
	m.Open("Session expired — refresh your login?", ModalOpts{
		Title:       "Re-authenticate",
		OkLabel:     "Login",
		CancelLabel: "Skip",
	})
	out := c.renderModalContent(Frame{Modal: m, Width: 120, Height: 30})
	if !strings.Contains(out, "Re-authenticate") {
		t.Error("rendered modal should embed title in border")
	}
	if !strings.Contains(out, "Session expired") {
		t.Error("rendered modal should embed prompt body")
	}
	if !strings.Contains(out, "Login") || !strings.Contains(out, "Skip") {
		t.Error("rendered modal should include both button labels")
	}
}
