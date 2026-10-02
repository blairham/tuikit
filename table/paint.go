package table

import (
	"regexp"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/blairham/tuikit/theme"
)

// PaintMode selects which background-fix branches [FixSelectedRow]
// performs. Use [PaintModeFor] to derive it from a [theme.Theme].
type PaintMode int

// Paint modes.
const (
	// PaintModeNone disables every transformation. Use when the theme
	// already gives bubbles table the styles it expects.
	PaintModeNone PaintMode = iota
	// PaintModeFull rewrites cell-padding backgrounds AND repaints
	// non-selected rows so mid-row resets don't bleed the terminal
	// default through. Match for [theme.Theme.PaintBackground] = true.
	PaintModeFull
)

// PaintModeFor returns the [PaintMode] that fits the given theme.
func PaintModeFor(t theme.Theme) PaintMode {
	if t.PaintBackground {
		return PaintModeFull
	}
	return PaintModeNone
}

// selectedBgMarker is the truecolor bg ANSI emitted by the Selected
// style; blackBgMarker is what Cell.Background = #000000 writes.
const (
	selectedBgMarker = "48;2;135;206;250"
	blackBgMarker    = "48;2;0;0;0"
)

var (
	resetRe  = regexp.MustCompile(`\x1b\[m`)
	fgCodeRe = regexp.MustCompile(`38;2;\d+;\d+;\d+|38;5;\d+`)
)

// FixSelectedRow post-processes a bubbles table's rendered output so:
//
//   - Selected rows read as black text on light-sky-blue: inner cell
//     backgrounds (Cell.Background = #000) are rewritten to the
//     selected bg, foregrounds to black, and mid-row resets are
//     replaced with reset+reapply-selected-bg so the selection
//     survives per-cell ANSI resets.
//
//   - Non-selected rows keep a continuous black background across cell
//     padding: mid-row resets are replaced with reset+reapply-black-bg
//     so styled inner spans (status badges, side glyphs, …) don't drop
//     the outer Cell.Background between cells. This pass is only run
//     under [PaintModeFull]; under [PaintModeNone] the chrome doesn't
//     paint backgrounds at all so non-selected lines pass through
//     unchanged.
//
// In both cases the trailing reset is preserved so line endings stay
// clean and don't bleed style onto subsequent lines.
//
// The selected-row fix runs under [PaintModeNone] too: bubbles
// table's `Cell.Render` emits `\x1b[m` between every cell, which
// clears the row-level Selected background after the first column
// and drops cell foregrounds back to their pre-Selected color. Without
// this rewrite, PaintModeNone themes show the highlight
// only on the first column and read the rest of the row in the
// pre-Selected cell-foreground color (typically invisible against
// the highlight if their colors collide).
//
// It assumes the default theme's colors (light-sky-blue selection, black
// canvas and selected text). A theme with others — a skin, or
// [theme.Theme.Inverted] — needs [FixRows].
func FixSelectedRow(view string, mode PaintMode) string {
	return painter{selBg: selectedBgMarker, selFg: "38;2;0;0;0", bg: blackBgMarker}.fix(view, mode)
}

// FixRows is [FixSelectedRow] with the theme's own colors: the selected
// row is found by [theme.Theme.Selection] and drawn in
// [theme.Theme.SelectionTextColor], and when the theme paints its canvas
// the other rows keep [theme.Theme.Bg] across cell resets.
func FixRows(view string, t theme.Theme) string {
	p := painter{
		selBg: sgrParams(theme.BackgroundSeq(t.Selection)),
		selFg: sgrParams(lipgloss.NewStyle().Foreground(t.SelectionTextColor()).Render(" ")),
		bg:    sgrParams(theme.BackgroundSeq(t.Bg)),
	}
	if p.selBg == "" || p.selFg == "" || p.bg == "" {
		return view
	}
	return p.fix(view, PaintModeFor(t))
}

// sgrParams is the parameters of the SGR sequence s starts with:
// "48;2;135;206;250" for "\x1b[48;2;135;206;250m …". "" if s starts with
// none.
func sgrParams(s string) string {
	rest, ok := strings.CutPrefix(s, "\x1b[")
	if !ok {
		return ""
	}
	params, _, ok := strings.Cut(rest, "m")
	if !ok {
		return ""
	}
	return params
}

// painter repaints a table with one set of colors, as SGR parameters: the
// selection's background and text, and the canvas.
type painter struct {
	selBg, selFg, bg string
}

func (p painter) fix(view string, mode PaintMode) string {
	lines := strings.Split(view, "\n")
	for i, line := range lines {
		switch {
		case strings.Contains(line, p.selBg):
			lines[i] = p.selectedLine(line)
		case mode == PaintModeFull:
			lines[i] = p.paintLine(line)
		}
	}
	return strings.Join(lines, "\n")
}

func (p painter) selectedLine(line string) string {
	line = strings.ReplaceAll(line, p.bg, p.selBg)
	line = fgCodeRe.ReplaceAllString(line, p.selFg)
	reapply := "\x1b[m\x1b[1;" + p.selFg + ";" + p.selBg + "m"
	line = resetRe.ReplaceAllString(line, reapply)
	if idx := strings.LastIndex(line, reapply); idx >= 0 {
		line = line[:idx] + "\x1b[m"
	}
	return line
}

func (p painter) paintLine(line string) string {
	// Fast path: pure-text lines (no ANSI) need no rewrite.
	if !strings.Contains(line, "\x1b[") {
		return line
	}
	return theme.ReassertBackground(line, "\x1b["+p.bg+"m")
}
