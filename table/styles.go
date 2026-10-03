// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package table

import (
	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/blairham/tuikit/theme"
)

// KeyMap returns the default bubbles table keymap with j/k removed from
// LineUp/LineDown, deliberately: tuikit apps handle j/k at the global
// level and translate them to arrow keys with [viewfsm.TranslateNavKey]
// before feeding the table, so the table's own j/k bindings would
// double-trigger (and would collide with app bindings for apps that
// claim j/k for something else).
//
// A table built with this keymap therefore does nothing with a literal
// j or k. Apps that want vim-style line movement must call
// [viewfsm.TranslateNavKey] (or equivalent) first; apps that do not
// should not advertise j/k in their help. The LineUp/LineDown help text
// reads "↑"/"↓" to match.
func KeyMap() table.KeyMap {
	km := table.DefaultKeyMap()
	km.LineUp.SetKeys("up")
	km.LineUp.SetHelp("↑", "up")
	km.LineDown.SetKeys("down")
	km.LineDown.SetHelp("↓", "down")
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
		Foreground(t.TableHeaderColor()).
		Bold(true).
		Padding(0, 1)
	cell := lipgloss.NewStyle().
		Foreground(t.TableTextColor()).
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

// Truncate shortens s to at most maxLen terminal cells, replacing the
// cut-off tail with "…" (which costs one cell). Width is measured in
// display cells, not bytes or runes: a wide rune (CJK, emoji) costs two,
// a cut never splits a rune or grapheme cluster, and ANSI escape
// sequences are preserved without counting toward the width. A maxLen
// below 1 means "no limit" and returns s unchanged.
func Truncate(s string, maxLen int) string {
	if maxLen < 1 {
		return s
	}
	return ansi.Truncate(s, maxLen, "…")
}
