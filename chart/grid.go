// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package chart

import (
	"strings"

	"github.com/blairham/tuikit/theme"
)

// Widget is anything a [Grid] lays out: [Sparkline], [Gauge] and [Grid]
// itself.
type Widget interface {
	Resize(width, height int)
	View() string
}

// Grid lays widgets out in rows of a fixed number of columns, left to
// right and top to bottom, sharing its size between them evenly: the
// first columns and rows take the cells left over from an uneven split.
// Columns are separated by a gap of blank cells (1 by default); rows touch,
// since each widget's header already sets it apart. Build one with
// [NewGrid].
type Grid struct {
	canvas  canvas
	widgets []Widget
	columns int
	gap     int
	width   int
	height  int
}

// NewGrid returns a grid of widgets in the given number of columns (at
// least 1), its blank cells painted in t's background.
func NewGrid(t theme.Theme, columns int, widgets ...Widget) *Grid {
	return &Grid{canvas: newCanvas(t), widgets: widgets, columns: max(columns, 1), gap: 1}
}

// SetGap sets the blank cells between columns; negative is taken as 0. The
// gap shrinks when the width is too small for it and a cell per column.
func (g *Grid) SetGap(n int) { g.gap = max(n, 0) }

// Resize sets the size [Grid.View] draws at and resizes every widget to
// its cell.
func (g *Grid) Resize(width, height int) {
	g.width, g.height = width, height
	ws, hs, _ := g.split()
	for i, w := range g.widgets {
		if w != nil {
			w.Resize(ws[i%g.columns], hs[i/g.columns])
		}
	}
}

// split is each column's width, each row's height and the gap between
// columns.
func (g *Grid) split() (ws, hs []int, gap int) {
	cols := g.columns
	rows := max((len(g.widgets)+cols-1)/cols, 1)
	gap = g.gap
	if g.width < cols+gap*(cols-1) {
		gap = max((g.width-cols)/max(cols-1, 1), 0)
	}
	return share(max(g.width-gap*(cols-1), 0), cols), share(max(g.height, 0), rows), gap
}

// share splits total into n parts that differ by at most one, the larger
// first.
func share(total, n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = total / n
		if i < total%n {
			out[i]++
		}
	}
	return out
}

// View draws every widget in its cell: exactly the height in lines, each
// exactly the width in cells. A widget's view is cut or padded to its
// cell, so one that draws the wrong size cannot push the others out of
// line, and an empty slot in the last row is blank.
//
// It returns "" until [Grid.Resize] has given it a size.
func (g *Grid) View() string {
	if g.width <= 0 || g.height <= 0 {
		return ""
	}
	ws, hs, gap := g.split()
	lines := make([]string, 0, g.height)
	for r, h := range hs {
		if h == 0 {
			continue
		}
		row := make([]strings.Builder, h)
		for c, w := range ws {
			cell := g.cellLines(r*g.columns+c, w, h)
			for i := range row {
				if c > 0 {
					row[i].WriteString(g.canvas.spaces(gap))
				}
				row[i].WriteString(cell[i])
			}
		}
		for i := range row {
			lines = append(lines, row[i].String())
		}
	}
	return strings.Join(lines, "\n")
}

// cellLines is widget i drawn to exactly w by h, or blank when there is no
// widget i.
func (g *Grid) cellLines(i, w, h int) []string {
	s := ""
	if i < len(g.widgets) && g.widgets[i] != nil && w > 0 {
		s = g.widgets[i].View()
	}
	return g.canvas.fit(s, w, h)
}
