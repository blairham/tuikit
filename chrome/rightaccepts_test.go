package chrome

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/blairham/tuikit/theme"
)

func typeInto(c *CommandBar, s string) {
	for _, r := range s {
		c.Update(tea.KeyPressMsg{Code: r, Text: string(r)}, nil)
	}
}

// TestRightArrowAcceptsSuggestion: → at the end of the input accepts the
// suggestion like tab, as k9s does; with nothing to accept, or mid-text, it
// is an ordinary cursor move.
func TestRightArrowAcceptsSuggestion(t *testing.T) {
	right := tea.KeyPressMsg{Code: tea.KeyRight}

	c := NewCommandBar(theme.Default(), CommandBarOpts{Prompt: ":", Suggestions: []string{"containers"}})
	c.Open()
	typeInto(c, "co")
	c.Update(right, nil)
	if got := c.Input().Value(); got != "containers" {
		t.Errorf("→ at the end = %q, want the suggestion accepted", got)
	}

	c = NewCommandBar(theme.Default(), CommandBarOpts{Prompt: ":", Suggestions: []string{"containers"}})
	c.Open()
	typeInto(c, "co")
	c.Update(tea.KeyPressMsg{Code: tea.KeyLeft}, nil)
	c.Update(right, nil)
	if got, pos := c.Input().Value(), c.Input().Position(); got != "co" || pos != 2 {
		t.Errorf("→ mid-text = %q at %d, want a cursor move only", got, pos)
	}

	c = NewCommandBar(theme.Default(), CommandBarOpts{Prompt: ":", Suggestions: []string{"containers"}})
	c.Open()
	typeInto(c, "zz")
	c.Update(right, nil)
	if got := c.Input().Value(); got != "zz" {
		t.Errorf("→ with no suggestion changed the value to %q", got)
	}
}
