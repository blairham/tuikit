// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package chart

import (
	"math"
	"slices"
	"testing"
)

func TestSeriesWraps(t *testing.T) {
	s := NewSeries(3)
	for i, want := range [][]float64{{1}, {1, 2}, {1, 2, 3}, {2, 3, 4}, {3, 4, 5}, {4, 5, 6}, {5, 6, 7}} {
		s.Push(float64(i + 1))
		if got := s.Values(); !slices.Equal(got, want) {
			t.Fatalf("after pushing %d: Values() = %v, want %v", i+1, got, want)
		}
		if got := s.Last(); got != float64(i+1) {
			t.Fatalf("after pushing %d: Last() = %v", i+1, got)
		}
		if got := s.Len(); got != len(want) {
			t.Fatalf("after pushing %d: Len() = %d, want %d", i+1, got, len(want))
		}
	}
	if s.Cap() != 3 {
		t.Errorf("Cap() = %d, want 3", s.Cap())
	}
	if got := s.Max(); got != 7 {
		t.Errorf("Max() = %v, want 7", got)
	}
	// Values is a copy: changing it leaves the series alone.
	s.Values()[0] = 99
	if got := s.Values(); got[0] != 5 {
		t.Errorf("Values() shares its backing array: %v", got)
	}
}

func TestSeriesMaxDropsOlderSamples(t *testing.T) {
	s := NewSeries(2)
	for _, v := range []float64{9, 1, 2} {
		s.Push(v)
	}
	if got := s.Max(); got != 2 {
		t.Errorf("Max() = %v, want 2: the 9 was pushed out", got)
	}
}

func TestSeriesEmpty(t *testing.T) {
	s := NewSeries(4)
	if got := s.Values(); len(got) != 0 {
		t.Errorf("Values() = %v, want empty", got)
	}
	if s.Last() != 0 || s.Max() != 0 || s.Len() != 0 {
		t.Errorf("empty series: Last %v, Max %v, Len %d; want 0, 0, 0", s.Last(), s.Max(), s.Len())
	}
}

func TestSeriesCapacityAtLeastOne(t *testing.T) {
	for _, c := range []int{0, -5} {
		s := NewSeries(c)
		s.Push(1)
		s.Push(2)
		if got := s.Values(); s.Cap() != 1 || !slices.Equal(got, []float64{2}) {
			t.Errorf("NewSeries(%d): Cap %d, Values %v; want 1, [2]", c, s.Cap(), got)
		}
	}
}

func TestSeriesMax(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   []float64
		want float64
	}{
		{name: "negatives", in: []float64{-3, -1, -2}, want: -1},
		{name: "NaN ignored", in: []float64{math.NaN(), 2, math.NaN()}, want: 2},
		{name: "only NaN", in: []float64{math.NaN()}, want: 0},
		{name: "+Inf", in: []float64{1, math.Inf(1)}, want: math.Inf(1)},
	} {
		s := NewSeries(len(tc.in))
		for _, v := range tc.in {
			s.Push(v)
		}
		if got := s.Max(); got != tc.want {
			t.Errorf("%s: Max() = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestSeriesReset(t *testing.T) {
	s := NewSeries(2)
	s.Push(1)
	s.Push(2)
	s.Push(3)
	s.Reset()
	if s.Len() != 0 || s.Cap() != 2 {
		t.Fatalf("after Reset: Len %d, Cap %d; want 0, 2", s.Len(), s.Cap())
	}
	s.Push(4)
	if got := s.Values(); !slices.Equal(got, []float64{4}) {
		t.Errorf("after Reset and Push(4): Values() = %v", got)
	}
}
