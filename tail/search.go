// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package tail

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// match is a visible line the search matches.
type match struct {
	buf int // index into Model.lines
	vis int // index into Model.visible
}

// compileSearch compiles a search expression the way table.ParseFilter
// compiles a plain regex filter, minus negation: case-insensitive, invalid
// UTF-8 as U+FFFD, and a literal match when the expression is not valid
// regex syntax. "" is no search.
func compileSearch(expr string) *regexp.Regexp {
	if expr == "" {
		return nil
	}
	expr = strings.ToValidUTF8(expr, string(utf8.RuneError))
	re, err := regexp.Compile("(?i)" + expr)
	if err != nil {
		re = regexp.MustCompile("(?i)" + regexp.QuoteMeta(expr))
	}
	return re
}

// SetSearch searches the lines the view shows for expr, k9s's describe-view
// search: unlike [Model.SetFilter] it hides nothing. [Model.View]
// highlights each match, and [Model.NextMatch] and [Model.PrevMatch] move
// between the lines that have one. The filter still applies, and search runs
// over what it lets through.
//
// expr is compiled like a plain regex filter: a case-insensitive regex,
// matched against each line with its ANSI styling stripped, falling back to
// a literal match when it is not valid syntax; bytes that are not valid
// UTF-8 become U+FFFD. A leading "!" is not negation here but part of the
// pattern — a search that highlighted everything except its matches would
// highlight nothing — and the filter's "-f" and "-l" modes are not search
// modes: "-f x" searches for that text. "" clears the search.
//
// A new search has no current match, and moves nothing: the first
// [Model.NextMatch] goes to the first match at or below the top of the view.
func (m *Model) SetSearch(expr string) {
	m.searchExpr = expr
	m.search = compileSearch(expr)
	m.curBuf = -1
	m.rematch()
	m.redraw()
}

// SearchExpr returns the expression last passed to [Model.SetSearch].
func (m *Model) SearchExpr() string { return m.searchExpr }

// Matches returns how many visible lines the search matches; 0 with no
// search. A line with several matches counts once: [Model.NextMatch] moves
// by line.
func (m *Model) Matches() int { return len(m.matches) }

// MatchIndex returns the 1-based position of the current match among
// [Model.Matches], 0 when there is none — no search, nothing matching, or no
// move made since [Model.SetSearch]. With Matches it makes the "3/17" a
// search prompt shows.
//
// The current match stays on its line as lines are appended, prepended,
// replaced or trimmed and as the filter changes, for as long as that line is
// still in the buffer, still shown, and still matches. Once it is not, there
// is no current match until the next move.
func (m *Model) MatchIndex() int {
	if len(m.matches) == 0 {
		return 0
	}
	return m.current + 1
}

// NextMatch makes the next matching line the current match, wrapping from the
// last to the first, and scrolls it into view. With no current match it
// starts from the first match at or below the top of the view. Like any
// scroll it pauses follow. It reports false, and does nothing, when no line
// matches.
func (m *Model) NextMatch() bool {
	n := len(m.matches)
	if n == 0 {
		return false
	}
	i := 0
	if m.current >= 0 {
		i = (m.current + 1) % n
	} else {
		top := m.topLine()
		for j, mt := range m.matches {
			if mt.vis >= top {
				i = j
				break
			}
		}
	}
	m.selectMatch(i)
	return true
}

// PrevMatch makes the previous matching line the current match, wrapping from
// the first to the last, and scrolls it into view. With no current match it
// starts from the last match above the top of the view. Like any scroll it
// pauses follow. It reports false, and does nothing, when no line matches.
func (m *Model) PrevMatch() bool {
	n := len(m.matches)
	if n == 0 {
		return false
	}
	i := n - 1
	if m.current >= 0 {
		i = (m.current - 1 + n) % n
	} else {
		top := m.topLine()
		for j := n - 1; j >= 0; j-- {
			if m.matches[j].vis < top {
				i = j
				break
			}
		}
	}
	m.selectMatch(i)
	return true
}

// SetSearchStyles sets how [Model.View] highlights search matches: match for
// every match, current for the matches on the current match's line. Pass the
// theme's [theme.Theme.SearchMatch] and [theme.Theme.SearchCurrent]; [New]
// starts from those of [theme.Default]. Only a style's colors and attributes
// are used — the highlight sits inside a line, so padding, borders and width
// do not apply, and a style that renders any of them highlights nothing.
func (m *Model) SetSearchStyles(match, current lipgloss.Style) {
	m.matchOn = sgrPrefix(match)
	m.currentOn = sgrPrefix(current)
	m.redraw()
}

// selectMatch makes matches[i] current, pauses follow, and shows it.
func (m *Model) selectMatch(i int) {
	m.current = i
	m.curBuf = m.matches[i].buf
	m.follow = false
	if !m.ready {
		return
	}
	m.viewport.SetContent(m.content())
	m.reveal(m.matches[i].vis)
}

// redraw re-renders the content after a change to the search or its styles,
// which leaves the lines, and so the scroll position, as they were.
func (m *Model) redraw() {
	if m.ready {
		m.viewport.SetContent(m.content())
	}
}

// rematch recomputes the matching lines. Every change to the buffer or the
// filter calls it, before any early return for an unready viewport, so
// [Model.Matches] is right before the first [Model.Resize] too. The current
// match is kept by its buffer index, which the changes that shift lines
// (prepend, trim) adjust.
func (m *Model) rematch() {
	m.matches = m.matches[:0]
	m.current = -1
	if m.search == nil {
		m.curBuf = -1
		return
	}
	vis := 0
	for i, l := range m.lines {
		if !m.shows(l) {
			continue
		}
		if m.search.MatchString(ansi.Strip(l.text)) {
			if i == m.curBuf {
				m.current = len(m.matches)
			}
			m.matches = append(m.matches, match{buf: i, vis: vis})
		}
		vis++
	}
	if m.current < 0 {
		m.curBuf = -1
	}
}

// content is the viewport's content: the visible lines, with the search's
// matches highlighted. Call [Model.rematch] first when the lines changed.
func (m *Model) content() string {
	if len(m.matches) == 0 {
		return strings.Join(m.visible, "\n")
	}
	lines := append([]string(nil), m.visible...)
	for j, mt := range m.matches {
		on := m.matchOn
		if j == m.current {
			on = m.currentOn
		}
		lines[mt.vis] = highlight(lines[mt.vis], m.search, on)
	}
	return strings.Join(lines, "\n")
}

// rows returns how many viewport rows text takes: one per line of it (the
// viewport splits a line holding "\n"), each soft-wrapped onto as many rows as
// its width needs when wrap is on — the arithmetic the viewport scrolls by.
func (m *Model) rows(text string) int {
	n := 0
	for sub := range strings.SplitSeq(text, "\n") {
		n += m.wrappedRows(sub)
	}
	return n
}

// wrappedRows is how many rows one viewport line takes.
func (m *Model) wrappedRows(sub string) int {
	w := m.viewport.Width()
	if !m.wrap || w <= 0 {
		return 1
	}
	return max(1, (ansi.StringWidth(sub)+w-1)/w)
}

// topLine returns the index of the visible line at the top of the view.
func (m *Model) topLine() int {
	if !m.ready {
		return 0
	}
	y, row := m.viewport.YOffset(), 0
	for i, text := range m.visible {
		row += m.rows(text)
		if row > y {
			return i
		}
	}
	return len(m.visible)
}

// reveal scrolls the view, if it must, so the first match on visible line
// vis is on screen: vertically always, and horizontally too when lines are
// not wrapped and the match lies past the right edge.
func (m *Model) reveal(vis int) {
	row := 0
	for _, text := range m.visible[:vis] {
		row += m.rows(text)
	}
	stripped := ansi.Strip(m.visible[vis])
	var colStart, colEnd int
	if loc := m.search.FindStringIndex(stripped); loc != nil {
		// The match may sit on a later line of a line holding "\n", or a
		// later wrapped row of a long one.
		subs := strings.Split(stripped[:loc[0]], "\n")
		for _, sub := range subs[:len(subs)-1] {
			row += m.wrappedRows(sub)
		}
		prefix := subs[len(subs)-1]
		if m.wrap {
			if w := m.viewport.Width(); w > 0 {
				row += ansi.StringWidth(prefix) / w
			}
		} else {
			colStart = ansi.StringWidth(prefix)
			colEnd = colStart + ansi.StringWidth(stripped[loc[0]:loc[1]])
		}
	}
	m.viewport.EnsureVisible(row, colStart, colEnd)
}
