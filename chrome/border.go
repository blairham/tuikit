package chrome

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/blairham/tuikit/theme"
)

// InjectBorderTitle rewrites the top border line of a lipgloss box so
// that `title` appears centered in it, producing the k9s-style
// border:
//
//	╭── streams(all)[12] ──╮
//	│ ...                  │
//	╰──────────────────────╯
//
// The title can carry its own ANSI styling (typical: theme.Title +
// theme.AccentAlt for the bracketed sub-label). The border characters
// are repainted in theme.Border foreground.
//
// If the input box has no top line (single-line input), the function
// returns it unchanged.
func InjectBorderTitle(box, title string, t theme.Theme) string {
	lines := strings.SplitN(box, "\n", 2)
	if len(lines) < 1 {
		return box
	}

	topWidth := lipgloss.Width(lines[0])
	titleWidth := lipgloss.Width(title) + 2 // +2 for the framing spaces

	leftPad := (topWidth - titleWidth - 2) / 2 // -2 for the corner chars
	rightPad := topWidth - titleWidth - 2 - leftPad
	if leftPad < 1 {
		leftPad = 1
	}
	if rightPad < 1 {
		rightPad = 1
	}

	borderStyle := lipgloss.NewStyle().Foreground(t.Border)
	spacerStyle := lipgloss.NewStyle()
	if t.PaintBackground {
		borderStyle = borderStyle.Background(t.Bg)
		spacerStyle = spacerStyle.Background(t.Bg)
	}
	spacer := spacerStyle.Render(" ")

	newTop := borderStyle.Render("╭"+strings.Repeat("─", leftPad)) +
		spacer + title + spacer +
		borderStyle.Render(strings.Repeat("─", rightPad)+"╮")

	if len(lines) > 1 {
		return newTop + "\n" + lines[1]
	}
	return newTop
}

// WrapText is a minimal word-wrap that respects an ANSI-free string.
// Inputs are typically status / error messages — for styled content
// callers should pre-render then size separately.
func WrapText(s string, width int) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}
	var lines []string
	cur := words[0]
	for _, w := range words[1:] {
		if len(cur)+1+len(w) > width {
			lines = append(lines, cur)
			cur = w
			continue
		}
		cur += " " + w
	}
	lines = append(lines, cur)
	return strings.Join(lines, "\n ")
}

// CenterInBox renders content centered horizontally and vertically
// inside a box of the given dimensions, painted with the theme's
// muted foreground over the canvas background. Used as a placeholder
// when a view has no data yet.
func CenterInBox(content string, w, h int, t theme.Theme) string {
	s := lipgloss.NewStyle().
		Width(w).
		Height(h).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(t.Muted)
	if t.PaintBackground {
		s = s.Background(t.Bg)
	}
	return s.Render(content)
}
