// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package chart

import "math"

// Series is a bounded ring buffer of samples: it holds the most recent
// Cap() values pushed, dropping the oldest once full. The zero value is
// unusable; build one with [NewSeries]. A Series is not safe for
// concurrent use — push from the same goroutine that renders, as with any
// Bubble Tea model.
type Series struct {
	buf   []float64
	start int // index of the oldest sample
	n     int
}

// NewSeries returns an empty series holding at most capacity samples. A
// capacity below 1 is taken as 1.
func NewSeries(capacity int) *Series {
	return &Series{buf: make([]float64, max(capacity, 1))}
}

// Push appends v as the newest sample, dropping the oldest when the series
// is full. Any float64 is kept as given, NaN and the infinities included;
// see the package doc for how they are drawn.
func (s *Series) Push(v float64) {
	if s.n < len(s.buf) {
		s.buf[(s.start+s.n)%len(s.buf)] = v
		s.n++
		return
	}
	s.buf[s.start] = v
	s.start = (s.start + 1) % len(s.buf)
}

// Values returns a copy of the samples, oldest first.
func (s *Series) Values() []float64 {
	out := make([]float64, s.n)
	for i := range out {
		out[i] = s.buf[(s.start+i)%len(s.buf)]
	}
	return out
}

// Last returns the newest sample, or 0 when the series is empty.
func (s *Series) Last() float64 {
	if s.n == 0 {
		return 0
	}
	return s.buf[(s.start+s.n-1)%len(s.buf)]
}

// Max returns the largest sample, ignoring NaN, or 0 when the series is
// empty or holds only NaN.
func (s *Series) Max() float64 {
	out, found := 0.0, false
	for i := range s.n {
		v := s.buf[(s.start+i)%len(s.buf)]
		if math.IsNaN(v) {
			continue
		}
		if !found || v > out {
			out, found = v, true
		}
	}
	return out
}

// Len returns the number of samples held.
func (s *Series) Len() int { return s.n }

// Cap returns the most samples the series holds.
func (s *Series) Cap() int { return len(s.buf) }

// Reset drops every sample, keeping the capacity.
func (s *Series) Reset() { s.start, s.n = 0, 0 }
