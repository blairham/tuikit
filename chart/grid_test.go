// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package chart

import (
	"fmt"
	"strings"
	"testing"

	"github.com/blairham/tuikit/theme"
)

// fill is a widget that draws its letter across whatever size it is given,
// or, with a fixed size set, that size regardless.
type fill struct {
	letter     string
	fixedW     int
	fixedH     int
	gotW, gotH int
}

func (f *fill) Resize(w, h int) { f.gotW, f.gotH = w, h }

func (f *fill) View() string {
	w, h := f.gotW, f.gotH
	if f.fixedW > 0 {
		w, h = f.fixedW, f.fixedH
	}
	lines := make([]string, h)
	for i := range lines {
		lines[i] = strings.Repeat(f.letter, w)
	}
	return strings.Join(lines, "\n")
}

func TestGridLayout(t *testing.T) {
	a, b, c := &fill{letter: "a"}, &fill{letter: "b"}, &fill{letter: "c"}
	g := NewGrid(theme.Default(), 2, a, b, c)
	g.Resize(8, 5)
	// 8 cells less a gap of 1 is 4 + 3; 5 lines over two rows is 3 + 2.
	for _, tc := range []struct {
		f    *fill
		w, h int
	}{{f: a, w: 4, h: 3}, {f: b, w: 3, h: 3}, {f: c, w: 4, h: 2}} {
		if tc.f.gotW != tc.w || tc.f.gotH != tc.h {
			t.Errorf("%s resized to %dx%d, want %dx%d", tc.f.letter, tc.f.gotW, tc.f.gotH, tc.w, tc.h)
		}
	}
	equalLines(t, stripped(t, g.View(), 8, 5),
		"aaaa bbb",
		"aaaa bbb",
		"aaaa bbb",
		"cccc    ",
		"cccc    ",
	)
}

func TestGridGap(t *testing.T) {
	a, b, c := &fill{letter: "a"}, &fill{letter: "b"}, &fill{letter: "c"}
	g := NewGrid(theme.Default(), 3, a, b, c)
	g.SetGap(2)
	g.Resize(10, 1)
	equalLines(t, stripped(t, g.View(), 10, 1), "aa  bb  cc")
	// Too narrow for the gap and a cell per column: the gap shrinks.
	g.Resize(5, 1)
	equalLines(t, stripped(t, g.View(), 5, 1), "a b c")
	g.Resize(4, 1)
	equalLines(t, stripped(t, g.View(), 4, 1), "aabc")
	// Narrower than the columns: some get nothing.
	g.Resize(2, 1)
	equalLines(t, stripped(t, g.View(), 2, 1), "ab")
	g.SetGap(-3)
	g.Resize(6, 1)
	equalLines(t, stripped(t, g.View(), 6, 1), "aabbcc")
}

func TestGridFitsMisbehavingWidgets(t *testing.T) {
	big := &fill{letter: "x", fixedW: 9, fixedH: 9}
	small := &fill{letter: "y", fixedW: 1, fixedH: 1}
	g := NewGrid(theme.Default(), 2, big, small)
	g.Resize(7, 2)
	equalLines(t, stripped(t, g.View(), 7, 2), "xxx y  ", "xxx    ")
}

func TestGridNests(t *testing.T) {
	inner := NewGrid(theme.Default(), 1, &fill{letter: "a"}, &fill{letter: "b"})
	g := NewGrid(theme.Default(), 2, inner, &fill{letter: "c"})
	g.SetGap(0)
	g.Resize(4, 2)
	equalLines(t, stripped(t, g.View(), 4, 2), "aacc", "bbcc")
}

func TestGridSizes(t *testing.T) {
	for _, cols := range []int{0, 1, 2, 3, 5} {
		for _, n := range []int{0, 1, 4, 7} {
			ws := make([]Widget, n)
			for i := range ws {
				ws[i] = &fill{letter: fmt.Sprint(i % 10)}
			}
			g := NewGrid(theme.NoPaintBackground(), cols, ws...)
			for _, w := range []int{1, 2, 5, 13} {
				for _, h := range []int{1, 2, 7} {
					g.Resize(w, h)
					stripped(t, g.View(), w, h)
				}
			}
		}
	}
	if got := NewGrid(theme.Default(), 2).View(); got != "" {
		t.Errorf("View before Resize = %q, want empty", got)
	}
}
