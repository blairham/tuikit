// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package chart

import (
	"math"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/blairham/tuikit/theme"
)

// SGR fragments for the default theme's chart colors.
const (
	paleGreenFg = "38;2;152;251;152"
	paleGreenBg = "48;2;152;251;152"
	orangeRedFg = "38;2;255;69;0"
	yellowFg    = "38;2;255;255;0"
	redFg       = "38;2;255;0;0"
)

// series returns a series holding vs.
func series(vs ...float64) *Series {
	s := NewSeries(max(len(vs), 1))
	for _, v := range vs {
		s.Push(v)
	}
	return s
}

// spark draws series at w by h with the given fixed max (0 for auto) and
// returns the view's lines, ANSI stripped.
func spark(t *testing.T, top float64, w, h int, ss ...*Series) []string {
	t.Helper()
	s := NewSparkline(theme.Default(), ss...)
	s.SetMax(top)
	s.Resize(w, h)
	return stripped(t, s.View(), w, h)
}

// stripped splits a view into lines, checks it is exactly w by h, and
// returns the lines with ANSI removed.
func stripped(t *testing.T, view string, w, h int) []string {
	t.Helper()
	lines := strings.Split(ansi.Strip(view), "\n")
	if len(lines) != h {
		t.Fatalf("view is %d lines, want %d:\n%s", len(lines), h, view)
	}
	for i, l := range lines {
		if got := ansi.StringWidth(l); got != w {
			t.Fatalf("line %d is %d cells, want %d: %q", i, got, w, l)
		}
	}
	return lines
}

func equalLines(t *testing.T, got []string, want ...string) {
	t.Helper()
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("got\n%q\nwant\n%q", got, want)
	}
}

func TestSparklineHeightOneBlocks(t *testing.T) {
	got := spark(t, 8, 9, 1, series(0, 1, 2, 3, 4, 5, 6, 7, 8))
	equalLines(t, got, " ▁▂▃▄▅▆▇█")
}

func TestSparklineHeightThreeStacks(t *testing.T) {
	// 24 levels: 1 is ▁ on the bottom row, 9 fills the bottom and starts
	// the middle, 17 starts the top.
	got := spark(t, 24, 8, 3, series(0, 1, 8, 9, 12, 16, 17, 24))
	equalLines(t, got,
		"      ▁█",
		"   ▁▄███",
		" ▁██████",
	)
}

func TestSparklineScale(t *testing.T) {
	for _, tc := range []struct {
		name string
		want string
		in   []float64
		top  float64
	}{
		{name: "auto scales to the largest", in: []float64{1, 2, 4}, top: 0, want: "▂▄█"},
		{name: "fixed max", in: []float64{1, 2, 4}, top: 8, want: "▁▂▄"},
		{name: "above the fixed max is drawn at it", in: []float64{1, 4}, top: 2, want: "▄█"},
		{name: "a bad max scales to the data", in: []float64{1, 2}, top: math.NaN(), want: "▄█"},
		{name: "an infinite max scales to the data", in: []float64{1, 2}, top: math.Inf(1), want: "▄█"},
		{name: "a negative max scales to the data", in: []float64{1, 2}, top: -5, want: "▄█"},
		{name: "a tiny positive value still shows", in: []float64{0.001, 100}, top: 100, want: "▁█"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			equalLines(t, spark(t, tc.top, len([]rune(tc.want)), 1, series(tc.in...)), tc.want)
		})
	}
}

func TestSparklineEmptyAndZero(t *testing.T) {
	for _, tc := range []struct {
		s    *Series
		name string
	}{
		{name: "no samples", s: NewSeries(4)},
		{name: "all zero", s: series(0, 0, 0)},
		{name: "nil series", s: nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, top := range []float64{0, 100} {
				equalLines(t, spark(t, top, 4, 2, tc.s), "    ", "    ")
			}
		})
	}
	t.Run("no series at all", func(t *testing.T) {
		equalLines(t, spark(t, 0, 3, 1), "   ")
	})
}

func TestSparklineSingleSample(t *testing.T) {
	equalLines(t, spark(t, 0, 4, 1, series(3)), "   █")
	equalLines(t, spark(t, 6, 4, 1, series(3)), "   ▄")
}

func TestSparklineShowsMostRecent(t *testing.T) {
	// Only the newest three samples are drawn.
	equalLines(t, spark(t, 8, 3, 1, series(8, 8, 0, 0, 4)), "  ▄")
	// And auto scale reads only what is on screen: the 100 has scrolled
	// off, so 2 is drawn full.
	equalLines(t, spark(t, 0, 2, 1, series(100, 1, 2)), "▄█")
}

func TestSparklineAlignsRight(t *testing.T) {
	equalLines(t, spark(t, 2, 5, 1, series(1, 2)), "   ▄█")
}

func TestSparklineSpecialValues(t *testing.T) {
	// NaN, negatives and -Inf draw empty; +Inf draws full and does not set
	// the scale, which comes from the 8.
	equalLines(t, spark(t, 0, 6, 1, series(math.NaN(), -1, math.Inf(-1), math.Inf(1), 4, 8)), "   █▄█")
	// Huge values scale like any other.
	equalLines(t, spark(t, 0, 2, 1, series(math.MaxFloat64/2, math.MaxFloat64)), "▄█")
}

func TestSparklineSeriesColors(t *testing.T) {
	s := NewSparkline(theme.Default(), series(8), series(4))
	s.Resize(1, 1)
	// Series 1 is lower, so it is drawn in front, in its own color, with
	// series 0 filling the rest of the cell behind it.
	got := s.View()
	if ansi.Strip(got) != "▄" {
		t.Fatalf("cell = %q, want ▄", ansi.Strip(got))
	}
	if !strings.Contains(got, orangeRedFg) || !strings.Contains(got, paleGreenBg) {
		t.Errorf("cell = %q, want series 1's color %s in front of series 0's %s", got, orangeRedFg, paleGreenBg)
	}

	s.SetSeries(series(8), series(0))
	got = s.View()
	if ansi.Strip(got) != "█" || !strings.Contains(got, paleGreenFg) || strings.Contains(got, orangeRedFg) {
		t.Errorf("series 0 alone = %q, want a full block in %s", got, paleGreenFg)
	}

	// Two partly filled series: the lower one in front, on the canvas.
	s.SetSeries(series(2, 8), series(6, 8))
	s.SetMax(8)
	s.Resize(2, 1)
	got = s.View()
	if !strings.HasPrefix(ansi.Strip(got), "▂") || !strings.Contains(got, paleGreenFg+";48;2;0;0;0m▂") {
		t.Errorf("view = %q, want series 0's ▂ in front on the canvas", got)
	}
}

func TestSparklineHeader(t *testing.T) {
	s := NewSparkline(theme.Default(), series(1, 2))
	s.SetTitle("CPU")
	s.SetLabel("12%")
	s.SetMax(2)
	s.Resize(10, 2)
	equalLines(t, stripped(t, s.View(), 10, 2), "CPU    12%", "        ▄█")

	// Too narrow for both: the title gives way first.
	s.Resize(5, 2)
	equalLines(t, stripped(t, s.View(), 5, 2), "… 12%", "   ▄█")
	s.Resize(2, 2)
	equalLines(t, stripped(t, s.View(), 2, 2), "1…", "▄█")

	// A height of 1 is all chart.
	s.Resize(4, 1)
	equalLines(t, stripped(t, s.View(), 4, 1), "  ▄█")

	// Titles and labels are drawn as one line of plain text.
	s.SetTitle("\x1b[31mC\nP\tU\x1b[m")
	s.SetLabel("")
	s.Resize(6, 2)
	equalLines(t, stripped(t, s.View(), 6, 2), "C P U ", "    ▄█")
}

func TestSparklineNoPaint(t *testing.T) {
	for _, tc := range []struct {
		th    theme.Theme
		paint bool
	}{{th: theme.Default(), paint: true}, {th: theme.NoPaintBackground(), paint: false}} {
		s := NewSparkline(tc.th, series(0, 1))
		s.SetTitle("x")
		s.Resize(4, 2)
		if got := strings.Contains(s.View(), "48;2;0;0;0"); got != tc.paint {
			t.Errorf("PaintBackground %v: canvas painted = %v in %q", tc.paint, got, s.View())
		}
	}
}

func TestSparklineUnsized(t *testing.T) {
	s := NewSparkline(theme.Default(), series(1))
	if got := s.View(); got != "" {
		t.Errorf("View before Resize = %q, want empty", got)
	}
}

func TestDimensions(t *testing.T) {
	g := NewGauge(theme.Default())
	g.Set(5, 9)
	g.SetTitle("日本語のタイトル")
	sp := NewSparkline(theme.Default(), series(1, 5, 3), series(2))
	sp.SetTitle("memory")
	sp.SetLabel("1.2 GiB")
	grid := NewGrid(theme.Default(), 2, NewSparkline(theme.Default(), series(1)), NewGauge(theme.Default()), nil)
	for _, w := range []Widget{g, sp, grid} {
		for _, width := range []int{1, 2, 3, 7, 40} {
			for _, height := range []int{1, 2, 3, 5} {
				w.Resize(width, height)
				stripped(t, w.View(), width, height)
			}
		}
	}
}
