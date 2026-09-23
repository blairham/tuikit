package chrome

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/blairham/tuikit/theme"
)

func TestConfirm_InitialState(t *testing.T) {
	t.Parallel()
	c := NewConfirm(theme.Default())
	if c.Active() {
		t.Error("new Confirm should be inactive")
	}
	if c.Prompt() != "" {
		t.Error("new Confirm should have empty prompt")
	}
}

func TestConfirm_OpenStoresPrompt(t *testing.T) {
	t.Parallel()
	c := NewConfirm(theme.Default())
	c.Open("Delete topic foo?")
	if !c.Active() {
		t.Error("Open should activate")
	}
	if c.Prompt() != "Delete topic foo?" {
		t.Errorf("Prompt() = %q; want %q", c.Prompt(), "Delete topic foo?")
	}
}

func TestConfirm_CloseClears(t *testing.T) {
	t.Parallel()
	c := NewConfirm(theme.Default())
	c.Open("question?")
	c.Close()
	if c.Active() || c.Prompt() != "" {
		t.Errorf("after Close: active=%v prompt=%q", c.Active(), c.Prompt())
	}
}

func TestConfirm_UpdateInactiveIsNoop(t *testing.T) {
	t.Parallel()
	c := NewConfirm(theme.Default())
	handled, cmd := c.Update(keyChar('y'), nil)
	if handled || cmd != nil {
		t.Errorf("inactive Update should be no-op; handled=%v cmd=%v", handled, cmd)
	}
}

func TestConfirm_KeysFireDispatchAndClose(t *testing.T) {
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
			c := NewConfirm(theme.Default())
			c.Open("Drop the table?")
			var got bool
			dispatch := func(yes bool) (string, tea.Cmd) {
				got = yes
				return "", nil
			}
			handled, _ := c.Update(tc.key, dispatch)
			if !handled {
				t.Errorf("%s: should be handled", tc.name)
			}
			if got != tc.want {
				t.Errorf("%s: dispatch received %v; want %v", tc.name, got, tc.want)
			}
			if c.Active() {
				t.Errorf("%s: should close after answer", tc.name)
			}
		})
	}
}

func TestConfirm_OtherKeysAreSwallowed(t *testing.T) {
	t.Parallel()
	c := NewConfirm(theme.Default())
	c.Open("are you sure?")
	called := false
	dispatch := func(bool) (string, tea.Cmd) {
		called = true
		return "", nil
	}
	for _, r := range "abcdefghijklmopqrstuvwxz" {
		handled, _ := c.Update(keyChar(r), dispatch)
		if !handled {
			t.Errorf("non y/n key %q should be swallowed (handled=true)", r)
		}
	}
	if called {
		t.Error("dispatch must not fire for non-y/n keys")
	}
	if !c.Active() {
		t.Error("bar should stay open after non-y/n keys")
	}
}

func TestConfirm_ErrMsgKeepsOpen(t *testing.T) {
	t.Parallel()
	c := NewConfirm(theme.Default())
	c.Open("Delete?")
	dispatch := func(bool) (string, tea.Cmd) { return "permission denied", nil }
	c.Update(keyChar('y'), dispatch)
	if !c.Active() {
		t.Error("non-empty errMsg should keep the bar open")
	}
	if c.Prompt() != "Delete?" {
		t.Error("prompt should be preserved while bar stays open")
	}
}

func TestConfirm_NonKeyMsgPassesThrough(t *testing.T) {
	t.Parallel()
	c := NewConfirm(theme.Default())
	c.Open("?")
	type dummy struct{}
	handled, _ := c.Update(dummy{}, nil)
	if handled {
		t.Error("non-KeyMsg should not be consumed when bar is active")
	}
}

func TestConfirm_NilDispatchClosesSafely(t *testing.T) {
	t.Parallel()
	c := NewConfirm(theme.Default())
	c.Open("ok?")
	c.Update(keyChar('y'), nil)
	if c.Active() {
		t.Error("nil dispatch should still close on y")
	}
}
