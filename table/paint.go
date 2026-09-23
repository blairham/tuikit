package table

import (
	"regexp"
	"strings"

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
func FixSelectedRow(view string, mode PaintMode) string {
	lines := strings.Split(view, "\n")
	for i, line := range lines {
		switch {
		case strings.Contains(line, selectedBgMarker):
			lines[i] = fixSelectedLine(line)
		case mode == PaintModeFull:
			lines[i] = paintCellBackgrounds(line)
		}
	}
	return strings.Join(lines, "\n")
}

func fixSelectedLine(line string) string {
	line = strings.ReplaceAll(line, blackBgMarker, selectedBgMarker)
	line = fgCodeRe.ReplaceAllString(line, "38;2;0;0;0")
	reapply := "\x1b[m\x1b[1;38;2;0;0;0;" + selectedBgMarker + "m"
	line = resetRe.ReplaceAllString(line, reapply)
	if idx := strings.LastIndex(line, reapply); idx >= 0 {
		line = line[:idx] + "\x1b[m"
	}
	return line
}

func paintCellBackgrounds(line string) string {
	// Fast path: pure-text lines (no ANSI) need no rewrite.
	if !strings.Contains(line, "\x1b[") {
		return line
	}
	reapply := "\x1b[m\x1b[" + blackBgMarker + "m"
	line = resetRe.ReplaceAllString(line, reapply)
	if idx := strings.LastIndex(line, reapply); idx >= 0 {
		line = line[:idx] + "\x1b[m"
	}
	return line
}
