// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package chart

import (
	"image/color"
	"math"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/blairham/tuikit/theme"
)

// blocks are the partial cells, an eighth to seven eighths full; a full
// cell is fullBlock.
var blocks = [...]rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇'}

const (
	fullBlock = '█'
	eighths   = 8 // levels per cell
)

// Sparkline draws one or more [Series] as columns of block characters,
// newest sample on the right, k9s pulses style. Each row of its height is
// eight levels (▁▂▃▄▅▆▇█), so a 3-row sparkline draws 24. Build one with
// [NewSparkline], size it with [Sparkline.Resize] and render
// [Sparkline.View]; it reads its series afresh each time it draws.
type Sparkline struct {
	titleStyle lipgloss.Style
	labelStyle lipgloss.Style
	canvas     canvas
	title      string
	label      string
	series     []*Series
	colors     []color.Color
	max        float64
	width      int
	height     int
}

// NewSparkline returns a sparkline over series drawn in t's colors: series
// i in [theme.Theme.ChartColor](i), so the first two are always told
// apart. It scales to the data until [Sparkline.SetMax] fixes a max.
func NewSparkline(t theme.Theme, series ...*Series) *Sparkline {
	s := &Sparkline{
		canvas:     newCanvas(t),
		titleStyle: t.Title,
		labelStyle: t.On(t.ChartColor(0)).Bold(true),
	}
	s.colors = []color.Color{t.ChartColor(0), t.ChartColor(1)}
	s.SetSeries(series...)
	return s
}

// SetSeries replaces the series drawn. A nil series draws as an empty one.
// Where two series share a cell, the lower one is drawn in front and the
// higher one behind it, so both stay visible; within a cell both have
// only partly filled, the higher one's top is lost.
func (s *Sparkline) SetSeries(series ...*Series) { s.series = series }

// SetMax fixes the value drawn at full height, such as 100 for a
// percentage; samples above it are drawn at it. A max that is not a
// positive finite number (0, the default) scales to the data instead: to
// the largest finite sample on screen.
func (s *Sparkline) SetMax(m float64) {
	if m <= 0 || math.IsInf(m, 0) || math.IsNaN(m) {
		m = 0
	}
	s.max = m
}

// SetTitle sets the text at the left of the header line, such as "CPU".
func (s *Sparkline) SetTitle(title string) { s.title = plain(title) }

// SetLabel sets the text at the right of the header line, typically the
// current value formatted by the app, such as "12%".
func (s *Sparkline) SetLabel(label string) { s.label = plain(label) }

// Resize sets the size [Sparkline.View] draws at: width cells by height
// lines, the header included.
func (s *Sparkline) Resize(width, height int) { s.width, s.height = width, height }

// View draws the sparkline: exactly the height in lines, each exactly the
// width in cells. A title or label takes the first line when the height
// is 2 or more; the samples fill the rest. When a series holds more
// samples than there are columns the most recent are drawn; fewer are
// aligned right. Empty and all-zero series draw blank columns.
//
// It returns "" until [Sparkline.Resize] has given it a size.
func (s *Sparkline) View() string {
	if s.width <= 0 || s.height <= 0 {
		return ""
	}
	lines := make([]string, 0, s.height)
	rows := s.height
	if rows > 1 && (s.title != "" || s.label != "") {
		lines = append(lines, s.canvas.header(s.title, s.label, s.titleStyle, s.labelStyle, s.width))
		rows--
	}
	return strings.Join(append(lines, s.plot(rows)...), "\n")
}

// visible is each series' most recent samples, at most the width.
func (s *Sparkline) visible() [][]float64 {
	out := make([][]float64, len(s.series))
	for i, sr := range s.series {
		if sr == nil {
			continue
		}
		v := sr.Values()
		out[i] = v[max(len(v)-s.width, 0):]
	}
	return out
}

// scale is the value drawn at full height.
func (s *Sparkline) scale(vis [][]float64) float64 {
	if s.max > 0 {
		return s.max
	}
	top := 0.0
	for _, vs := range vis {
		for _, v := range vs {
			if !math.IsInf(v, 0) && v > top { // NaN compares false
				top = v
			}
		}
	}
	return top
}

// level is how many eighths of a cell v fills in a column of `levels`
// eighths with top drawn full: 0 for NaN, zero, negatives and -Inf; levels
// for +Inf and anything at or above top; otherwise v's share rounded, but
// never 0 for a positive v.
func level(v, top float64, levels int) int {
	switch {
	case math.IsNaN(v) || v <= 0:
		return 0
	case v >= top:
		return levels
	}
	return min(max(int(math.Round(v/top*float64(levels))), 1), levels)
}

// plot draws rows lines of columns.
func (s *Sparkline) plot(rows int) []string {
	vis := s.visible()
	top := s.scale(vis)
	levels := rows * eighths
	// lv[i][x] is series i's level in column x; right-aligned.
	lv := make([][]int, len(vis))
	for i, vs := range vis {
		lv[i] = make([]int, s.width)
		off := s.width - len(vs)
		for j, v := range vs {
			lv[i][off+j] = level(v, top, levels)
		}
	}
	lines := make([]string, rows)
	cells := make([]cell, s.width)
	for r := range rows {
		floor := (rows - 1 - r) * eighths // eighths below this line
		for x := range cells {
			cells[x] = s.cell(lv, x, floor)
		}
		lines[r] = s.canvas.line(cells)
	}
	return lines
}

// cell is column x's cell whose bottom sits floor eighths up: the
// least-filled series that reaches it in front, and a series that fills it
// completely behind.
func (s *Sparkline) cell(lv [][]int, x, floor int) cell {
	front, fill := -1, 0
	behind := -1
	for i := range lv {
		f := min(max(lv[i][x]-floor, 0), eighths)
		if f == 0 {
			continue
		}
		if front < 0 || f < fill {
			if front >= 0 && fill == eighths && behind < 0 {
				behind = front
			}
			front, fill = i, f
		} else if f == eighths && behind < 0 {
			behind = i
		}
	}
	if front < 0 {
		return cell{r: ' '}
	}
	c := cell{r: fullBlock, fg: s.color(front)}
	if fill < eighths {
		c.r = blocks[fill-1]
		if behind >= 0 {
			c.bg = s.color(behind)
		}
	}
	return c
}

// color is series i's color: the theme's two chart colors, alternating.
func (s *Sparkline) color(i int) color.Color { return s.colors[i%len(s.colors)] }
