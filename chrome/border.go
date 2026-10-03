// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package chrome

import (
	"image/color"
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
// are repainted in the focused-border color ([theme.Theme.FocusBorder]),
// matching [theme.Theme.TableBorder] — this is the main content box's
// title. A box drawn in another color (a modal, the help overlay, a
// secondary pane) uses [InjectBorderTitleColor] with that color, so the
// top line does not change color where it meets the side borders.
//
// If the input box has no top line (single-line input), the function
// returns it unchanged.
func InjectBorderTitle(box, title string, t theme.Theme) string {
	return InjectBorderTitleColor(box, title, t.FocusBorder(), t)
}

// InjectBorderTitleColor is [InjectBorderTitle] with the repainted top
// line drawn in border. Pass the same color the box's side borders use.
func InjectBorderTitleColor(box, title string, border color.Color, t theme.Theme) string {
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

	borderStyle := lipgloss.NewStyle().Foreground(border)
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
