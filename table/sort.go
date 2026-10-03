// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package table

import (
	"cmp"
	"slices"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/table"
	"github.com/charmbracelet/x/ansi"
)

// Conventional sort-column bindings (k9s uses shift+← / shift+→). Like the
// chrome's key names, these are names, not behavior: apps match them in
// their own Update and pass the key to [Sorter.HandleKey].
const (
	KeySortPrev = "shift+left"
	KeySortNext = "shift+right"
)

// Sort direction indicators [Sorter.Columns] appends to the active column's
// title, as k9s draws them.
const (
	SortAscIndicator  = "↑"
	SortDescIndicator = "↓"
)

// Sorter owns a table's sort column and direction and sorts rows by them.
// Apps keep one per table, route shift+←/→ through [Sorter.HandleKey], and
// on every refresh call [Sorter.Sort] on the new rows and [Sorter.Columns]
// on the column set before handing both to the bubbles table, so the order
// and the header indicator survive the refresh.
//
// The zero value sorts ascending on column 0 with [NaturalCompare].
type Sorter struct {
	compare func(col int, a, b string) int
	col     int
	desc    bool
}

// NewSorter returns a Sorter on column col, descending when desc is true.
func NewSorter(col int, desc bool) *Sorter {
	return &Sorter{col: max(col, 0), desc: desc}
}

// Column returns the index of the active sort column.
func (s *Sorter) Column() int { return s.col }

// Descending reports whether the sort runs descending.
func (s *Sorter) Descending() bool { return s.desc }

// SortBy makes col the sort column, ascending. Selecting the column that is
// already active reverses the direction instead, as k9s's shift+<letter>
// column shortcuts do — so an app binds one key per column to SortBy.
func (s *Sorter) SortBy(col int) {
	if col < 0 {
		return
	}
	if col == s.col {
		s.desc = !s.desc
		return
	}
	s.col, s.desc = col, false
}

// Reverse flips the sort direction without changing the column.
func (s *Sorter) Reverse() { s.desc = !s.desc }

// SetCompare replaces the cell comparison. fn returns a negative number
// when a sorts before b in ascending order, zero when they tie, and a
// positive number otherwise; col is the column being compared, so one
// function can treat an age column differently from a name column. Cells
// arrive with ANSI escapes already stripped. nil restores
// [NaturalCompare].
func (s *Sorter) SetCompare(fn func(col int, a, b string) int) { s.compare = fn }

// HandleKey moves the sort column on [KeySortPrev] / [KeySortNext] across a
// table of ncols columns, stopping at the first and last column. A move
// resets the direction to ascending. It reports whether the key was a sort
// key, so the app can stop routing it; a sort key is reported handled even
// at the edge, where it does nothing.
func (s *Sorter) HandleKey(key string, ncols int) bool {
	var step int
	switch key {
	case KeySortPrev:
		step = -1
	case KeySortNext:
		step = 1
	default:
		return false
	}
	if ncols <= 0 {
		return true
	}
	next := min(max(s.col+step, 0), ncols-1)
	if next != s.col {
		s.col, s.desc = next, false
	}
	return true
}

// Sort orders rows in place by the active column and direction. The sort is
// stable, so rows that tie keep their incoming order and a refresh with
// unchanged data does not shuffle them. A row too short to have the column
// sorts as an empty cell.
func (s *Sorter) Sort(rows []table.Row) {
	compare := s.compare
	if compare == nil {
		compare = func(_ int, a, b string) int { return NaturalCompare(a, b) }
	}
	slices.SortStableFunc(rows, func(a, b table.Row) int {
		c := compare(s.col, cell(a, s.col), cell(b, s.col))
		if s.desc {
			return -c
		}
		return c
	})
}

// Columns returns a copy of cols with the direction indicator appended to
// the active column's title. Widths are left as they are, so a column sized
// exactly to its title should leave a cell for the indicator.
func (s *Sorter) Columns(cols []table.Column) []table.Column {
	out := slices.Clone(cols)
	if s.col < len(out) {
		ind := SortAscIndicator
		if s.desc {
			ind = SortDescIndicator
		}
		out[s.col].Title += ind
	}
	return out
}

// NaturalCompare is the default cell comparison: two cells that both parse
// as numbers compare numerically (so "10" sorts after "9"), and anything
// else compares case-insensitively, falling back to a byte comparison so
// the order is total. Surrounding spaces are ignored.
func NaturalCompare(a, b string) int {
	a, b = strings.TrimSpace(a), strings.TrimSpace(b)
	if fa, errA := strconv.ParseFloat(a, 64); errA == nil {
		if fb, errB := strconv.ParseFloat(b, 64); errB == nil {
			if c := cmp.Compare(fa, fb); c != 0 {
				return c
			}
		}
	}
	if c := strings.Compare(strings.ToLower(a), strings.ToLower(b)); c != 0 {
		return c
	}
	return strings.Compare(a, b)
}

func cell(r table.Row, col int) string {
	if col < len(r) {
		return ansi.Strip(r[col])
	}
	return ""
}
