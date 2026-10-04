// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package chart

import (
	"image/color"
	"math"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/blairham/tuikit/theme"
)

// barEighths are the partial bar cells, an eighth to seven eighths full,
// filled from the left.
var barEighths = [...]rune{'▏', '▎', '▍', '▌', '▋', '▊', '▉'}

// track is the unfilled part of a gauge's bar.
const track = '─'

// Level is the state a [Gauge]'s thresholds put its value in.
type Level int

// The levels, in rising severity.
const (
	LevelOK       Level = iota // no threshold crossed: the theme's first chart color
	LevelWarn                  // the warn threshold crossed: the theme's warn color
	LevelCritical              // the critical threshold crossed: the theme's error color
)

// Gauge draws one value against a total as a horizontal bar, such as pods
// running out of pods wanted, with a label ("3/7 running"). Warn and
// critical thresholds switch the bar and label to the theme's warn and
// error colors. Build one with [NewGauge], size it with [Gauge.Resize] and
// render [Gauge.View].
type Gauge struct {
	titleStyle lipgloss.Style
	colors     [3]color.Color // by Level
	trackColor color.Color
	canvas     canvas
	title      string
	label      string
	value      float64
	total      float64
	warn       float64
	critical   float64
	width      int
	height     int
	low        bool
}

// NewGauge returns an empty gauge drawn in t's colors, with no thresholds.
func NewGauge(t theme.Theme) *Gauge {
	return &Gauge{
		colors:     [3]color.Color{t.ChartColor(0), t.Status.Warn, t.Status.Error},
		canvas:     newCanvas(t),
		trackColor: t.Muted,
		titleStyle: t.Title,
		warn:       math.NaN(),
		critical:   math.NaN(),
	}
}

// Set sets the value and the total it is drawn against. The bar is
// value/total of the width: a value that is NaN, negative or -Inf draws
// empty, and one at or above the total (+Inf included) draws full. A
// total that is not a positive finite number draws an empty bar at
// [LevelOK].
func (g *Gauge) Set(value, total float64) { g.value, g.total = value, total }

// SetTitle sets the text at the left of the header line, such as "Pods".
func (g *Gauge) SetTitle(title string) { g.title = plain(title) }

// SetLabel sets the value's label, such as "3/7 running". Empty, the
// default, draws the value and total as "value/total".
func (g *Gauge) SetLabel(label string) { g.label = plain(label) }

// SetHighThresholds makes a high value the bad one, as for CPU: the gauge
// is at [LevelWarn] when value/total is at least warn and at
// [LevelCritical] when it is at least critical. Both are fractions of the
// total; a NaN one is never crossed.
func (g *Gauge) SetHighThresholds(warn, critical float64) {
	g.warn, g.critical, g.low = warn, critical, false
}

// SetLowThresholds makes a low value the bad one, as for pods running out
// of pods wanted: the gauge is at [LevelWarn] when value/total is below
// warn and at [LevelCritical] when it is below critical. Both are
// fractions of the total, so SetLowThresholds(1, 0.5) warns whenever any
// are missing and is critical when fewer than half are there. A NaN one is
// never crossed.
func (g *Gauge) SetLowThresholds(warn, critical float64) {
	g.warn, g.critical, g.low = warn, critical, true
}

// ClearThresholds removes the thresholds: the gauge stays at [LevelOK].
func (g *Gauge) ClearThresholds() { g.warn, g.critical = math.NaN(), math.NaN() }

// Ratio is the share of the bar filled, from 0 to 1.
func (g *Gauge) Ratio() float64 {
	switch {
	case g.total <= 0 || math.IsInf(g.total, 0) || math.IsNaN(g.total):
		return 0
	case math.IsNaN(g.value) || g.value <= 0:
		return 0
	case g.value >= g.total:
		return 1
	}
	return g.value / g.total
}

// Level is the state the thresholds put the value in. A gauge with no
// total to measure against is at [LevelOK].
func (g *Gauge) Level() Level {
	if g.total <= 0 || math.IsInf(g.total, 0) || math.IsNaN(g.total) {
		return LevelOK
	}
	r := g.Ratio()
	crossed := func(th float64) bool {
		if g.low {
			return r < th
		}
		return r >= th
	}
	switch {
	case crossed(g.critical):
		return LevelCritical
	case crossed(g.warn):
		return LevelWarn
	}
	return LevelOK
}

// Resize sets the size [Gauge.View] draws at: width cells by height lines,
// the header included.
func (g *Gauge) Resize(width, height int) { g.width, g.height = width, height }

// labelText is the label drawn: the one set, or "value/total".
func (g *Gauge) labelText() string {
	if g.label != "" {
		return g.label
	}
	return format(g.value) + "/" + format(g.total)
}

func format(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

// View draws the gauge: exactly the height in lines, each exactly the width
// in cells. With a height of 2 or more the first line is the header — the
// title on the left, the label on the right in the level's color — and
// every line after it is the bar. With a height of 1 the line is the
// title, the bar and the label, dropping the title and then the label
// when the width is too small for them and a bar.
//
// It returns "" until [Gauge.Resize] has given it a size.
func (g *Gauge) View() string {
	if g.width <= 0 || g.height <= 0 {
		return ""
	}
	style := g.canvas.style(g.colors[g.Level()], nil).Bold(true)
	label := g.labelText()
	if g.height == 1 {
		return g.inline(label, style)
	}
	lines := make([]string, g.height)
	lines[0] = g.canvas.header(g.title, label, g.titleStyle, style, g.width)
	bar := g.bar(g.width)
	for i := 1; i < len(lines); i++ {
		lines[i] = bar
	}
	return strings.Join(lines, "\n")
}

// inline is the one-line form: "title bar label".
func (g *Gauge) inline(label string, style lipgloss.Style) string {
	title := g.title
	if title != "" && width(title)+1+width(label)+1+1 > g.width {
		title = ""
	}
	if width(title)+1+width(label)+1 > g.width {
		label = ""
	}
	var b strings.Builder
	n := g.width
	if title != "" {
		b.WriteString(g.titleStyle.Render(title))
		b.WriteString(g.canvas.spaces(1))
		n -= width(title) + 1
	}
	if label != "" {
		n -= width(label) + 1
	}
	b.WriteString(g.bar(n))
	if label != "" {
		b.WriteString(g.canvas.spaces(1))
		b.WriteString(style.Render(label))
	}
	return b.String()
}

// bar is the bar n cells wide, filled to the ratio in eighths of a cell. A
// positive ratio fills at least an eighth, so it never looks empty.
func (g *Gauge) bar(n int) string {
	if n <= 0 {
		return ""
	}
	r := g.Ratio()
	e := int(math.Round(r * float64(n*eighths)))
	if r > 0 && e == 0 {
		e = 1
	}
	fg := g.colors[g.Level()]
	cells := make([]cell, n)
	for x := range cells {
		switch fill := e - x*eighths; {
		case fill >= eighths:
			cells[x] = cell{r: fullBlock, fg: fg}
		case fill > 0:
			cells[x] = cell{r: barEighths[fill-1], fg: fg}
		default:
			cells[x] = cell{r: track, fg: g.trackColor}
		}
	}
	return g.canvas.line(cells)
}
