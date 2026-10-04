// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package tree

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/blairham/tuikit/theme"
)

// sample is a fresh copy of:
//
//	deploy/web           a
//	├─ rs/web-7d9f       b
//	│  ├─ pod/web-x1     c
//	│  └─ pod/web-y2     d
//	│     └─ container/nginx  e
//	└─ svc/web           f
//	cm/settings          g
func sample() []*Node {
	return []*Node{
		{ID: "a", Label: "deploy/web", Children: []*Node{
			{ID: "b", Label: "rs/web-7d9f", Children: []*Node{
				{ID: "c", Label: "pod/web-x1"},
				{ID: "d", Label: "pod/web-y2", Children: []*Node{
					{ID: "e", Label: "container/nginx"},
				}},
			}},
			{ID: "f", Label: "svc/web"},
		}},
		{ID: "g", Label: "cm/settings"},
	}
}

// open returns a model over sample with every level open.
func open(t *testing.T) *Model {
	t.Helper()
	m := New(theme.Default())
	m.SetDefaultDepth(-1)
	m.SetRoots(sample())
	m.Resize(40, 10)
	return m
}

// ids is the visible rows' IDs, in order.
func ids(m *Model) string {
	out := make([]string, len(m.rows))
	for i, r := range m.rows {
		out[i] = r.node.ID
	}
	return strings.Join(out, " ")
}

// sel is the selected node's ID, or "-" for none.
func sel(m *Model) string {
	n, ok := m.Selected()
	if !ok {
		return "-"
	}
	return n.ID
}

// keys feeds each key and fails on one the tree does not report handled.
func keys(t *testing.T, m *Model, ks ...string) {
	t.Helper()
	for _, k := range ks {
		if !m.HandleKey(k) {
			t.Fatalf("HandleKey(%q) = false", k)
		}
	}
}

func TestDefaultDepth(t *testing.T) {
	for _, tc := range []struct {
		want  string
		depth int
	}{
		{depth: 0, want: "a g"},
		{depth: 1, want: "a b f g"},
		{depth: 2, want: "a b c d f g"},
		{depth: -1, want: "a b c d e f g"},
	} {
		m := New(theme.Default())
		m.SetDefaultDepth(tc.depth)
		m.SetRoots(sample())
		if got := ids(m); got != tc.want {
			t.Errorf("depth %d: rows %q, want %q", tc.depth, got, tc.want)
		}
	}
	m := New(theme.Default())
	m.SetRoots(sample())
	if got := ids(m); got != "a b f g" {
		t.Errorf("New's default depth: rows %q, want the roots open only", got)
	}
}

func TestUpDownHomeEnd(t *testing.T) {
	m := open(t)
	steps := []struct{ key, want string }{
		{"up", "a"}, // already first
		{"down", "b"},
		{"j", "c"},
		{"down", "d"},
		{"k", "c"},
		{"end", "g"},
		{"down", "g"}, // already last
		{"home", "a"},
		{"G", "g"},
		{"g", "a"},
	}
	for _, s := range steps {
		keys(t, m, s.key)
		if got := sel(m); got != s.want {
			t.Fatalf("after %q: selected %q, want %q", s.key, got, s.want)
		}
	}
}

func TestPageKeys(t *testing.T) {
	m := open(t)
	m.Resize(40, 3)
	for _, s := range []struct{ key, want string }{
		{"pgdown", "d"},
		{"ctrl+f", "g"},
		{"pgdown", "g"},
		{"pgup", "d"},
		{"ctrl+b", "a"},
	} {
		keys(t, m, s.key)
		if got := sel(m); got != s.want {
			t.Fatalf("after %q: selected %q, want %q", s.key, got, s.want)
		}
	}
}

// TestRightLeft walks right and left over every kind of row: an open
// parent, a closed parent, a leaf, a root leaf and a closed root.
func TestRightLeft(t *testing.T) {
	m := New(theme.Default())
	m.SetRoots(sample()) // a and g open (g is a leaf); b closed
	steps := []struct {
		key, sel, rows string
	}{
		{"right", "b", "a b f g"},       // a is open: into its first child
		{"right", "b", "a b c d f g"},   // b is closed: open it, stay
		{"l", "c", "a b c d f g"},       // b is open: into c
		{"right", "c", "a b c d f g"},   // c is a leaf: nothing
		{"left", "b", "a b c d f g"},    // c is a leaf: up to b
		{"left", "b", "a b f g"},        // b is open: close it, stay
		{"h", "a", "a b f g"},           // b is closed: up to a
		{"left", "a", "a g"},            // a is open: close it
		{"left", "a", "a g"},            // a is a closed root: nothing
		{"right", "a", "a b f g"},       // a is closed: open it
		{"end", "g", "a b f g"},         //
		{"right", "g", "a b f g"},       // g is a root leaf: nothing
		{"left", "g", "a b f g"},        // and has no parent
		{"up", "f", "a b f g"},          //
		{"left", "a", "a b f g"},        // f is a leaf: up to a
		{"down", "b", "a b f g"},        //
		{"right", "b", "a b c d f g"},   //
		{"down", "c", "a b c d f g"},    //
		{"down", "d", "a b c d f g"},    // d is closed (depth 2)
		{"right", "d", "a b c d e f g"}, // open it
		{"right", "e", "a b c d e f g"}, // into e
		{"left", "d", "a b c d e f g"},  // e is a leaf: up to d
		{"left", "d", "a b c d f g"},    // close d
		{"left", "b", "a b c d f g"},    // d is closed: up to b
		{KeyToggle, "b", "a b f g"},     // space closes b
		{KeyToggle, "b", "a b c d f g"}, // and opens it
		{"down", "c", "a b c d f g"},    //
		{KeyToggle, "c", "a b c d f g"}, // space on a leaf: nothing
		{"home", "a", "a b c d f g"},    //
		{KeyToggle, "a", "a g"},         //
		{KeyToggle, "a", "a b c d f g"}, // b kept its state under a closed a
		{"right", "b", "a b c d f g"},   //
		{"right", "c", "a b c d f g"},   //
		{"down", "d", "a b c d f g"},    //
		{"right", "d", "a b c d e f g"}, //
		{"home", "a", "a b c d e f g"},  //
		{"down", "b", "a b c d e f g"},  //
		{"left", "b", "a b f g"},        //
		{"right", "b", "a b c d e f g"}, // d kept its state under a closed b
	}
	for i, s := range steps {
		keys(t, m, s.key)
		if got := sel(m); got != s.sel {
			t.Fatalf("step %d %q: selected %q, want %q", i, s.key, got, s.sel)
		}
		if got := ids(m); got != s.rows {
			t.Fatalf("step %d %q: rows %q, want %q", i, s.key, got, s.rows)
		}
	}
}

func TestUnboundKeys(t *testing.T) {
	m := open(t)
	keys(t, m, "down")
	for _, k := range []string{"enter", "esc", "/", "x", "q", "ctrl+d", ""} {
		if m.HandleKey(k) {
			t.Errorf("HandleKey(%q) = true; the tree does not use it", k)
		}
	}
	if got := sel(m); got != "b" {
		t.Errorf("unbound keys moved the cursor to %q", got)
	}
	if got := ids(m); got != "a b c d e f g" {
		t.Errorf("unbound keys changed the rows to %q", got)
	}
}

func TestEmptyTree(t *testing.T) {
	m := New(theme.Default())
	m.Resize(10, 3)
	for _, k := range []string{"up", "down", "pgup", "pgdown", "home", "end", "left", "right", KeyToggle} {
		if !m.HandleKey(k) {
			t.Errorf("HandleKey(%q) = false on an empty tree; it is still a tree key", k)
		}
	}
	if _, ok := m.Selected(); ok {
		t.Error("Selected on an empty tree reported a node")
	}
	if p := m.SelectedPath(); p != nil {
		t.Errorf("SelectedPath on an empty tree = %v", p)
	}
	if got := strings.Count(m.View(), "\n") + 1; got != 3 {
		t.Errorf("empty view is %d lines, want the height", got)
	}
	m.ExpandAll()
	m.CollapseAll()
	m.SetFilter("x")
}

func TestSelectedPath(t *testing.T) {
	m := open(t)
	keys(t, m, "down", "down", "down", "down") // e
	path := m.SelectedPath()
	got := make([]string, 0, len(path))
	for _, n := range path {
		got = append(got, n.ID)
	}
	if want := []string{"a", "b", "d", "e"}; !slices.Equal(got, want) {
		t.Errorf("SelectedPath = %v, want %v", got, want)
	}
	keys(t, m, "end")
	if p := m.SelectedPath(); len(p) != 1 || p[0].ID != "g" {
		t.Errorf("SelectedPath on a root = %v, want just the root", p)
	}
}

func TestCounts(t *testing.T) {
	m := New(theme.Default())
	m.SetRoots(sample())
	if got := m.Count(); got != 7 {
		t.Errorf("Count = %d, want 7", got)
	}
	if got := m.VisibleCount(); got != 4 {
		t.Errorf("VisibleCount = %d, want 4", got)
	}
	m.SetRoots([]*Node{nil, {ID: "x", Children: []*Node{nil, {ID: "y"}}}})
	if got := m.Count(); got != 2 {
		t.Errorf("Count with nil nodes = %d, want 2 (nil is skipped)", got)
	}
}

func TestExpandCollapseAll(t *testing.T) {
	m := New(theme.Default())
	m.SetRoots(sample())
	m.ExpandAll()
	if got := ids(m); got != "a b c d e f g" {
		t.Fatalf("ExpandAll: rows %q", got)
	}
	keys(t, m, "down", "down", "down", "down") // e
	m.CollapseAll()
	if got := ids(m); got != "a g" {
		t.Fatalf("CollapseAll: rows %q", got)
	}
	if got := sel(m); got != "a" {
		t.Errorf("CollapseAll: selected %q, want e's root a", got)
	}
}

// TestSetRootsKeepsState refreshes the tree with new copies of the same
// nodes, as a poll does, and checks the user's place survives.
func TestSetRootsKeepsState(t *testing.T) {
	m := New(theme.Default())
	m.SetRoots(sample())
	keys(t, m, "down", "right", "down", "down", "right", "down") // open b, d; on e
	keys(t, m, "home", "down", "down")                           // on c
	m.SetDefaultDepth(-1)                                        // only new nodes see this

	next := sample()
	// A new pod sorts before c, so c's row moves down by one.
	b := next[0].Children[0]
	b.Children = append(
		[]*Node{{ID: "c0", Label: "pod/web-a0", Children: []*Node{{ID: "c0x", Label: "x"}}}},
		b.Children...)
	m.SetRoots(next)
	if got := sel(m); got != "c" {
		t.Errorf("selected %q after a refresh, want c", got)
	}
	if got := ids(m); got != "a b c0 c0x c d e f g" {
		t.Errorf("rows %q: open state lost, or the new node not at the new default depth", got)
	}

	// Close b; it stays closed across another refresh.
	keys(t, m, "left", "left")
	m.SetRoots(sample())
	if got := ids(m); got != "a b f g" {
		t.Errorf("rows %q: b's closed state lost on refresh", got)
	}
	if got := sel(m); got != "b" {
		t.Errorf("selected %q, want b", got)
	}
}

func TestSetRootsRemovedSelection(t *testing.T) {
	m := open(t)
	keys(t, m, "down", "down", "down", "down") // e

	// e goes: its parent d takes the cursor.
	next := sample()
	next[0].Children[0].Children[1].Children = nil
	m.SetRoots(next)
	if got := sel(m); got != "d" {
		t.Fatalf("e removed: selected %q, want its parent d", got)
	}

	// b and everything under it go: up to a.
	next = sample()
	next[0].Children = next[0].Children[1:]
	m.SetRoots(next)
	if got := sel(m); got != "a" {
		t.Fatalf("d's parent removed: selected %q, want the surviving ancestor a", got)
	}

	// The whole root goes: no ancestor left, so the row is clamped.
	keys(t, m, "down") // f, row 1
	m.SetRoots([]*Node{{ID: "z", Label: "z"}})
	if got := sel(m); got != "z" {
		t.Fatalf("root removed: selected %q, want the clamped row z", got)
	}

	m.SetRoots(nil)
	if got := sel(m); got != "-" {
		t.Fatalf("emptied: selected %q", got)
	}
	m.SetRoots(sample())
	if got := sel(m); got != "a" {
		t.Fatalf("refilled: selected %q, want a", got)
	}
}

// TestSetRootsForgetsRemovedIDs checks that state for IDs no longer in the
// tree is dropped, so the map does not grow without bound and a node that
// comes back starts at the default.
func TestSetRootsForgetsRemovedIDs(t *testing.T) {
	m := New(theme.Default())
	m.SetRoots(sample())
	keys(t, m, "left") // close a
	m.SetRoots([]*Node{{ID: "g", Label: "cm/settings"}})
	if _, ok := m.expanded["a"]; ok {
		t.Error("a's state kept after a left the tree")
	}
	m.SetRoots(sample())
	if got := ids(m); got != "a b f g" {
		t.Errorf("rows %q: a came back closed; it should start at the default", got)
	}
}

// TestSetRootsHiddenSelection moves the selected node under a closed
// node: the cursor goes to that node, its new nearest ancestor on screen,
// not to its old parent.
func TestSetRootsHiddenSelection(t *testing.T) {
	m := open(t)
	keys(t, m, "down", "down", "down", "down") // e, under d
	keys(t, m, "home", "down", "down")         // c
	keys(t, m, KeyToggle)                      // c is a leaf: nothing
	keys(t, m, "down", "down")                 // e

	next := sample()
	b := next[0].Children[0]
	b.Children[0].Children = []*Node{{ID: "e", Label: "container/nginx"}} // e moves under c
	b.Children[1].Children = nil
	m.expanded["c"] = false // c, now a parent, is closed
	m.SetRoots(next)
	if got := sel(m); got != "c" {
		t.Errorf("e moved under a closed c: selected %q, want c", got)
	}
}

func TestFilterShowsAncestors(t *testing.T) {
	m := New(theme.Default()) // b, d closed
	m.SetRoots(sample())
	m.Resize(40, 10)
	for _, tc := range []struct {
		expr, want string
	}{
		{"nginx", "a b d e"},
		{"NGINX", "a b d e"},       // case-insensitive
		{"pod/web", "a b c d"},     // d matches, e does not: d drawn as a leaf
		{"!web", "a b d e g"},      // e and g lack "web"; a, b, d lead to e
		{"-f cngx", "a b d e"},     // fuzzy: c-o-n-…-n-g-x
		{"-f setgs", "g"},          //
		{"web-7d9f$", "a b"},       // regex
		{"-l app=web", ""},         // nodes have no labels
		{"nomatch", ""},            //
		{"", "a b f g"},            // the user's state again
		{"rs", "a b"},              // b matches; its children do not
		{"!", "a b f g"},           // a bare "!" is no filter
		{"[", ""},                  // invalid regex: literal "["
		{"deploy|settings", "a g"}, //
		{"container/nginx|svc", "a b d e f"},
	} {
		m.SetFilter(tc.expr)
		if got := ids(m); got != tc.want {
			t.Errorf("filter %q: rows %q, want %q", tc.expr, got, tc.want)
		}
	}
}

func TestFilterStripsANSI(t *testing.T) {
	m := New(theme.Default())
	m.SetDefaultDepth(-1)
	m.SetRoots([]*Node{
		{ID: "a", Label: "\x1b[31mpod\x1b[m/web"},
		{ID: "b", Label: "pod/api"},
	})
	m.SetFilter("pod/web")
	if got := ids(m); got != "a" {
		t.Errorf("filter across an SGR: rows %q, want a", got)
	}
	m.SetFilter("31m")
	if got := ids(m); got != "" {
		t.Errorf("filter matched an escape sequence's bytes: rows %q", got)
	}
}

// TestFilterExpansionIsSeparate opens and closes nodes under a filter and
// checks the user's own state comes back when the filter is cleared.
func TestFilterExpansionIsSeparate(t *testing.T) {
	m := New(theme.Default())
	m.SetRoots(sample()) // a open, b closed
	m.SetFilter("nginx")
	keys(t, m, "down", "left") // close b in the filtered view
	if got := ids(m); got != "a b" {
		t.Fatalf("closing b under a filter: rows %q", got)
	}
	m.SetFilter("")
	if got := ids(m); got != "a b f g" {
		t.Errorf("filter cleared: rows %q, want the user's state", got)
	}
	keys(t, m, "right") // open b for real
	m.SetFilter("nginx")
	if got := ids(m); got != "a b d e" {
		t.Errorf("a new filter starts its view open: rows %q", got)
	}
	m.CollapseAll()
	if got := ids(m); got != "a" {
		t.Errorf("CollapseAll under a filter: rows %q", got)
	}
	m.SetFilter("")
	if got := ids(m); got != "a b c d f g" {
		t.Errorf("CollapseAll under a filter reached the user's state: rows %q", got)
	}
}

func TestFilterMovesCursor(t *testing.T) {
	m := open(t)
	keys(t, m, "end") // g
	m.SetFilter("nginx")
	if got := sel(m); got != "e" {
		t.Errorf("g hidden by the filter, with no ancestor: selected %q, want the row clamped to e", got)
	}
	keys(t, m, "end") // e
	m.SetFilter("pod")
	if got := sel(m); got != "d" {
		t.Errorf("e hidden: selected %q, want its parent d", got)
	}
}

// TestFilterSurvivesSetRoots checks a refresh re-applies the filter.
func TestFilterSurvivesSetRoots(t *testing.T) {
	m := New(theme.Default())
	m.SetRoots(sample())
	m.SetFilter("nginx")
	m.SetRoots(sample())
	if got := ids(m); got != "a b d e" {
		t.Errorf("rows %q after a refresh under a filter", got)
	}
}

func TestScrolling(t *testing.T) {
	m := New(theme.Default())
	roots := make([]*Node, 20)
	for i := range roots {
		roots[i] = &Node{ID: fmt.Sprint(i), Label: fmt.Sprintf("node-%02d", i)}
	}
	m.SetRoots(roots)
	m.Resize(12, 5)
	first := func() string { return strings.TrimSpace(strings.SplitN(ansi.Strip(m.View()), "\n", 2)[0]) }

	for range 4 {
		keys(t, m, "down")
	}
	if got := first(); got != "node-00" {
		t.Errorf("cursor on the last row on screen scrolled to %q", got)
	}
	keys(t, m, "down")
	if got := first(); got != "node-01" {
		t.Errorf("cursor one below the screen: top row %q, want node-01", got)
	}
	keys(t, m, "end")
	if got := first(); got != "node-15" {
		t.Errorf("end: top row %q, want node-15", got)
	}
	keys(t, m, "up", "up", "up", "up")
	if got := first(); got != "node-15" {
		t.Errorf("up within the screen scrolled to %q", got)
	}
	keys(t, m, "up")
	if got := first(); got != "node-14" {
		t.Errorf("up past the top: top row %q, want node-14", got)
	}
	keys(t, m, "pgup")
	if got, s := first(), sel(m); got != "node-09" || s != "9" {
		t.Errorf("pgup: top row %q, cursor %s; want node-09, 9", got, s)
	}
	m.Resize(12, 2)
	keys(t, m, "end")
	m.Resize(12, 8)
	if got := first(); got != "node-12" {
		t.Errorf("grown with the cursor at the end: top row %q, want node-12, no blank rows", got)
	}
	m.SetRoots(roots[:3])
	if got := first(); got != "node-00" {
		t.Errorf("content shrunk: top row %q, want node-00", got)
	}
	keys(t, m, "home")
	if got := first(); got != "node-00" {
		t.Errorf("home: top row %q", got)
	}
}

// TestViewGuides pins the drawn tree: a closing └─ for a last child, ├─
// for the others, │ carried down past an ancestor with siblings below it
// and a blank past one without, and the markers.
func TestViewGuides(t *testing.T) {
	m := New(theme.Default())
	m.SetDefaultDepth(-1)
	roots := sample()
	// A second child under d, and a closed node, to cover every guide.
	d := roots[0].Children[0].Children[1]
	d.Children = append(d.Children, &Node{ID: "e2", Label: "container/sidecar"})
	roots[0].Children = append(
		roots[0].Children,
		&Node{ID: "h", Label: "hpa/web", Children: []*Node{{ID: "i", Label: "x"}}},
	)
	m.SetRoots(roots)
	m.expanded["h"] = false
	m.rebuild()
	m.Resize(40, 12)
	keys(t, m, "end")

	want := []string{
		"▾ deploy/web",
		"├─ ▾ rs/web-7d9f",
		"│  ├─ pod/web-x1",
		"│  └─ ▾ pod/web-y2",
		"│     ├─ container/nginx",
		"│     └─ container/sidecar",
		"├─ svc/web",
		"└─ ▸ hpa/web",
		"cm/settings",
		"", "", "",
	}
	got := strings.Split(ansi.Strip(m.View()), "\n")
	if len(got) != len(want) {
		t.Fatalf("view is %d lines, want %d", len(got), len(want))
	}
	for i := range want {
		if w := ansi.StringWidth(got[i]); w != 40 {
			t.Errorf("line %d is %d cells wide, want the full 40 (painted)", i, w)
		}
		if g := strings.TrimRight(got[i], " "); g != want[i] {
			t.Errorf("line %d = %q, want %q", i, g, want[i])
		}
	}
}

func TestViewSelection(t *testing.T) {
	th := theme.Default()
	m := New(th)
	m.SetRoots([]*Node{{ID: "a", Label: "\x1b[32mone\x1b[m"}, {ID: "b", Label: "\x1b[32mtwo\x1b[m"}})
	m.Resize(20, 2)
	keys(t, m, "down")
	lines := strings.Split(m.View(), "\n")
	selBg := theme.BackgroundSeq(th.Selection)[2:] // "48;2;…m"
	selBg = strings.TrimSuffix(selBg, "m")
	if strings.Contains(lines[0], selBg) {
		t.Errorf("unselected row carries the selection background: %q", lines[0])
	}
	if !strings.Contains(lines[0], "\x1b[32mone") {
		t.Errorf("unselected row lost its label's color: %q", lines[0])
	}
	if !strings.Contains(lines[1], selBg) {
		t.Errorf("selected row lacks the selection background: %q", lines[1])
	}
	if strings.Contains(lines[1], "\x1b[32m") {
		t.Errorf("selected row kept the label's color over the selection: %q", lines[1])
	}
	if w := ansi.StringWidth(lines[1]); w != 20 {
		t.Errorf("selected row is %d cells, want the full width", w)
	}
	// After the label's own reset, the theme's text color and canvas come
	// back before the padding.
	on := openSeq(th.On(th.TableTextColor()))
	if !strings.Contains(lines[0], "one\x1b[m"+on) {
		t.Errorf("text color not re-asserted after the label's reset: %q", lines[0])
	}
}

func TestViewNoPaintBackground(t *testing.T) {
	th := theme.NoPaintBackground()
	m := New(th)
	m.SetRoots([]*Node{{ID: "a", Label: "one"}, {ID: "b", Label: "two"}})
	m.Resize(20, 3)
	lines := strings.Split(m.View(), "\n")
	if bg := theme.BackgroundSeq(th.Bg); strings.Contains(m.View(), strings.TrimSuffix(bg[2:], "m")) {
		t.Errorf("unpainted theme painted the canvas: %q", m.View())
	}
	if w := ansi.StringWidth(lines[1]); w != 3 {
		t.Errorf("unpainted row padded to %d cells", w)
	}
	if lines[2] != "" {
		t.Errorf("unpainted blank row = %q", lines[2])
	}
	if w := ansi.StringWidth(lines[0]); w != 20 {
		t.Errorf("the selected row is %d cells; it spans the width either way", w)
	}
}

func TestViewTruncates(t *testing.T) {
	m := New(theme.Default())
	m.SetDefaultDepth(-1)
	m.SetRoots([]*Node{{ID: "a", Label: "root", Children: []*Node{
		{ID: "b", Label: "\x1b[33mpod/a-very-long-name\x1b[m"},
		{ID: "c", Label: "日本語の長いラベル"},
	}}})
	m.Resize(12, 3)
	got := strings.Split(ansi.Strip(m.View()), "\n")
	want := []string{"▾ root", "├─ pod/a-ve…", "└─ 日本語の…"}
	for i := range want {
		if g := strings.TrimRight(got[i], " "); g != want[i] {
			t.Errorf("line %d = %q, want %q", i, g, want[i])
		}
		if w := ansi.StringWidth(got[i]); w > 12 {
			t.Errorf("line %d is %d cells, over the width", i, w)
		}
	}
	keys(t, m, "down")
	if g := strings.TrimRight(strings.Split(ansi.Strip(m.View()), "\n")[1], " "); g != "├─ pod/a-ve…" {
		t.Errorf("selected row = %q, want it cut the same way", g)
	}
}

func TestViewOneLinePerRow(t *testing.T) {
	m := New(theme.Default())
	m.SetRoots([]*Node{{ID: "a", Label: "one\ntwo\tthree\r"}, {ID: "b", Label: "b"}})
	m.Resize(30, 2)
	got := strings.Split(ansi.Strip(m.View()), "\n")
	if len(got) != 2 {
		t.Fatalf("a label with a line break drew %d lines", len(got))
	}
	if g := strings.TrimRight(got[0], " "); g != "one two three" {
		t.Errorf("row = %q", g)
	}
}

func TestViewBeforeResize(t *testing.T) {
	m := New(theme.Default())
	m.SetRoots(sample())
	if v := m.View(); v != "" {
		t.Errorf("View before Resize = %q, want \"\"", v)
	}
}
