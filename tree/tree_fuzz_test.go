// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package tree

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/blairham/tuikit/table"
	"github.com/blairham/tuikit/theme"
)

// fuzzKeys are the keys a fuzzed byte picks from.
var fuzzKeys = []string{"up", "down", "pgup", "pgdown", "home", "end", "left", "right", KeyToggle, "enter"}

// FuzzTree feeds two labels, as an app might write them (styled, wide,
// broken), a filter expression as typed at the prompt, a width and a run of
// keys. Whatever they are, nothing panics, and:
//
//   - the filter keeps exactly the nodes that match, ANSI stripped, or have
//     a descendant that does — every one of them on screen, since a new
//     filter opens the path to each match — and the row under the cursor
//     is one of them;
//   - the view is exactly the height in lines, none of them wider than the
//     width once stripped, in a painted theme and an unpainted one, before
//     the keys and after.
func FuzzTree(f *testing.F) {
	for _, s := range []struct {
		a    string
		b    string
		expr string
		keys []byte
		w    uint8
	}{
		{a: "deploy/web", b: "pod/web-1", expr: "web", w: 20, keys: []byte{1, 1, 6, 7}},
		{a: "\x1b[32mgreen\x1b[m", b: "plain", expr: "!green", w: 8, keys: []byte{5, 6, 6, 8}},
		{a: "日本語のラベル", b: "x", expr: "-f 本ラ", w: 5, keys: []byte{1, 7, 1}},
		{a: "a\nb\tc\rd", b: "e", expr: "", w: 3, keys: []byte{1, 1, 1}},
		{a: "\x1b[", b: "\x1b]8;;u\x07link\x1b]8;;\x07", expr: "[", w: 1, keys: nil},
		{a: "e\u0301\u0301", b: "\xff\xfe", expr: "\xff", w: 2, keys: []byte{8, 8}},
		{a: "long-label-with-no-end", b: "", expr: "-l app=x", w: 40, keys: []byte{3, 2}},
	} {
		f.Add(s.a, s.b, s.expr, s.w, s.keys)
	}
	f.Fuzz(func(t *testing.T, a, b, expr string, w uint8, ks []byte) {
		width := int(w%48) + 1
		filter := table.ParseFilter(expr)
		for _, th := range []theme.Theme{theme.Default(), theme.NoPaintBackground()} {
			m := New(th)
			m.SetDefaultDepth(1)
			m.SetRoots(fuzzRoots(a, b))
			m.Resize(width, 4)
			m.SetFilter(expr)
			checkFilter(t, m, filter)
			checkView(t, m, width, 4)
			for _, k := range ks {
				m.HandleKey(fuzzKeys[int(k)%len(fuzzKeys)])
			}
			m.SetRoots(fuzzRoots(b, a))
			checkView(t, m, width, 4)
		}
	})
}

// fuzzRoots builds a three-level tree over two labels.
func fuzzRoots(a, b string) []*Node {
	return []*Node{
		{ID: "r", Label: a, Children: []*Node{
			{ID: "x", Label: b, Children: []*Node{{ID: "y", Label: a + b}}},
			{ID: "z", Label: b + a},
		}},
		{ID: "s", Label: b},
	}
}

func checkFilter(t *testing.T, m *Model, filter table.RowFilter) {
	t.Helper()
	var keep func(*Node) bool
	keep = func(n *Node) bool {
		if filter.Empty() || filter.MatchesAny(ansi.Strip(n.Label)) {
			return true
		}
		for _, c := range n.Children {
			if keep(c) {
				return true
			}
		}
		return false
	}
	shown := map[string]bool{}
	for _, r := range m.rows {
		shown[r.node.ID] = true
		if !keep(r.node) {
			t.Fatalf("row %q (%q) shown; neither it nor a descendant matches", r.node.ID, r.node.Label)
		}
	}
	if filter.Empty() {
		return
	}
	var walk func([]*Node)
	walk = func(nodes []*Node) {
		for _, n := range nodes {
			if keep(n) && !shown[n.ID] {
				t.Fatalf("node %q (%q) kept by the filter but not on screen", n.ID, n.Label)
			}
			walk(n.Children)
		}
	}
	walk(m.roots)
	if n, ok := m.Selected(); ok && !shown[n.ID] {
		t.Fatalf("selected %q is not a row", n.ID)
	}
}

func checkView(t *testing.T, m *Model, width, height int) {
	t.Helper()
	lines := strings.Split(m.View(), "\n")
	if len(lines) != height {
		t.Fatalf("view is %d lines, want %d", len(lines), height)
	}
	for i, l := range lines {
		if got := ansi.StringWidth(ansi.Strip(l)); got > width {
			t.Fatalf("line %d is %d cells, over the width %d: %q", i, got, width, l)
		}
	}
}
