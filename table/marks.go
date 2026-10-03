// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package table

import (
	"charm.land/bubbles/v2/table"
	"github.com/charmbracelet/x/ansi"

	"github.com/blairham/tuikit/theme"
)

// Conventional mark bindings (k9s's). Like the sort keys, these are names:
// apps match them in their own Update, or pass every key to
// [Marks.HandleKey], which knows them.
const (
	// KeyMarkToggle marks or unmarks the row under the cursor.
	KeyMarkToggle = "space"
	// KeyMarkRange marks every row from the last one marked to the cursor.
	KeyMarkRange = "ctrl+space"
	// KeyMarkClear clears every mark.
	KeyMarkClear = "ctrl+\\"
)

// keyMarkRangeNUL is how a terminal without enhanced keyboard reporting
// can deliver ctrl+space: as the NUL byte, which reads as ctrl+@.
const keyMarkRangeNUL = "ctrl+@"

// Marks tracks rows marked for a bulk action by a stable key rather than by
// position, so marks follow their rows through a re-sort, a filter and a
// refresh that reorders the data. Apps keep one per table, route keys
// through [Marks.HandleKey], act on [Marks.Selected], and render the rows
// through [Marks.Style] so marked rows stand out.
type Marks struct {
	key  func(table.Row) string
	set  map[string]struct{}
	last string
}

// NewMarks returns an empty mark set that identifies a row by key. A nil
// key uses the row's first cell with ANSI stripped, which fits a table
// whose first column is unique (a name); a namespaced table should key on
// namespace and name together.
func NewMarks(key func(table.Row) string) *Marks {
	if key == nil {
		key = func(r table.Row) string { return cell(r, 0) }
	}
	return &Marks{key: key, set: map[string]struct{}{}}
}

// HandleKey applies a mark binding to rows, the rows as currently displayed,
// with the cursor at index cursor (bubbles table's Cursor). It reports
// whether key was a mark binding. A mark binding with the cursor off the
// rows is still reported handled and does nothing.
func (m *Marks) HandleKey(key string, rows []table.Row, cursor int) bool {
	switch key {
	case KeyMarkToggle:
		if cursor >= 0 && cursor < len(rows) {
			m.Toggle(rows[cursor])
		}
	case KeyMarkRange, keyMarkRangeNUL:
		m.MarkRange(rows, cursor)
	case KeyMarkClear:
		m.Clear()
	default:
		return false
	}
	return true
}

// Toggle marks row, or unmarks it if it is already marked. A row that gets
// marked becomes the anchor for the next [Marks.MarkRange].
func (m *Marks) Toggle(row table.Row) {
	k := m.key(row)
	if _, ok := m.set[k]; ok {
		delete(m.set, k)
		if m.last == k {
			m.last = ""
		}
		return
	}
	m.set[k] = struct{}{}
	m.last = k
}

// MarkRange marks every row between the last row marked and the cursor,
// both included, in display order. With no anchor on screen it marks the
// cursor row alone. The cursor row becomes the new anchor.
func (m *Marks) MarkRange(rows []table.Row, cursor int) {
	if cursor < 0 || cursor >= len(rows) {
		return
	}
	from := cursor
	if m.last != "" {
		for i, r := range rows {
			if m.key(r) == m.last {
				from = i
				break
			}
		}
	}
	lo, hi := min(from, cursor), max(from, cursor)
	for _, r := range rows[lo : hi+1] {
		m.set[m.key(r)] = struct{}{}
	}
	m.last = m.key(rows[cursor])
}

// Clear removes every mark.
func (m *Marks) Clear() {
	clear(m.set)
	m.last = ""
}

// IsMarked reports whether row is marked.
func (m *Marks) IsMarked(row table.Row) bool {
	_, ok := m.set[m.key(row)]
	return ok
}

// Len returns the number of marks.
func (m *Marks) Len() int { return len(m.set) }

// Prune drops the marks whose rows are not in rows, for a refresh in which
// resources went away. Marks survive a refresh without it; call it with the
// full data set, not a filtered view, or a filter will clear marks it hides.
func (m *Marks) Prune(rows []table.Row) {
	keep := make(map[string]struct{}, len(rows))
	for _, r := range rows {
		keep[m.key(r)] = struct{}{}
	}
	for k := range m.set {
		if _, ok := keep[k]; !ok {
			delete(m.set, k)
		}
	}
	if _, ok := m.set[m.last]; !ok {
		m.last = ""
	}
}

// Selected returns the rows a bulk action applies to: the marked rows, in
// display order, or the row under the cursor when nothing on screen is
// marked — so one action handler serves both a single row and a marked
// set, as in k9s. It returns nil when there is neither.
func (m *Marks) Selected(rows []table.Row, cursor int) []table.Row {
	var out []table.Row
	for _, r := range rows {
		if m.IsMarked(r) {
			out = append(out, r)
		}
	}
	if len(out) == 0 && cursor >= 0 && cursor < len(rows) {
		out = []table.Row{rows[cursor]}
	}
	return out
}

// Style returns a copy of rows with every marked row's cells drawn in
// [theme.Theme.MarkStyle], replacing their own styling so the whole row
// reads as marked. Unmarked rows are shared with the input, not copied.
// Pass the result to the bubbles table and keep the unstyled rows for
// [Marks.HandleKey] and [Marks.Selected], so a custom key never sees the
// mark styling.
func (m *Marks) Style(rows []table.Row, t theme.Theme) []table.Row {
	if len(m.set) == 0 {
		return rows
	}
	out := make([]table.Row, len(rows))
	for i, r := range rows {
		if !m.IsMarked(r) {
			out[i] = r
			continue
		}
		styled := make(table.Row, len(r))
		for j, c := range r {
			styled[j] = t.MarkStyle.Render(ansi.Strip(c))
		}
		out[i] = styled
	}
	return out
}
