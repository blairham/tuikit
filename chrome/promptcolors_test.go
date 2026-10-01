package chrome

import (
	"image/color"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/blairham/tuikit/theme"
)

// fgSeq is the SGR a foreground color renders as under lipgloss.
func fgSeq(c color.Color) string {
	s := lipgloss.NewStyle().Foreground(c).Render("x")
	return s[2:strings.IndexByte(s, 'm')]
}

// TestCommandBarUsesK9sPromptColors renders a command bar mid-completion
// and checks each part against k9s's default skin: aqua border, cadetblue
// typed text, dodgerblue suggestion. The border must not follow Accent —
// an app recoloring its accent used to recolor the bar.
func TestCommandBarUsesK9sPromptColors(t *testing.T) {
	th := theme.Default()
	th.Accent = lipgloss.Color("#2496ED") // an app's own accent
	th.Rebuild()
	bar := NewCommandBar(th, CommandBarOpts{Prompt: ":", Suggestions: []string{"containers"}})
	bar.Open()
	for _, r := range "co" {
		bar.Update(tea.KeyPressMsg{Code: r, Text: string(r)}, func(string) (string, tea.Cmd) { return "", nil })
	}
	out := New(Config{Theme: th}).renderCommandBar(bar.Input(), 60)

	for part, c := range map[string]color.Color{
		"border (aqua)":           th.CommandBorder,
		"typed text (cadetblue)":  th.InputText,
		"suggestion (dodgerblue)": th.Suggestion,
	} {
		if !strings.Contains(out, fgSeq(c)) {
			t.Errorf("command bar has no %s: %q", part, out)
		}
	}
	if strings.Contains(out, fgSeq(th.Accent)) {
		t.Error("command bar still draws in the app's accent color")
	}
	for name, got := range map[string]color.Color{
		"CommandBorder": th.CommandBorder, "InputText": th.InputText, "Suggestion": th.Suggestion,
	} {
		want := map[string]string{"CommandBorder": "#00FFFF", "InputText": "#5F9EA0", "Suggestion": "#1E90FF"}[name]
		if fgSeq(got) != fgSeq(lipgloss.Color(want)) {
			t.Errorf("%s default is not k9s's %s", name, want)
		}
	}
}
