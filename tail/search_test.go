// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package tail

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// hitLines returns n lines "line 00".."line nn", with " hit" on those in hits.
func hitLines(n int, hits ...int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("line %02d", i)
	}
	for _, h := range hits {
		out[h] += " hit"
	}
	return out
}

// topRow is the first row of the rendered view, stripped.
func topRow(m *Model) string {
	return ansi.Strip(strings.SplitN(m.View(), "\n", 2)[0])
}

// step is one move of a navigation table and what it must leave behind.
type step struct {
	move  func(*Model) bool
	name  string
	top   string // the view's first row must contain this
	index int
}

func next(m *Model) bool { return m.NextMatch() }
func prev(m *Model) bool { return m.PrevMatch() }

func TestSearchNextPrevWrapAround(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		steps []step
		// rows scrolled down from the top before searching
		scroll int
	}{
		{
			name: "next from the top, around the end and back",
			steps: []step{
				{next, "n", "line 00", 1}, // line 02 is already on screen
				{next, "n", "line 05", 2},
				{next, "n", "line 08", 3}, // line 09, as far as the view scrolls
				{next, "n wraps", "line 02", 1},
				{prev, "N wraps", "line 08", 3},
				{prev, "N", "line 05", 2},
			},
		},
		{
			name: "prev first goes to the last match",
			steps: []step{
				{prev, "N", "line 08", 3},
				{prev, "N", "line 05", 2},
			},
		},
		{
			name:   "first next starts at the top of the view",
			scroll: 4,
			steps: []step{
				{next, "n", "line 04", 2}, // line 05, below the top row
				{next, "n", "line 08", 3},
				{next, "n wraps", "line 02", 1},
			},
		},
		{
			name:   "first prev starts above the top of the view",
			scroll: 6,
			steps: []step{
				{prev, "N", "line 05", 2}, // line 05 is above the top row
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := New()
			m.AppendLines(hitLines(12, 2, 5, 9))
			m.Resize(20, 4)
			m.HandleScrollKey("g")
			for range tc.scroll {
				m.HandleScrollKey("j")
			}
			m.follow = true // set directly, not to jump: a move must pause it
			m.SetSearch("HIT")
			if m.Matches() != 3 || m.MatchIndex() != 0 {
				t.Fatalf("after SetSearch: %d/%d, want 0/3", m.MatchIndex(), m.Matches())
			}
			for i, s := range tc.steps {
				if !s.move(m) {
					t.Fatalf("step %d (%s) returned false", i, s.name)
				}
				if m.MatchIndex() != s.index {
					t.Errorf("step %d (%s): MatchIndex %d, want %d", i, s.name, m.MatchIndex(), s.index)
				}
				if got := topRow(m); !strings.Contains(got, s.top) {
					t.Errorf("step %d (%s): top row %q, want %q", i, s.name, got, s.top)
				}
				if m.Follow() {
					t.Errorf("step %d (%s): follow still on", i, s.name)
				}
			}
		})
	}
}

func TestSearchNoMatches(t *testing.T) {
	t.Parallel()
	for _, expr := range []string{"", "absent"} {
		m := New()
		m.AppendLines(hitLines(5))
		m.Resize(20, 3)
		before := m.View()
		m.SetSearch(expr)
		if m.NextMatch() || m.PrevMatch() {
			t.Errorf("%q: a move reported a match", expr)
		}
		if m.Matches() != 0 || m.MatchIndex() != 0 {
			t.Errorf("%q: %d/%d, want 0/0", expr, m.MatchIndex(), m.Matches())
		}
		if !m.Follow() {
			t.Errorf("%q: a move with nothing to move to paused follow", expr)
		}
		if m.View() != before {
			t.Errorf("%q: the view changed:\n%q\n%q", expr, before, m.View())
		}
	}
}

func TestSearchExprAndBang(t *testing.T) {
	t.Parallel()
	m := New()
	m.AppendLines([]string{"!x", "x", "y", "a[b"})
	m.SetSearch("!x")
	if m.SearchExpr() != "!x" || m.Matches() != 1 {
		t.Errorf("!x: expr %q, %d matches; want the literal line only", m.SearchExpr(), m.Matches())
	}
	m.SetSearch("[") // bad syntax: a literal
	if m.Matches() != 1 {
		t.Errorf("[: %d matches, want 1", m.Matches())
	}
	m.SetSearch("")
	if m.SearchExpr() != "" || m.Matches() != 0 {
		t.Errorf("cleared: expr %q, %d matches", m.SearchExpr(), m.Matches())
	}
}

// TestSearchMatchesFollowTheBuffer moves to a match, changes the buffer, and
// checks the count and that the current match is still the same line.
func TestSearchMatchesFollowTheBuffer(t *testing.T) {
	t.Parallel()
	cases := []struct {
		change    func(*Model)
		name      string
		curLine   string // the current match's line afterwards; "" for none
		matches   int
		wantIndex int
	}{
		{
			name:      "append",
			change:    func(m *Model) { m.AppendLines([]string{"new hit", "new"}) },
			matches:   4,
			wantIndex: 2,
			curLine:   "line 05 hit",
		},
		{
			name:      "prepend",
			change:    func(m *Model) { m.PrependLines([]string{"old hit", "old", "old hit"}) },
			matches:   5,
			wantIndex: 4,
			curLine:   "line 05 hit",
		},
		{
			name: "replace keeps the line where it still matches",
			change: func(m *Model) {
				m.ReplaceLines(append(hitLines(8, 1, 5), "extra hit"))
			},
			matches:   3,
			wantIndex: 2,
			curLine:   "line 05 hit",
		},
		{
			name:    "replace drops it where it no longer matches",
			change:  func(m *Model) { m.ReplaceLines(hitLines(8, 1, 6)) },
			matches: 2,
		},
		{
			name:      "trim above it",
			change:    func(m *Model) { m.SetMaxLines(8) },
			matches:   2,
			wantIndex: 1,
			curLine:   "line 05 hit",
		},
		{
			name:    "trim it",
			change:  func(m *Model) { m.SetMaxLines(4) },
			matches: 1,
		},
		{
			name:      "append past the cap",
			change:    func(m *Model) { m.SetMaxLines(12); m.AppendLines(hitLines(3, 0)) },
			matches:   3,
			wantIndex: 1,
			curLine:   "line 05 hit",
		},
		{
			name:      "filter keeping it",
			change:    func(m *Model) { m.SetFilter("hit|line 0[0-4]") },
			matches:   3,
			wantIndex: 2,
			curLine:   "line 05 hit",
		},
		{
			name:    "filter hiding it",
			change:  func(m *Model) { m.SetFilter("!05") },
			matches: 2,
		},
		{
			name:   "clear",
			change: func(m *Model) { m.Clear() },
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			for _, ready := range []bool{false, true} {
				m := New()
				m.AppendLines(hitLines(12, 2, 5, 9))
				if ready {
					m.Resize(30, 4)
					m.HandleScrollKey("g")
				}
				m.SetSearch("hit")
				m.NextMatch()
				m.NextMatch()
				if m.MatchIndex() != 2 {
					t.Fatalf("setup: MatchIndex %d, want 2", m.MatchIndex())
				}
				tc.change(m)
				if m.Matches() != tc.matches || m.MatchIndex() != tc.wantIndex {
					t.Errorf("ready=%v: %d/%d, want %d/%d",
						ready, m.MatchIndex(), m.Matches(), tc.wantIndex, tc.matches)
				}
				if got := currentLine(m); got != tc.curLine {
					t.Errorf("ready=%v: current line %q, want %q", ready, got, tc.curLine)
				}
			}
		})
	}
}

// currentLine is the text of the current match's line, "" with none.
func currentLine(m *Model) string {
	if m.MatchIndex() == 0 {
		return ""
	}
	return m.visible[m.matches[m.MatchIndex()-1].vis]
}

// TestSetSearchStartsOver: a new search, even the same expression again,
// has no current match until the next move.
func TestSetSearchStartsOver(t *testing.T) {
	t.Parallel()
	m := New()
	m.AppendLines(hitLines(6, 1, 3))
	m.SetSearch("hit")
	m.NextMatch()
	m.NextMatch()
	for _, expr := range []string{"hit", "line 03"} {
		m.SetSearch(expr)
		if m.MatchIndex() != 0 {
			t.Errorf("SetSearch(%q) kept current match %d", expr, m.MatchIndex())
		}
	}
}

func TestSearchCombinedWithFilter(t *testing.T) {
	t.Parallel()
	m := New()
	m.AppendLines([]string{"err db hit", "info db", "err net", "err db hit again", "info db hit"})
	m.SetFilter("err")
	m.SetSearch("db")
	if m.Matches() != 2 {
		t.Fatalf("db within err: %d matches, want 2", m.Matches())
	}
	m.SetFilter("")
	if m.Matches() != 4 {
		t.Errorf("filter cleared: %d matches, want 4", m.Matches())
	}
	m.SetFilter("!hit")
	if m.Matches() != 1 {
		t.Errorf("negated filter: %d matches, want 1", m.Matches())
	}
	if m.VisibleCount() != 2 {
		t.Errorf("search hid lines: %d visible, want 2", m.VisibleCount())
	}
}

const (
	testOn  = "\x1b[7m"
	testCur = "\x1b[1;7m"
)

func TestHighlightPreservesStyling(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, line, expr, want string
	}{
		{
			name: "plain",
			line: "a hit b hit",
			expr: "hit",
			want: "a " + testOn + "hit" + sgrReset + " b " + testOn + "hit" + sgrReset,
		},
		{
			name: "colored around the match is restored after it",
			line: "\x1b[32m\"key\": 1\x1b[m tail",
			expr: "key",
			want: "\x1b[32m\"" + testOn + "key" + sgrReset + "\x1b[32m\": 1\x1b[m tail",
		},
		{
			name: "every SGR since the last reset is replayed",
			line: "\x1b[1m\x1b[0;31mred \x1b[4mhit\x1b[m",
			expr: "red",
			want: "\x1b[1m\x1b[0;31m" + testOn + "red" + sgrReset + "\x1b[1m\x1b[0;31m \x1b[4mhit\x1b[m",
		},
		{
			name: "styling inside a match is held back and replayed after",
			line: "a\x1b[31mb\x1b[mc\x1b[32md",
			expr: "abc",
			want: testOn + "abc" + sgrReset + "\x1b[32md",
		},
		{
			name: "a match across styling, ending in it",
			line: "x\x1b[33mhi\x1b[mt y",
			expr: "hit",
			want: "x\x1b[33m" + testOn + "hit" + sgrReset + " y",
		},
		{
			name: "the escape's own text does not match",
			line: "\x1b[32mgreen\x1b[m",
			expr: "32m",
			want: "\x1b[32mgreen\x1b[m",
		},
		{
			name: "wide runes",
			line: "日本 語hit",
			expr: "語h",
			want: "日本 " + testOn + "語h" + sgrReset + "it",
		},
		{
			name: "a match inside a grapheme cluster widens to all of it",
			line: "cafe\u0301 x",
			expr: "\u0301",
			want: "caf" + testOn + "e\u0301" + sgrReset + " x",
		},
		{
			name: "a match ending inside a grapheme cluster widens to its end",
			line: "0\u0a03 x",
			expr: "0",
			want: testOn + "0\u0a03" + sgrReset + " x",
		},
		{
			name: "non-SGR escapes pass through",
			line: "\x1b]8;;http://x\x07link\x1b]8;;\x07",
			expr: "link",
			want: "\x1b]8;;http://x\x07" + testOn + "link" + sgrReset + "\x1b]8;;\x07",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := highlight(tc.line, compileSearch(tc.expr), testOn)
			if got != tc.want {
				t.Errorf("highlight(%q, %q)\n got  %q\n want %q", tc.line, tc.expr, got, tc.want)
			}
			if ansi.Strip(got) != ansi.Strip(tc.line) {
				t.Errorf("text changed: %q -> %q", ansi.Strip(tc.line), ansi.Strip(got))
			}
		})
	}
}

// TestViewHighlightsMatches renders through the model: matches carry the
// match style, the current match's line the current style, and the text is
// unchanged.
func TestViewHighlightsMatches(t *testing.T) {
	t.Parallel()
	for _, wrap := range []bool{false, true} {
		m := New()
		m.SetWrap(wrap)
		m.SetSearchStyles(lipgloss.NewStyle().Reverse(true), lipgloss.NewStyle().Bold(true).Reverse(true))
		lines := []string{"\x1b[32mfirst hit\x1b[m", "plain", "second hit here and a long tail to wrap", "x"}
		m.AppendLines(lines)
		m.Resize(20, 8)
		plain := ansi.Strip(m.View())
		m.SetSearch("hit")
		m.NextMatch()
		m.NextMatch()
		v := m.View()
		if ansi.Strip(v) != plain {
			t.Errorf("wrap=%v: text changed\n%q\n%q", wrap, plain, ansi.Strip(v))
		}
		if !strings.Contains(v, "\x1b[32mfirst "+testOn+"hit") {
			t.Errorf("wrap=%v: first line not highlighted as a match: %q", wrap, v)
		}
		if !strings.Contains(v, "second "+testCur+"hit") {
			t.Errorf("wrap=%v: current line not highlighted as current: %q", wrap, v)
		}
		if strings.Contains(v, testCur+"hit\x1b[m\x1b[32m") || strings.Count(v, testCur+"hit") != 1 {
			t.Errorf("wrap=%v: current style on the wrong line: %q", wrap, v)
		}
		if got := m.VisibleLines(); strings.Join(got, "\n") != strings.Join(lines, "\n") {
			t.Errorf("wrap=%v: VisibleLines carries the highlight: %q", wrap, got)
		}
	}
}

// TestSearchRevealsUnderWrap scrolls to a match below several wrapped lines:
// the offset is in rows, which a line count would undershoot.
func TestSearchRevealsUnderWrap(t *testing.T) {
	t.Parallel()
	m := New()
	m.SetWrap(true)
	long := strings.Repeat("abcdefghij", 3) // three rows at width 10
	m.AppendLines([]string{long, long, long, long, "target hit"})
	m.Resize(10, 2)
	m.HandleScrollKey("g")
	m.SetSearch("hit")
	if !m.NextMatch() {
		t.Fatal("no match")
	}
	if !strings.Contains(m.View(), "hit") {
		t.Errorf("match not on screen under wrap: %q", m.View())
	}
}

// TestSearchRevealsPastTheRightEdge scrolls sideways to a match beyond the
// viewport's width when lines are not wrapped.
func TestSearchRevealsPastTheRightEdge(t *testing.T) {
	t.Parallel()
	m := New()
	m.AppendLines([]string{strings.Repeat(".", 40) + "hit"})
	m.Resize(10, 1)
	m.SetSearch("hit")
	m.NextMatch()
	if !strings.Contains(m.View(), "hit") {
		t.Errorf("match not on screen: %q", m.View())
	}
}

func TestSgrPrefix(t *testing.T) {
	t.Parallel()
	if got := sgrPrefix(lipgloss.NewStyle().Reverse(true)); got != testOn {
		t.Errorf("reverse: %q", got)
	}
	if got := sgrPrefix(lipgloss.NewStyle().Reverse(true).Padding(0, 1)); got != "" {
		t.Errorf("padded: %q, want none", got)
	}
}

// compileSearch must agree with the filter on every expression the filter
// does not negate: one rule for what a pattern means.
func TestCompileSearchMatchesTheFilter(t *testing.T) {
	t.Parallel()
	re := compileSearch("A.c")
	if !re.MatchString("xabcx") || re.String() != regexp.MustCompile("(?i)A.c").String() {
		t.Errorf("A.c: %v", re)
	}
	if compileSearch("\xffz").String() != "(?i)�z" {
		t.Errorf("invalid UTF-8 not replaced: %v", compileSearch("\xffz"))
	}
}
