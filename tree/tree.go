// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package tree

import (
	"github.com/charmbracelet/x/ansi"

	"github.com/blairham/tuikit/table"
	"github.com/blairham/tuikit/theme"
)

// KeyToggle is the binding that expands or collapses the node under the
// cursor: space, as in k9s's xray. Enter is deliberately not bound, so an
// app can give it its own action (drill into the selected resource).
const KeyToggle = "space"

// Node is one entry in the tree. ID is the node's identity across
// [Model.SetRoots] calls — expansion state and the cursor follow it — so it
// must be unique within the tree and stable from one refresh to the next
// (for a resource, its kind, namespace and name together). Label is what
// the row shows; it may carry SGR color, and it is what [Model.SetFilter]
// matches, with ANSI stripped. A nil child is skipped.
type Node struct {
	ID       string
	Label    string
	Children []*Node
}

// row is one visible line of the tree.
type row struct {
	node   *Node
	guide  string // the box-drawing prefix, e.g. "│  ├─ "
	parent int    // row index of the parent; -1 for a root
	kids   bool   // has children shown under the current filter
	open   bool
}

// Model is a tree view: the content, which nodes are open, the cursor and
// the scroll offset. Build one with [New], feed it with [Model.SetRoots],
// route keys through [Model.HandleKey] and render [Model.View].
type Model struct {
	filtered map[string]bool // expansion state while a filter is set; SetFilter resets it
	passes   map[*Node]bool  // what the filter keeps; nil when there is no filter
	byID     map[string]int  // row index of each visible ID
	expanded map[string]bool // the user's expansion state, by ID
	styles   styles
	roots    []*Node
	rows     []row
	selPath  []string // IDs from the root down to the selected node
	filter   table.RowFilter
	depth    int // nodes shallower than this start open; negative opens all
	cursor   int
	offset   int
	width    int
	height   int
	count    int
}

// New returns an empty tree drawn in t's colors. New nodes start with only
// the roots open; see [Model.SetDefaultDepth].
func New(t theme.Theme) *Model {
	return &Model{
		expanded: map[string]bool{},
		filtered: map[string]bool{},
		byID:     map[string]int{},
		styles:   newStyles(t),
		depth:    1,
	}
}

// SetDefaultDepth sets how deep a node seen for the first time starts
// open: a node at depth d (a root is 0) starts expanded when d < n. 1, the
// default, opens the roots only; 0 starts everything collapsed; a negative
// n opens every level. Nodes already seen keep their state, so set this
// before the first [Model.SetRoots].
func (m *Model) SetDefaultDepth(n int) { m.depth = n }

// SetRoots replaces the content. A node keeps its expansion state from the
// last call when its ID is the same, and the cursor stays on the selected
// node's ID wherever that node now sits. When the selected node is under a
// collapsed node, the cursor moves to its nearest ancestor on screen; when
// it has left the tree, to the nearest of the ancestors it had that is
// still on screen; failing both, it stays on the same row, clamped to the
// rows there are. Expansion state for IDs no longer in the tree is dropped.
//
// The tree must be a tree: a node reachable twice, or a cycle, is not
// supported.
func (m *Model) SetRoots(roots []*Node) {
	m.roots = roots
	seen := make(map[string]bool, len(m.expanded))
	m.count = 0
	m.record(roots, 0, seen)
	for id := range m.expanded {
		if !seen[id] {
			delete(m.expanded, id)
		}
	}
	m.refilter()
	m.rebuild()
}

// record counts nodes and gives each newly seen ID its default state.
func (m *Model) record(nodes []*Node, depth int, seen map[string]bool) {
	for _, n := range nodes {
		if n == nil {
			continue
		}
		m.count++
		seen[n.ID] = true
		if _, ok := m.expanded[n.ID]; !ok {
			m.expanded[n.ID] = m.depth < 0 || depth < m.depth
		}
		m.record(n.Children, depth+1, seen)
	}
}

// Count returns the number of nodes in the tree, shown or not.
func (m *Model) Count() int { return m.count }

// VisibleCount returns the number of rows the tree has now: the nodes the
// filter keeps whose ancestors are all open.
func (m *Model) VisibleCount() int { return len(m.rows) }

// SetFilter shows only what matches expr, a [table.ParseFilter]
// expression: a regex (case-insensitive, "!" to negate) or "-f term" for a
// fuzzy match. A node is shown when its label, ANSI stripped, matches or
// when any of its descendants' does, so every match is reachable; the
// ancestors of a match start open. Nodes carry no labels in the "-l"
// sense, so a label selector shows nothing.
//
// Opening and closing nodes while a filter is set changes the filtered
// view only: the next SetFilter starts it afresh, and an empty filter
// shows the whole tree in the expansion state the user left it in.
func (m *Model) SetFilter(expr string) {
	m.filter = table.ParseFilter(expr)
	clear(m.filtered)
	m.refilter()
	m.rebuild()
}

// refilter recomputes which nodes the filter keeps.
func (m *Model) refilter() {
	if m.filter.Empty() {
		m.passes = nil
		return
	}
	m.passes = map[*Node]bool{}
	m.markPasses(m.roots)
}

// markPasses records each node that matches or has a matching descendant,
// and reports whether any of nodes did.
func (m *Model) markPasses(nodes []*Node) bool {
	found := false
	for _, n := range nodes {
		if n == nil {
			continue
		}
		pass := m.markPasses(n.Children)
		if !pass {
			pass = m.filter.MatchesAny(ansi.Strip(n.Label))
		}
		if pass {
			m.passes[n] = true
			found = true
		}
	}
	return found
}

// shown is the children of a node (or the roots) that the filter keeps.
func (m *Model) shown(nodes []*Node) []*Node {
	out := make([]*Node, 0, len(nodes))
	for _, n := range nodes {
		if n != nil && (m.passes == nil || m.passes[n]) {
			out = append(out, n)
		}
	}
	return out
}

// isOpen reports whether n's children are drawn.
func (m *Model) isOpen(n *Node) bool {
	if m.passes != nil {
		return !m.filtered[n.ID]
	}
	return m.expanded[n.ID]
}

// setOpen opens or closes n in the state that is in force.
func (m *Model) setOpen(n *Node, open bool) {
	if m.passes != nil {
		m.filtered[n.ID] = !open
		return
	}
	m.expanded[n.ID] = open
}

// rebuild recomputes the rows and puts the cursor back on the selected
// node, or the nearest ancestor of it that is still shown.
func (m *Model) rebuild() {
	m.rows = m.rows[:0]
	clear(m.byID)
	m.walk(m.roots, -1, "", true)
	m.relocate()
	m.follow()
}

// walk appends the rows for nodes, children of the row at parent (-1 for
// the roots), with prefix as the guides drawn for their ancestors.
func (m *Model) walk(nodes []*Node, parent int, prefix string, roots bool) {
	kids := m.shown(nodes)
	for i, n := range kids {
		last := i == len(kids)-1
		guide, next := "", ""
		if !roots {
			guide, next = prefix+"├─ ", prefix+"│  "
			if last {
				guide, next = prefix+"└─ ", prefix+"   "
			}
		}
		r := row{node: n, guide: guide, parent: parent, open: m.isOpen(n)}
		r.kids = len(m.shown(n.Children)) > 0
		idx := len(m.rows)
		m.rows = append(m.rows, r)
		if _, dup := m.byID[n.ID]; !dup {
			m.byID[n.ID] = idx
		}
		if r.kids && r.open {
			m.walk(n.Children, idx, next, false)
		}
	}
}

// relocate puts the cursor back on the selected node. When that node is
// not on screen it goes to the nearest ancestor that is: the node's
// ancestors where it now sits, or, when it has left the tree, the ones it
// had. Failing both, the cursor keeps its row, clamped to the rows.
func (m *Model) relocate() {
	if len(m.selPath) > 0 {
		for _, path := range [][]string{m.pathTo(m.selPath[len(m.selPath)-1]), m.selPath} {
			for i := len(path) - 1; i >= 0; i-- {
				if idx, ok := m.byID[path[i]]; ok {
					m.cursor = idx
					m.remember()
					return
				}
			}
		}
	}
	m.cursor = min(max(m.cursor, 0), max(len(m.rows)-1, 0))
	m.remember()
}

// pathTo returns the IDs from a root down to the first node with ID id,
// shown or not; nil when there is none.
func (m *Model) pathTo(id string) []string {
	var path []string
	var find func([]*Node) bool
	find = func(nodes []*Node) bool {
		for _, n := range nodes {
			if n == nil {
				continue
			}
			path = append(path, n.ID)
			if n.ID == id || find(n.Children) {
				return true
			}
			path = path[:len(path)-1]
		}
		return false
	}
	if !find(m.roots) {
		return nil
	}
	return path
}

// remember records the path to the row under the cursor.
func (m *Model) remember() {
	m.selPath = m.selPath[:0]
	if m.cursor >= len(m.rows) {
		return
	}
	for i := m.cursor; i >= 0; i = m.rows[i].parent {
		m.selPath = append(m.selPath, m.rows[i].node.ID)
	}
	for i, j := 0, len(m.selPath)-1; i < j; i, j = i+1, j-1 {
		m.selPath[i], m.selPath[j] = m.selPath[j], m.selPath[i]
	}
}

// Selected returns the node under the cursor; false when the tree shows no
// rows.
func (m *Model) Selected() (*Node, bool) {
	if m.cursor >= len(m.rows) {
		return nil, false
	}
	return m.rows[m.cursor].node, true
}

// SelectedPath returns the nodes from the selected node's root down to the
// selected node, both included; nil when the tree shows no rows.
func (m *Model) SelectedPath() []*Node {
	if m.cursor >= len(m.rows) {
		return nil
	}
	var path []*Node
	for i := m.cursor; i >= 0; i = m.rows[i].parent {
		path = append(path, m.rows[i].node)
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}

// ExpandAll opens every node. While a filter is set it opens every node of
// the filtered view.
func (m *Model) ExpandAll() { m.setAll(m.roots, true) }

// CollapseAll closes every node, leaving the roots. The cursor moves up to
// the root it was under. While a filter is set it closes the filtered view.
func (m *Model) CollapseAll() { m.setAll(m.roots, false) }

func (m *Model) setAll(roots []*Node, open bool) {
	var each func([]*Node)
	each = func(nodes []*Node) {
		for _, n := range nodes {
			if n != nil {
				m.setOpen(n, open)
				each(n.Children)
			}
		}
	}
	each(roots)
	m.rebuild()
}

// Resize sets the size [Model.View] draws at and keeps the cursor on
// screen.
func (m *Model) Resize(width, height int) {
	m.width, m.height = width, height
	m.follow()
}

// follow scrolls so the cursor is on screen, and no further down than the
// rows need.
func (m *Model) follow() {
	h := max(m.height, 1)
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+h {
		m.offset = m.cursor - h + 1
	}
	m.offset = max(min(m.offset, len(m.rows)-h), 0)
}
