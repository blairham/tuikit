// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package chrome

import (
	"fmt"
	"slices"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/blairham/tuikit/theme"
)

// These tests drive the bars the way an app does: Open on the trigger
// key, then every key press through Update.

func keyUp() tea.KeyPressMsg   { return tea.KeyPressMsg{Code: tea.KeyUp} }
func keyDown() tea.KeyPressMsg { return tea.KeyPressMsg{Code: tea.KeyDown} }
func keyTab() tea.KeyPressMsg  { return tea.KeyPressMsg{Code: tea.KeyTab} }
func keyCtrl(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Mod: tea.ModCtrl}
}

func okDispatch(string) (string, tea.Cmd) { return "", nil }

// submit opens the bar, types s and presses enter.
func submit(c *CommandBar, s string) {
	c.Open()
	typeInto(c, s)
	c.Update(keyEnter(), okDispatch)
}

// press sends k through Update and returns the bar's value afterward.
func press(c *CommandBar, k tea.KeyPressMsg) string {
	c.Update(k, okDispatch)
	return c.Value()
}

func newHistoryBar(t *testing.T, entries ...string) *CommandBar {
	t.Helper()
	c := NewCommandBar(theme.Default(), CommandBarOpts{})
	for _, e := range entries {
		submit(c, e)
	}
	return c
}

func TestCommandBar_HistoryRecallOrder(t *testing.T) {
	t.Parallel()
	c := newHistoryBar(t, "pods", "deploy", "svc")
	c.Open()

	steps := []struct {
		want string
		key  tea.KeyPressMsg
	}{
		{key: keyUp(), want: "svc"},
		{key: keyUp(), want: "deploy"},
		{key: keyUp(), want: "pods"},
		{key: keyUp(), want: "pods"}, // oldest: stays put
		{key: keyDown(), want: "deploy"},
		{key: keyDown(), want: "svc"},
		{key: keyDown(), want: ""}, // past the newest: the (empty) draft
		{key: keyDown(), want: ""}, // not recalling: nothing to do
		{key: keyUp(), want: "svc"},
	}
	for i, s := range steps {
		if got := press(c, s.key); got != s.want {
			t.Fatalf("step %d (%s): value = %q, want %q", i, s.key, got, s.want)
		}
	}
	if pos, n := c.Input().Position(), len([]rune(c.Value())); pos != n {
		t.Errorf("cursor at %d after recall, want the end (%d)", pos, n)
	}
}

func TestCommandBar_HistoryDraftRestored(t *testing.T) {
	t.Parallel()
	c := newHistoryBar(t, "pods", "deploy")
	c.Open()
	typeInto(c, "no")
	if got := press(c, keyUp()); got != "deploy" {
		t.Fatalf("up = %q, want deploy", got)
	}
	if got := press(c, keyUp()); got != "pods" {
		t.Fatalf("up = %q, want pods", got)
	}
	press(c, keyDown())
	if got := press(c, keyDown()); got != "no" {
		t.Errorf("down past the newest = %q, want the draft %q", got, "no")
	}
}

func TestCommandBar_HistoryRecording(t *testing.T) {
	t.Parallel()
	c := newHistoryBar(t, "a", "a", "b", "a", "", "   ", "  c  ")
	// Consecutive repeats collapse; a repeat of an older entry does not.
	// Empty and blank submissions are skipped; values are trimmed.
	want := []string{"a", "b", "a", "c"}
	if got := c.History(); !slices.Equal(got, want) {
		t.Errorf("History() = %q, want %q", got, want)
	}
}

func TestCommandBar_HistoryRecordsDispatchErrors(t *testing.T) {
	t.Parallel()
	c := NewCommandBar(theme.Default(), CommandBarOpts{})
	c.Open()
	typeInto(c, "typo")
	c.Update(keyEnter(), func(string) (string, tea.Cmd) { return "unknown command", nil })
	if !c.Active() {
		t.Fatal("dispatch error should keep the bar open")
	}
	if got := c.History(); !slices.Equal(got, []string{"typo"}) {
		t.Errorf("History() = %q, want the failed command recorded", got)
	}
}

func TestCommandBar_HistoryCap(t *testing.T) {
	t.Parallel()
	c := NewCommandBar(theme.Default(), CommandBarOpts{})
	for i := range HistoryLimit + 5 {
		submit(c, fmt.Sprintf("cmd%d", i))
	}
	got := c.History()
	if len(got) != HistoryLimit {
		t.Fatalf("len(History()) = %d, want %d", len(got), HistoryLimit)
	}
	if got[0] != "cmd5" || got[len(got)-1] != fmt.Sprintf("cmd%d", HistoryLimit+4) {
		t.Errorf("History() runs %q..%q, want the newest %d kept", got[0], got[len(got)-1], HistoryLimit)
	}
}

func TestCommandBar_HistoryCancelDoesNotRecord(t *testing.T) {
	t.Parallel()
	c := newHistoryBar(t, "pods")
	c.Open()
	typeInto(c, "abandoned")
	c.Update(keyEsc(), okDispatch)
	if got := c.History(); !slices.Equal(got, []string{"pods"}) {
		t.Errorf("History() = %q after esc, want only the submitted command", got)
	}
}

func TestCommandBar_HistoryEditRecalled(t *testing.T) {
	t.Parallel()
	c := newHistoryBar(t, "pods")
	c.Open()
	press(c, keyUp())
	typeInto(c, " -A")
	if got := c.Value(); got != "pods -A" {
		t.Fatalf("edited value = %q, want %q", got, "pods -A")
	}
	c.Update(keyEnter(), okDispatch)
	if got := c.History(); !slices.Equal(got, []string{"pods", "pods -A"}) {
		t.Errorf("History() = %q, want the edited command appended", got)
	}
}

// TestCommandBar_HistoryOpenResets: reopening mid-recall starts again
// from the newest entry with a fresh draft.
func TestCommandBar_HistoryOpenResets(t *testing.T) {
	t.Parallel()
	c := newHistoryBar(t, "pods", "deploy")
	c.Open()
	press(c, keyUp())
	press(c, keyUp()) // at "pods"
	c.OpenWith("")    // reopened while still active
	if got := press(c, keyUp()); got != "deploy" {
		t.Errorf("up after reopening = %q, want the newest entry", got)
	}
	if got := press(c, keyDown()); got != "" {
		t.Errorf("down past the newest after reopening = %q, want the new (empty) draft", got)
	}
}

// TestCommandBar_HistorySubmitResets: a dispatch error keeps the bar
// open, and the next up starts from the newest entry again.
func TestCommandBar_HistorySubmitResets(t *testing.T) {
	t.Parallel()
	c := newHistoryBar(t, "pods", "deploy")
	fail := func(string) (string, tea.Cmd) { return "nope", nil }
	c.Open()
	press(c, keyUp())
	press(c, keyUp()) // at "pods"
	typeInto(c, "x")
	c.Update(keyEnter(), fail)
	if got := press(c, keyUp()); got != "podsx" {
		t.Fatalf("first up after submit = %q, want the just-submitted podsx", got)
	}
	if got := press(c, keyUp()); got != "deploy" {
		t.Errorf("second up after submit = %q, want deploy", got)
	}
}

func TestCommandBar_HistoryEmptyUpDownConsumed(t *testing.T) {
	t.Parallel()
	c := NewCommandBar(theme.Default(), CommandBarOpts{})
	c.Open()
	typeInto(c, "po")
	for _, k := range []tea.KeyPressMsg{keyUp(), keyDown()} {
		handled, _ := c.Update(k, okDispatch)
		if !handled {
			t.Errorf("%s: handled = false, want the bar to consume it", k)
		}
		if got := c.Value(); got != "po" {
			t.Errorf("%s with no history changed the value to %q", k, got)
		}
	}
}

// TestCommandBar_SuggestionsBesideHistory: tab and → still accept a
// suggestion, ctrl+n / ctrl+p still cycle them, and up/down no longer do.
func TestCommandBar_SuggestionsBesideHistory(t *testing.T) {
	t.Parallel()
	open := func() *CommandBar {
		c := NewCommandBar(theme.Default(), CommandBarOpts{Suggestions: []string{"pods", "podsecurity", "pv"}})
		c.Open()
		typeInto(c, "po")
		return c
	}

	c := open()
	if got := press(c, keyTab()); got != "pods" {
		t.Errorf("tab = %q, want the first suggestion accepted", got)
	}
	c = open()
	if got := press(c, tea.KeyPressMsg{Code: tea.KeyRight}); got != "pods" {
		t.Errorf("→ = %q, want the first suggestion accepted", got)
	}

	c = open()
	press(c, keyCtrl('n'))
	if got := c.Input().CurrentSuggestion(); got != "podsecurity" {
		t.Errorf("ctrl+n: current suggestion = %q, want podsecurity", got)
	}
	press(c, keyCtrl('p'))
	if got := c.Input().CurrentSuggestion(); got != "pods" {
		t.Errorf("ctrl+p: current suggestion = %q, want pods", got)
	}

	c = open()
	press(c, keyDown())
	if got := c.Input().CurrentSuggestion(); got != "pods" {
		t.Errorf("down cycled suggestions to %q; it belongs to history", got)
	}
}

// TestCommandBar_HistoryRefreshesSuggestFn: a recalled value gets
// suggestions for itself, not for the draft it replaced.
func TestCommandBar_HistoryRefreshesSuggestFn(t *testing.T) {
	t.Parallel()
	all := []string{"pods", "deploy"}
	c := NewCommandBar(theme.Default(), CommandBarOpts{SuggestFn: func(string) []string { return all }})
	c.SetHistory([]string{"dep"})
	c.Open()
	typeInto(c, "po")
	press(c, keyUp())
	if got := c.Input().CurrentSuggestion(); got != "deploy" {
		t.Errorf("suggestion after recalling %q = %q, want deploy", c.Value(), got)
	}
}

func TestCommandBar_SetHistoryRoundTrip(t *testing.T) {
	t.Parallel()
	c := NewCommandBar(theme.Default(), CommandBarOpts{})
	seed := []string{"pods", "", "deploy", "deploy", "svc"}
	c.SetHistory(seed)
	want := []string{"pods", "deploy", "svc"}
	if got := c.History(); !slices.Equal(got, want) {
		t.Fatalf("History() = %q, want %q", got, want)
	}

	seed[0] = "mutated"
	got := c.History()
	got[1] = "mutated"
	if again := c.History(); !slices.Equal(again, want) {
		t.Errorf("History() = %q after mutating the seed and a returned slice, want %q", again, want)
	}

	c.Open()
	if v := press(c, keyUp()); v != "svc" {
		t.Errorf("up after SetHistory = %q, want svc", v)
	}

	long := make([]string, HistoryLimit+10)
	for i := range long {
		long[i] = fmt.Sprintf("c%d", i)
	}
	c.SetHistory(long)
	if got := c.History(); len(got) != HistoryLimit || got[0] != "c10" {
		t.Errorf("SetHistory over the cap kept %d entries from %q, want the newest %d", len(got), got[0], HistoryLimit)
	}
}

// --- FilterBar ---

func commitFilter(f *FilterBar, s string) {
	f.OpenWith("")
	for _, r := range s {
		f.Update(keyChar(r), nil)
	}
	f.Update(keyEnter(), nil)
}

func TestFilterBar_HistoryRecallAndDraft(t *testing.T) {
	t.Parallel()
	f := newFilter(t)
	commitFilter(f, "err")
	commitFilter(f, "warn")

	var seen []string
	onFilter := func(v string) { seen = append(seen, v) }

	f.OpenWith("ti")
	steps := []struct {
		want string
		key  tea.KeyPressMsg
	}{
		{key: keyUp(), want: "warn"},
		{key: keyUp(), want: "err"},
		{key: keyUp(), want: "err"},
		{key: keyDown(), want: "warn"},
		{key: keyDown(), want: "ti"},
	}
	for i, s := range steps {
		f.Update(s.key, onFilter)
		if got := f.Value(); got != s.want {
			t.Fatalf("step %d (%s): value = %q, want %q", i, s.key, got, s.want)
		}
	}
	// The view re-filters on every recall that changes the value.
	if want := []string{"warn", "err", "warn", "ti"}; !slices.Equal(seen, want) {
		t.Errorf("onFilter calls = %q, want %q", seen, want)
	}
}

func TestFilterBar_HistoryRecording(t *testing.T) {
	t.Parallel()
	f := newFilter(t)
	commitFilter(f, "err")
	commitFilter(f, "err")
	commitFilter(f, "")
	f.OpenWith("")
	f.Update(keyChar('x'), nil)
	f.Update(keyEsc(), nil) // canceled: not recorded
	if got := f.History(); !slices.Equal(got, []string{"err"}) {
		t.Errorf("History() = %q, want only the committed filter once", got)
	}
}

func TestFilterBar_HistoryEditAndReset(t *testing.T) {
	t.Parallel()
	f := newFilter(t)
	f.SetHistory([]string{"err", "warn"})

	f.Open()
	f.Update(keyUp(), nil)
	f.Update(keyUp(), nil) // at "err"
	f.Open()               // reopened mid-recall
	f.Update(keyUp(), nil)
	if got := f.Value(); got != "warn" {
		t.Fatalf("up after reopening = %q, want the newest entry", got)
	}
	f.Update(keyChar('!'), nil)
	f.Update(keyEnter(), nil)
	if got := f.History(); !slices.Equal(got, []string{"err", "warn", "warn!"}) {
		t.Errorf("History() = %q, want the edited filter appended", got)
	}

	got := f.History()
	got[0] = "mutated"
	if again := f.History(); again[0] != "err" {
		t.Errorf("History() returned a shared slice: %q", again)
	}
}
