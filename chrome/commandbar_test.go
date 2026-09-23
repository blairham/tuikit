package chrome

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/blairham/tuikit/theme"
)

func newBar(t *testing.T, opts CommandBarOpts) *CommandBar {
	t.Helper()
	return NewCommandBar(theme.Default(), opts)
}

func keyEsc() tea.KeyPressMsg   { return tea.KeyPressMsg{Code: tea.KeyEsc} }
func keyEnter() tea.KeyPressMsg { return tea.KeyPressMsg{Code: tea.KeyEnter} }
func keyChar(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

func TestCommandBar_InitialState(t *testing.T) {
	t.Parallel()
	bar := newBar(t, CommandBarOpts{})
	if bar.Active() {
		t.Error("new bar should be inactive")
	}
	if got := bar.Value(); got != "" {
		t.Errorf("new bar value = %q; want empty", got)
	}
	if got := bar.Error(); got != "" {
		t.Errorf("new bar error = %q; want empty", got)
	}
	if bar.Input() == nil {
		t.Error("Input() must not return nil")
	}
}

func TestCommandBar_OpenClearsError(t *testing.T) {
	t.Parallel()
	bar := newBar(t, CommandBarOpts{})
	bar.SetError("boom")
	if bar.Error() != "boom" {
		t.Fatalf("error not stored: %q", bar.Error())
	}
	bar.Open()
	if bar.Error() != "" {
		t.Errorf("Open should clear error; got %q", bar.Error())
	}
	if !bar.Active() {
		t.Error("Open should activate")
	}
}

func TestCommandBar_OpenWithPrefill(t *testing.T) {
	t.Parallel()
	bar := newBar(t, CommandBarOpts{})
	bar.OpenWith("topic foo")
	if !bar.Active() {
		t.Error("OpenWith should activate")
	}
	if got := bar.Value(); got != "topic foo" {
		t.Errorf("Value() = %q; want %q", got, "topic foo")
	}
}

func TestCommandBar_UpdateInactiveIsNoop(t *testing.T) {
	t.Parallel()
	bar := newBar(t, CommandBarOpts{})
	handled, cmd := bar.Update(keyChar('x'), nil)
	if handled {
		t.Error("Update on inactive bar should return handled=false")
	}
	if cmd != nil {
		t.Error("Update on inactive bar should not return a cmd")
	}
	if bar.Value() != "" {
		t.Errorf("inactive bar should not consume input; got %q", bar.Value())
	}
}

func TestCommandBar_EscClosesAndClears(t *testing.T) {
	t.Parallel()
	bar := newBar(t, CommandBarOpts{})
	bar.OpenWith("partial")
	bar.SetError("stale")

	handled, cmd := bar.Update(keyEsc(), nil)
	if !handled {
		t.Error("Esc should be handled when active")
	}
	if cmd != nil {
		t.Error("Esc should not return a cmd")
	}
	if bar.Active() {
		t.Error("Esc should deactivate the bar")
	}
	if bar.Value() != "" {
		t.Errorf("Esc should clear value; got %q", bar.Value())
	}
	if bar.Error() != "" {
		t.Errorf("Esc should clear error; got %q", bar.Error())
	}
}

func TestCommandBar_EnterEmptyValueClosesSilently(t *testing.T) {
	t.Parallel()
	bar := newBar(t, CommandBarOpts{})
	bar.Open()

	dispatchCalls := 0
	dispatch := func(string) (string, tea.Cmd) {
		dispatchCalls++
		return "", nil
	}
	handled, _ := bar.Update(keyEnter(), dispatch)
	if !handled {
		t.Error("Enter on empty value should be handled")
	}
	if dispatchCalls != 0 {
		t.Errorf("dispatch called %d times on empty Enter; want 0", dispatchCalls)
	}
	if bar.Active() {
		t.Error("Enter on empty value should close the bar")
	}
}

func TestCommandBar_EnterDispatchSuccessCloses(t *testing.T) {
	t.Parallel()
	bar := newBar(t, CommandBarOpts{})
	bar.OpenWith("quit")

	called := ""
	dispatch := func(v string) (string, tea.Cmd) {
		called = v
		return "", nil
	}
	handled, _ := bar.Update(keyEnter(), dispatch)
	if !handled {
		t.Error("Enter should be handled")
	}
	if called != "quit" {
		t.Errorf("dispatch received %q; want %q", called, "quit")
	}
	if bar.Active() {
		t.Error("successful dispatch should close the bar")
	}
}

func TestCommandBar_EnterDispatchErrorKeepsBarOpen(t *testing.T) {
	t.Parallel()
	bar := newBar(t, CommandBarOpts{})
	bar.OpenWith("bogus")

	dispatch := func(v string) (string, tea.Cmd) {
		return "unknown command: " + v, nil
	}
	handled, _ := bar.Update(keyEnter(), dispatch)
	if !handled {
		t.Error("Enter should be handled even on error")
	}
	if !bar.Active() {
		t.Error("error dispatch should keep the bar open")
	}
	if got := bar.Error(); got != "unknown command: bogus" {
		t.Errorf("Error() = %q; want unknown command: bogus", got)
	}
	if bar.Value() != "bogus" {
		t.Errorf("value should be preserved on error; got %q", bar.Value())
	}
}

func TestCommandBar_EnterTrimsWhitespace(t *testing.T) {
	t.Parallel()
	bar := newBar(t, CommandBarOpts{})
	bar.OpenWith("  spaced   ")

	var got string
	dispatch := func(v string) (string, tea.Cmd) {
		got = v
		return "", nil
	}
	_, _ = bar.Update(keyEnter(), dispatch)
	if got != "spaced" {
		t.Errorf("dispatch saw %q; want %q (trimmed)", got, "spaced")
	}
}

func TestCommandBar_EnterForwardsDispatchCmd(t *testing.T) {
	t.Parallel()
	bar := newBar(t, CommandBarOpts{})
	bar.OpenWith("go")
	sentinel := func() tea.Msg { return "done" }

	dispatch := func(string) (string, tea.Cmd) {
		return "", sentinel
	}
	_, gotCmd := bar.Update(keyEnter(), dispatch)
	if gotCmd == nil {
		t.Fatal("expected dispatch cmd to be returned")
	}
	if got, ok := gotCmd().(string); !ok || got != "done" {
		t.Errorf("dispatched cmd returned %v; want %q", gotCmd(), "done")
	}
}

func TestCommandBar_TypingForwardsToTextinput(t *testing.T) {
	t.Parallel()
	bar := newBar(t, CommandBarOpts{})
	bar.Open()

	for _, r := range "abc" {
		handled, _ := bar.Update(keyChar(r), nil)
		if !handled {
			t.Errorf("typing %q should be handled", r)
		}
	}
	if got := bar.Value(); got != "abc" {
		t.Errorf("Value() after typing abc = %q", got)
	}
}

func TestCommandBar_SuggestFnRefreshes(t *testing.T) {
	t.Parallel()
	calls := 0
	var lastValue string
	suggest := func(v string) []string {
		calls++
		lastValue = v
		return []string{v + "!"}
	}
	bar := newBar(t, CommandBarOpts{SuggestFn: suggest})

	bar.Open()
	if calls != 1 || lastValue != "" {
		t.Errorf("Open should seed suggestions; calls=%d lastValue=%q", calls, lastValue)
	}
	bar.Update(keyChar('a'), nil)
	if calls != 2 || lastValue != "a" {
		t.Errorf("keystroke should refresh suggestions; calls=%d lastValue=%q", calls, lastValue)
	}
	bar.Update(keyChar('b'), nil)
	if calls != 3 || lastValue != "ab" {
		t.Errorf("second keystroke should refresh suggestions; calls=%d lastValue=%q", calls, lastValue)
	}
}

func TestCommandBar_StaticSuggestionsEnableShow(t *testing.T) {
	t.Parallel()
	bar := newBar(t, CommandBarOpts{Suggestions: []string{"alpha", "beta"}})
	if !bar.Input().ShowSuggestions {
		t.Error("static Suggestions should enable ShowSuggestions on the textinput")
	}
}

func TestCommandBar_SetSuggestionsTogglesShow(t *testing.T) {
	t.Parallel()
	bar := newBar(t, CommandBarOpts{})
	if bar.Input().ShowSuggestions {
		t.Error("empty bar should not show suggestions")
	}
	bar.SetSuggestions([]string{"x", "y"})
	if !bar.Input().ShowSuggestions {
		t.Error("SetSuggestions with values should enable ShowSuggestions")
	}
	bar.SetSuggestions(nil)
	if bar.Input().ShowSuggestions {
		t.Error("SetSuggestions(nil) should disable ShowSuggestions")
	}
}

func TestCommandBar_NonKeyMsgForwardsUnhandled(t *testing.T) {
	t.Parallel()
	bar := newBar(t, CommandBarOpts{})
	bar.Open()

	type dummyMsg struct{}
	handled, _ := bar.Update(dummyMsg{}, nil)
	if handled {
		t.Error("non-KeyMsg should pass through with handled=false so the app router still sees it")
	}
}

func TestCommandBar_PromptDefault(t *testing.T) {
	t.Parallel()
	bar := newBar(t, CommandBarOpts{})
	if got := bar.Input().Prompt; got != ": " {
		t.Errorf("default prompt = %q; want %q", got, ": ")
	}
}

func TestCommandBar_PromptOverride(t *testing.T) {
	t.Parallel()
	bar := newBar(t, CommandBarOpts{Prompt: ">"})
	if got := bar.Input().Prompt; got != "> " {
		t.Errorf("custom prompt = %q; want %q", got, "> ")
	}
}
