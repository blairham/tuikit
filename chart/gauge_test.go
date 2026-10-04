// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package chart

import (
	"math"
	"strings"
	"testing"

	"github.com/blairham/tuikit/theme"
)

// gauge returns a gauge at value/total, sized w by h.
func gauge(value, total float64, w, h int) *Gauge {
	g := NewGauge(theme.Default())
	g.Set(value, total)
	g.Resize(w, h)
	return g
}

func TestGaugeBar(t *testing.T) {
	for _, tc := range []struct {
		name         string
		want         string
		value, total float64
		w            int
	}{
		{name: "half", value: 1, total: 2, w: 10, want: "███─── 1/2"},
		{name: "empty", value: 0, total: 4, w: 8, want: "──── 0/4"},
		{name: "full", value: 4, total: 4, w: 8, want: "████ 4/4"},
		{name: "eighths", value: 11, total: 32, w: 10, want: "█▍── 11/32"},
		{name: "a sliver still shows", value: 1, total: 1000, w: 11, want: "▏─── 1/1000"},
		{name: "over the total is full", value: 9, total: 4, w: 8, want: "████ 9/4"},
		{name: "negative is empty", value: -3, total: 4, w: 9, want: "──── -3/4"},
		{name: "NaN is empty", value: math.NaN(), total: 4, w: 10, want: "──── NaN/4"},
		{name: "+Inf is full", value: math.Inf(1), total: 4, w: 10, want: "███ +Inf/4"},
		{name: "no total is empty", value: 3, total: 0, w: 8, want: "──── 3/0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := gauge(tc.value, tc.total, tc.w, 1)
			equalLines(t, stripped(t, g.View(), tc.w, 1), tc.want)
			if r := g.Ratio(); !(r >= 0 && r <= 1) {
				t.Errorf("Ratio() = %v, want it in [0, 1]", r)
			}
		})
	}
}

func TestGaugeHeader(t *testing.T) {
	g := gauge(3, 7, 20, 3)
	g.SetTitle("Pods")
	g.SetLabel("3/7 running")
	equalLines(t, stripped(t, g.View(), 20, 3),
		"Pods     3/7 running",
		"████████▋───────────", // 3/7 of 20 cells is 8 and 5/8
		"████████▋───────────", // 3/7 of 20 cells is 8 and 5/8
	)
}

func TestGaugeInlineDropsTitleThenLabel(t *testing.T) {
	g := gauge(1, 2, 12, 1)
	g.SetTitle("Pods")
	equalLines(t, stripped(t, g.View(), 12, 1), "Pods █▌─ 1/2")
	g.Resize(9, 1)
	equalLines(t, stripped(t, g.View(), 9, 1), "██▌── 1/2")
	g.Resize(3, 1)
	equalLines(t, stripped(t, g.View(), 3, 1), "█▌─")
}

func TestGaugeThresholds(t *testing.T) {
	for _, tc := range []struct {
		name  string
		fg    string
		value float64
		want  Level
		low   bool
	}{
		{name: "high: under warn", value: 5, want: LevelOK, fg: paleGreenFg},
		{name: "high: at warn", value: 8, want: LevelWarn, fg: yellowFg},
		{name: "high: at critical", value: 9, want: LevelCritical, fg: redFg},
		{name: "high: full", value: 10, want: LevelCritical, fg: redFg},
		{name: "low: all there", low: true, value: 10, want: LevelOK, fg: paleGreenFg},
		{name: "low: one missing", low: true, value: 9, want: LevelWarn, fg: yellowFg},
		{name: "low: at half", low: true, value: 5, want: LevelWarn, fg: yellowFg},
		{name: "low: under half", low: true, value: 4, want: LevelCritical, fg: redFg},
		{name: "low: none", low: true, value: 0, want: LevelCritical, fg: redFg},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := gauge(tc.value, 10, 20, 2)
			if tc.low {
				g.SetLowThresholds(1, 0.5)
			} else {
				g.SetHighThresholds(0.8, 0.9)
			}
			if got := g.Level(); got != tc.want {
				t.Fatalf("Level() = %v, want %v", got, tc.want)
			}
			lines := strings.Split(g.View(), "\n")
			for _, fg := range []string{paleGreenFg, yellowFg, redFg} {
				if got := strings.Contains(lines[0], fg); got != (fg == tc.fg) {
					t.Errorf("label %q: has %s = %v, want only %s", lines[0], fg, got, tc.fg)
				}
				if tc.value > 0 {
					if got := strings.Contains(lines[1], fg); got != (fg == tc.fg) {
						t.Errorf("bar %q: has %s = %v, want only %s", lines[1], fg, got, tc.fg)
					}
				}
			}
		})
	}
}

func TestGaugeThresholdsOff(t *testing.T) {
	g := gauge(10, 10, 10, 1)
	if g.Level() != LevelOK {
		t.Errorf("no thresholds: Level() = %v, want OK", g.Level())
	}
	g.SetHighThresholds(math.NaN(), math.NaN())
	if g.Level() != LevelOK {
		t.Errorf("NaN thresholds: Level() = %v, want OK", g.Level())
	}
	g.SetHighThresholds(0.5, 0.9)
	g.ClearThresholds()
	if g.Level() != LevelOK {
		t.Errorf("after ClearThresholds: Level() = %v, want OK", g.Level())
	}
	g.SetLowThresholds(1, 0.5)
	for _, total := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		g.Set(0, total)
		if g.Level() != LevelOK || g.Ratio() != 0 {
			t.Errorf("total %v: Level() = %v, Ratio() = %v; want OK, 0", total, g.Level(), g.Ratio())
		}
	}
}

func TestGaugeUnsized(t *testing.T) {
	if got := NewGauge(theme.Default()).View(); got != "" {
		t.Errorf("View before Resize = %q, want empty", got)
	}
}
