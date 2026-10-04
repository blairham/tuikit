// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package tree

// The named keys HandleKey reads, as tea.KeyPressMsg.String spells them.
const (
	keyUp     = "up"
	keyDown   = "down"
	keyPgUp   = "pgup"
	keyPgDown = "pgdown"
	keyHome   = "home"
	keyEnd    = "end"
	keyRight  = "right"
	keyLeft   = "left"
)

// HandleKey applies a navigation key and reports whether the tree uses it.
// Keys it does not use — enter among them — return false so the app can
// fall through to its own bindings. A bound key that has nothing to do
// (up on the first row, right on a leaf) is still reported handled.
//
// Bindings:
//
//	k / up             previous row
//	j / down           next row
//	ctrl+b / pgup      a page up
//	ctrl+f / pgdown    a page down
//	g / home           first row
//	G / end            last row
//	l / right          expand; on an open node, go to its first child
//	h / left           collapse; on a closed node or a leaf, go to its parent
//	space              expand or collapse ([KeyToggle])
//
// The vim letters are the ones [viewfsm.TranslateNavKey] rewrites, so an
// app that translates them first gets the same moves.
func (m *Model) HandleKey(key string) bool {
	switch key {
	case "k", keyUp:
		m.moveTo(m.cursor - 1)
	case "j", keyDown:
		m.moveTo(m.cursor + 1)
	case "ctrl+b", keyPgUp:
		m.moveTo(m.cursor - m.page())
	case "ctrl+f", keyPgDown:
		m.moveTo(m.cursor + m.page())
	case "g", keyHome:
		m.moveTo(0)
	case "G", keyEnd:
		m.moveTo(len(m.rows) - 1)
	case "l", keyRight:
		m.right()
	case "h", keyLeft:
		m.left()
	case KeyToggle:
		m.toggle()
	default:
		return false
	}
	return true
}

// page is how far a page key moves: the height, at least one row.
func (m *Model) page() int { return max(m.height, 1) }

// moveTo puts the cursor on row i, clamped to the rows there are.
func (m *Model) moveTo(i int) {
	if len(m.rows) == 0 {
		return
	}
	m.cursor = min(max(i, 0), len(m.rows)-1)
	m.remember()
	m.follow()
}

func (m *Model) right() {
	if m.cursor >= len(m.rows) {
		return
	}
	r := m.rows[m.cursor]
	switch {
	case !r.kids:
	case !r.open:
		m.setOpen(r.node, true)
		m.rebuild()
	default:
		m.moveTo(m.cursor + 1) // an open node's first child is the next row
	}
}

func (m *Model) left() {
	if m.cursor >= len(m.rows) {
		return
	}
	r := m.rows[m.cursor]
	switch {
	case r.kids && r.open:
		m.setOpen(r.node, false)
		m.rebuild()
	case r.parent >= 0:
		m.moveTo(r.parent)
	}
}

func (m *Model) toggle() {
	if m.cursor >= len(m.rows) {
		return
	}
	if r := m.rows[m.cursor]; r.kids {
		m.setOpen(r.node, !r.open)
		m.rebuild()
	}
}
