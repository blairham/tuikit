package table

import (
	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"

	"github.com/blairham/tuikit/theme"
)

// KeyMap returns the default bubbles table keymap with j/k removed.
// tuikit apps handle j/k at the global level (via [viewfsm.Router] or
// equivalent) and re-feed arrow keys into the table, so the table's own
// j/k bindings would double-trigger.
func KeyMap() table.KeyMap {
	km := table.DefaultKeyMap()
	km.LineUp.SetKeys("up")
	km.LineDown.SetKeys("down")
	return km
}

// Styles returns themed bubbles table styles. Honors
// [theme.Theme.PaintBackground] — when true (default), cell and header
// backgrounds are explicitly Bg so cell padding renders on the canvas
// color instead of the terminal default. When false, no Background is
// set, which lets Selected.Foreground overrides take effect cleanly
// (the PaintModeNone model).
func Styles(t theme.Theme) table.Styles {
	s := table.DefaultStyles()

	header := lipgloss.NewStyle().
		Foreground(t.Value).
		Bold(true).
		Padding(0, 1)
	cell := lipgloss.NewStyle().
		Foreground(t.Selection).
		Padding(0, 1)

	if t.PaintBackground {
		header = header.Background(t.Bg)
		cell = cell.Background(t.Bg)
	}

	s.Header = header
	s.Cell = cell
	s.Selected = lipgloss.NewStyle().
		Background(t.Selection).
		Bold(true)

	return s
}

// StylesWithWidth returns [Styles] with the Selected style padded to
// the full table width so the highlight stretches edge-to-edge. Call
// with the inner panel width whenever the panel resizes.
func StylesWithWidth(t theme.Theme, width int) table.Styles {
	s := Styles(t)
	if width > 0 {
		s.Selected = s.Selected.Width(width).MaxWidth(width).Inline(true)
	}
	return s
}

// FitHeight returns the height to pass to a bubbles table given the
// number of data rows and the maximum height available. The +1 accounts
// for the header row.
func FitHeight(rows, maxHeight int) int {
	desired := rows + 1
	if maxHeight > 0 && desired > maxHeight {
		desired = maxHeight
	}
	if desired < 1 {
		return 1
	}
	return desired
}

// Truncate shortens s to maxLen runes, appending "…" if it was cut.
func Truncate(s string, maxLen int) string {
	if maxLen < 1 || len(s) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return "…"
	}
	return s[:maxLen-1] + "…"
}
