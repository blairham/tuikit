// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package chart

import (
	"encoding/binary"
	"math"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/blairham/tuikit/theme"
)

// floats reads b as little-endian float64s, eight bytes each, so a fuzzer
// reaches every bit pattern: NaNs, infinities, negative zero, subnormals.
func floats(b []byte) []float64 {
	out := make([]float64, 0, len(b)/8)
	for len(b) >= 8 {
		out = append(out, math.Float64frombits(binary.LittleEndian.Uint64(b)))
		b = b[8:]
	}
	return out
}

func fuzzBytes(vs ...float64) []byte {
	out := make([]byte, 0, 8*len(vs))
	for _, v := range vs {
		out = binary.LittleEndian.AppendUint64(out, math.Float64bits(v))
	}
	return out
}

// plotRunes is everything a sparkline's chart rows may draw.
const plotRunes = " ▁▂▃▄▅▆▇█"

// FuzzChart feeds arbitrary samples, a max, a title and a size to a
// two-series sparkline, a gauge and a grid of both. Whatever they are,
// nothing panics, and:
//
//   - every view is exactly the height in lines, each exactly the width in
//     cells, in a painted theme and an unpainted one;
//   - a sparkline's chart rows draw only blocks and spaces — no NaN, no
//     stray glyph — and a one-series column is never fuller above than
//     below;
//   - a gauge's ratio is in [0, 1].
func FuzzChart(f *testing.F) {
	for _, s := range []struct {
		title   string
		samples []byte
		top     float64
		w, h    uint8
	}{
		{samples: fuzzBytes(1, 2, 3, 4), top: 0, w: 10, h: 3, title: "CPU"},
		{samples: fuzzBytes(math.NaN(), math.Inf(1), math.Inf(-1), -1, 0), top: 100, w: 4, h: 1, title: ""},
		{samples: fuzzBytes(math.MaxFloat64, -math.MaxFloat64, math.SmallestNonzeroFloat64), top: math.NaN(), w: 2, h: 2},
		{samples: nil, top: math.Inf(1), w: 1, h: 1, title: "\x1b[31mred\x1b["},
		{samples: fuzzBytes(5, 5, 5, 5, 5, 5, 5, 5, 5), top: 5, w: 3, h: 8, title: "日本語\n\t\xff"},
		{samples: fuzzBytes(0.5, math.Copysign(0, -1), 1e-300, 1e300), top: 1e-310, w: 60, h: 5, title: "a long title for a narrow chart"},
	} {
		f.Add(s.samples, s.top, s.title, s.w, s.h)
	}
	f.Fuzz(func(t *testing.T, raw []byte, top float64, title string, w, h uint8) {
		vs := floats(raw)
		width, height := int(w%80)+1, int(h%12)+1
		a, b := NewSeries(len(vs)/2+1), NewSeries(len(vs)+1)
		for i, v := range vs {
			if i%2 == 0 {
				a.Push(v)
			}
			b.Push(v)
		}
		for _, th := range []theme.Theme{theme.Default(), theme.NoPaintBackground()} {
			s := NewSparkline(th, a, b)
			s.SetMax(top)
			s.SetTitle(title)
			s.SetLabel(title)
			s.Resize(width, height)
			lines := checkSize(t, "sparkline", s.View(), width, height)
			checkRunes(t, lines[headerRows(s, height):])

			one := NewSparkline(th, b)
			one.SetMax(top)
			one.Resize(width, height)
			checkColumns(t, checkSize(t, "one-series sparkline", one.View(), width, height))

			g := NewGauge(th)
			g.Set(a.Last(), b.Last())
			g.SetTitle(title)
			g.SetHighThresholds(top, b.Max())
			g.Resize(width, height)
			checkSize(t, "gauge", g.View(), width, height)
			if r := g.Ratio(); !(r >= 0 && r <= 1) {
				t.Fatalf("gauge ratio %v for %v/%v", r, a.Last(), b.Last())
			}

			grid := NewGrid(th, int(w%4)+1, s, g, nil)
			grid.SetGap(int(h % 3))
			grid.Resize(width, height)
			checkSize(t, "grid", grid.View(), width, height)
		}
	})
}

// headerRows is how many lines s spends on its header at height h.
func headerRows(s *Sparkline, h int) int {
	if h > 1 && (s.title != "" || s.label != "") {
		return 1
	}
	return 0
}

// checkSize checks view is exactly w by h and returns its lines, stripped.
func checkSize(t *testing.T, name, view string, w, h int) []string {
	t.Helper()
	lines := strings.Split(view, "\n")
	if len(lines) != h {
		t.Fatalf("%s view is %d lines, want %d: %q", name, len(lines), h, view)
	}
	for i, l := range lines {
		if got := ansi.StringWidth(l); got != w {
			t.Fatalf("%s line %d is %d cells, want %d: %q", name, i, got, w, l)
		}
		lines[i] = ansi.Strip(l)
	}
	return lines
}

// checkRunes checks chart rows hold only blocks and spaces.
func checkRunes(t *testing.T, rows []string) {
	t.Helper()
	for i, r := range rows {
		for _, c := range r {
			if !strings.ContainsRune(plotRunes, c) {
				t.Fatalf("chart row %d draws %q: %q", i, c, r)
			}
		}
	}
}

// checkColumns checks a one-series chart's columns: blocks and spaces
// only, and nothing above a cell short of full.
func checkColumns(t *testing.T, rows []string) {
	t.Helper()
	checkRunes(t, rows)
	grid := make([][]rune, len(rows))
	for i, r := range rows {
		grid[i] = []rune(r)
	}
	for i := 1; i < len(grid); i++ {
		for x, c := range grid[i] {
			if c != fullBlock && grid[i-1][x] != ' ' {
				t.Fatalf("column %d has %q above %q:\n%s", x, grid[i-1][x], c, strings.Join(rows, "\n"))
			}
		}
	}
}
