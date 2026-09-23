package chrome

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/blairham/tuikit/theme"
)

func newFilter(t *testing.T) *FilterBar {
	t.Helper()
	return NewFilterBar(theme.Default(), FilterBarOpts{Placeholder: "type to filter"})
}

func TestFilterBar_InitialState(t *testing.T) {
	t.Parallel()
	f := newFilter(t)
	if f.Active() {
		t.Error("new FilterBar should be inactive")
	}
	if f.Value() != "" {
		t.Error("new FilterBar should have empty value")
	}
}

func TestFilterBar_OpenKeepsExistingValue(t *testing.T) {
	t.Parallel()
	f := newFilter(t)
	f.SetValue("preserved")
	f.Open()
	if !f.Active() {
		t.Error("Open should activate")
	}
	if f.Value() != "preserved" {
		t.Errorf("Open should not clear value; got %q", f.Value())
	}
}

func TestFilterBar_OpenWithReplacesValue(t *testing.T) {
	t.Parallel()
	f := newFilter(t)
	f.SetValue("old")
	f.OpenWith("new")
	if f.Value() != "new" {
		t.Errorf("OpenWith should replace; got %q", f.Value())
	}
}

func TestFilterBar_TypingInvokesOnFilter(t *testing.T) {
	t.Parallel()
	f := newFilter(t)
	f.Open()

	var got []string
	onFilter := func(v string) { got = append(got, v) }
	for _, r := range "abc" {
		f.Update(keyChar(r), onFilter)
	}
	if len(got) != 3 || got[0] != "a" || got[1] != "ab" || got[2] != "abc" {
		t.Errorf("onFilter sequence = %v; want [a ab abc]", got)
	}
}

func TestFilterBar_EnterClosesKeepsValue(t *testing.T) {
	t.Parallel()
	f := newFilter(t)
	f.OpenWith("matched")

	calls := 0
	handled, _ := f.Update(keyEnter(), func(string) { calls++ })
	if !handled {
		t.Error("Enter should be handled")
	}
	if f.Active() {
		t.Error("Enter should close")
	}
	if f.Value() != "matched" {
		t.Errorf("Enter should keep value; got %q", f.Value())
	}
	if calls != 0 {
		t.Error("Enter should not re-invoke onFilter (live filter already applied)")
	}
}

func TestFilterBar_EscClosesClearsAndNotifies(t *testing.T) {
	t.Parallel()
	f := newFilter(t)
	f.OpenWith("doomed")

	var got string
	onFilter := func(v string) { got = v }
	handled, _ := f.Update(keyEsc(), onFilter)
	if !handled {
		t.Error("Esc should be handled")
	}
	if f.Active() || f.Value() != "" {
		t.Errorf("Esc should clear; active=%v value=%q", f.Active(), f.Value())
	}
	if got != "" {
		t.Errorf("Esc should notify with empty; got %q", got)
	}
}

func TestFilterBar_UpdateInactiveIsNoop(t *testing.T) {
	t.Parallel()
	f := newFilter(t)
	calls := 0
	handled, _ := f.Update(keyChar('a'), func(string) { calls++ })
	if handled || calls != 0 {
		t.Errorf("inactive Update should be no-op; handled=%v calls=%d", handled, calls)
	}
}

func TestFilterBar_NonKeyMsgPassesThrough(t *testing.T) {
	t.Parallel()
	f := newFilter(t)
	f.Open()
	type dummy struct{}
	handled, _ := f.Update(dummy{}, nil)
	if handled {
		t.Error("non-KeyMsg should pass through with handled=false")
	}
}

func TestFilterBar_NilOnFilterIsSafe(t *testing.T) {
	t.Parallel()
	f := newFilter(t)
	f.Open()
	// Should not panic on any keystroke when onFilter is nil.
	f.Update(keyChar('x'), nil)
	f.Update(keyEnter(), nil)
	f.Open()
	f.Update(keyEsc(), nil)
}

func TestFilterBar_ClearWithoutOpen(t *testing.T) {
	t.Parallel()
	f := newFilter(t)
	f.SetValue("stale")
	f.Clear()
	if f.Active() || f.Value() != "" {
		t.Errorf("Clear should empty value and stay inactive; active=%v value=%q", f.Active(), f.Value())
	}
}

// Compile-time check that FilterBar's Update signature matches what apps
// expect: a key-aware Update that takes an OnFilter callback.
var _ func(tea.Msg, OnFilter) (bool, tea.Cmd) = (*FilterBar)(nil).Update
