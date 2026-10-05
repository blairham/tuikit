// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package tree

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/blairham/tuikit/theme"
)

// Markers drawn before the label of a node with children.
const (
	markerOpen   = "▾ "
	markerClosed = "▸ "
	ellipsis     = "…"
	reset        = "\x1b[m"
)

// styles is what View paints with, derived from the theme once.
type styles struct {
	guide    lipgloss.Style // guides and markers
	selected lipgloss.Style // the row under the cursor
	textOn   string         // SGR that opens a label: text color, and Bg when painting
	bgOn     string         // SGR that opens the canvas; "" when not painting
}

func newStyles(t theme.Theme) styles {
	s := styles{
		guide: t.On(t.XrayGraphicColor()),
		selected: lipgloss.NewStyle().
			Background(t.XrayCursorColor()).
			Foreground(t.XrayCursorTextColor()).
			Bold(true),
		textOn: openSeq(t.On(t.XrayTextColor())),
	}
	if t.PaintBackground {
		s.bgOn = theme.BackgroundSeq(t.Bg)
	}
	return s
}

// openSeq is the SGR sequence st opens a span with: what it writes before
// the text.
func openSeq(st lipgloss.Style) string {
	out := st.Render("x")
	if i := strings.IndexByte(out, 'x'); i > 0 {
		return out[:i]
	}
	return ""
}

// View draws the rows on screen, height lines of at most width cells: the
// guides, a ▾ (open) or ▸ (closed) marker on a node with children, and the
// label. Colored labels keep their color, with the theme's text color and
// background put back after each reset inside them. The row under the
// cursor is drawn across the full width in the theme's selection colors,
// as a table draws its selected row; its label's own colors give way to
// the selection's. Labels too wide for the row are cut with "…". Line
// breaks and tabs in a label are drawn as spaces, so a row is one line.
//
// It returns "" until [Model.Resize] has given it a size.
func (m *Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	lines := make([]string, m.height)
	for i := range lines {
		idx := m.offset + i
		switch {
		case idx >= len(m.rows):
			lines[i] = m.blank()
		case idx == m.cursor:
			lines[i] = m.selectedLine(m.rows[idx])
		default:
			lines[i] = m.line(m.rows[idx])
		}
	}
	return strings.Join(lines, "\n")
}

// prefix is a row's guides and marker.
func prefix(r row) string {
	switch {
	case !r.kids:
		return r.guide
	case r.open:
		return r.guide + markerOpen
	default:
		return r.guide + markerClosed
	}
}

// oneLine draws line breaks and tabs as spaces.
func oneLine(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '\n', '\r', '\t', '\v', '\f':
			return ' '
		}
		return r
	}, s)
}

func (m *Model) line(r row) string {
	text := oneLine(r.node.Label)
	out := m.compose(r, text)
	if cells(out) > m.width {
		// An escape sequence the label leaves unfinished can swallow the
		// "…" and the text after the cut, so the cut line measures wider
		// than it was cut to. Such a label is drawn without its styling.
		out = m.compose(r, ansi.Strip(text))
	}
	return m.pad(out)
}

// compose is a row's guides and label, cut to the width.
func (m *Model) compose(r row, label string) string {
	label = theme.ReassertBackground(m.styles.textOn+label+reset, m.styles.textOn)
	out := ansi.Truncate(m.styles.guide.Render(prefix(r))+label, m.width, ellipsis)
	if !strings.HasSuffix(out, reset) {
		out += reset
	}
	return out
}

// cells is how wide s is drawn: its width with escape sequences removed.
func cells(s string) int { return ansi.StringWidth(ansi.Strip(s)) }

func (m *Model) selectedLine(r row) string {
	text := ansi.Truncate(prefix(r)+ansi.Strip(oneLine(r.node.Label)), m.width, ellipsis)
	if w := cells(text); w < m.width {
		text += strings.Repeat(" ", m.width-w)
	}
	return m.styles.selected.Render(text)
}

// pad fills the rest of a painted line with the canvas.
func (m *Model) pad(s string) string {
	if m.styles.bgOn == "" {
		return s
	}
	if w := cells(s); w < m.width {
		s += m.styles.bgOn + strings.Repeat(" ", m.width-w) + reset
	}
	return s
}

func (m *Model) blank() string { return m.pad("") }
