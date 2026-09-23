package chrome

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/blairham/tuikit/theme"
)

func newPrompt(t *testing.T, opts PromptOpts) *Prompt {
	t.Helper()
	return NewPrompt(theme.Default(), opts)
}

func TestPrompt_InitialState(t *testing.T) {
	t.Parallel()
	p := newPrompt(t, PromptOpts{})
	if p.Active() || p.Value() != "" || p.Error() != "" {
		t.Errorf("new Prompt should be zero state; got active=%v value=%q err=%q", p.Active(), p.Value(), p.Error())
	}
	if p.Input() == nil {
		t.Error("Input() must not return nil")
	}
}

func TestPrompt_OpenWithPrefill(t *testing.T) {
	t.Parallel()
	p := newPrompt(t, PromptOpts{})
	p.Open("retention.ms=86400000", OpenOpts{})
	if !p.Active() {
		t.Error("Open should activate")
	}
	if p.Value() != "retention.ms=86400000" {
		t.Errorf("Value() = %q; want prefill", p.Value())
	}
}

func TestPrompt_OpenSelectAllMutesPrefill(t *testing.T) {
	t.Parallel()
	p := newPrompt(t, PromptOpts{})
	p.Open("1h", OpenOpts{SelectAll: true})
	if p.Value() != "" {
		t.Errorf("SelectAll should leave value empty; got %q", p.Value())
	}
	if p.Input().Placeholder != "1h" {
		t.Errorf("SelectAll should show prefill in placeholder slot; got %q", p.Input().Placeholder)
	}
}

func TestPrompt_SelectAllPlaceholderClearsOnFirstKeystroke(t *testing.T) {
	t.Parallel()
	p := newPrompt(t, PromptOpts{})
	p.Open("1h", OpenOpts{SelectAll: true})
	p.Update(keyChar('5'), nil)
	if p.Input().Placeholder != "" {
		t.Errorf("first keystroke should drop the SelectAll placeholder; got %q", p.Input().Placeholder)
	}
}

func TestPrompt_OpenWithCustomPromptLabel(t *testing.T) {
	t.Parallel()
	p := newPrompt(t, PromptOpts{Prompt: "rename to"})
	p.Open("old", OpenOpts{Prompt: "save to"})
	if p.Input().Prompt != "save to " {
		t.Errorf("OpenOpts.Prompt should override; got %q", p.Input().Prompt)
	}
}

func TestPrompt_EscClosesAndClears(t *testing.T) {
	t.Parallel()
	p := newPrompt(t, PromptOpts{})
	p.Open("preset", OpenOpts{})
	p.SetError("stale")
	handled, _ := p.Update(keyEsc(), nil)
	if !handled {
		t.Error("Esc should be handled")
	}
	if p.Active() || p.Value() != "" || p.Error() != "" {
		t.Errorf("Esc should clear all state; got active=%v value=%q err=%q", p.Active(), p.Value(), p.Error())
	}
}

func TestPrompt_EnterEmptyCloses(t *testing.T) {
	t.Parallel()
	p := newPrompt(t, PromptOpts{})
	p.Open("", OpenOpts{})
	called := false
	dispatch := func(string) (string, tea.Cmd) {
		called = true
		return "", nil
	}
	p.Update(keyEnter(), dispatch)
	if called {
		t.Error("Empty value should not invoke dispatch")
	}
	if p.Active() {
		t.Error("Empty Enter should close the bar")
	}
}

func TestPrompt_EnterDispatchesValue(t *testing.T) {
	t.Parallel()
	p := newPrompt(t, PromptOpts{})
	p.Open("seed", OpenOpts{})
	var got string
	dispatch := func(v string) (string, tea.Cmd) {
		got = v
		return "", nil
	}
	p.Update(keyEnter(), dispatch)
	if got != "seed" {
		t.Errorf("dispatch saw %q; want seed", got)
	}
	if p.Active() {
		t.Error("successful dispatch should close")
	}
}

func TestPrompt_DispatchErrorKeepsOpen(t *testing.T) {
	t.Parallel()
	p := newPrompt(t, PromptOpts{})
	p.Open("bogus", OpenOpts{})
	dispatch := func(string) (string, tea.Cmd) { return "invalid duration", nil }
	p.Update(keyEnter(), dispatch)
	if !p.Active() {
		t.Error("error should keep prompt open")
	}
	if p.Value() != "bogus" {
		t.Errorf("value should be preserved on error; got %q", p.Value())
	}
	if p.Error() != "invalid duration" {
		t.Errorf("Error() = %q", p.Error())
	}
}

func TestPrompt_InactiveUpdateIsNoop(t *testing.T) {
	t.Parallel()
	p := newPrompt(t, PromptOpts{})
	handled, _ := p.Update(keyChar('a'), nil)
	if handled {
		t.Error("inactive Update should be no-op")
	}
}

func TestPrompt_DefaultPrompt(t *testing.T) {
	t.Parallel()
	p := newPrompt(t, PromptOpts{})
	if p.Input().Prompt != "> " {
		t.Errorf("default prompt = %q; want \"> \"", p.Input().Prompt)
	}
}
