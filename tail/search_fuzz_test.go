// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package tail

import (
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"

	"github.com/blairham/tuikit/table"
)

// FuzzSearch feeds a search expression, as typed at the prompt, and a line
// of styled text, as a container or a colorizer would write it. Whatever the
// two are, searching must not panic, and highlighting may only ever add or
// hold back styling: stripping ANSI from a highlighted line gives the line's
// text back, and so does stripping a rendered view, wrapped or not, compared
// with the same view unsearched. A line matches exactly when the filter
// built from the same expression would keep it, so the two never disagree
// about what a pattern means; that is checked for the filter's plain regex
// form without its leading "!" — search takes "!", "-f" and "-l" literally.
//
// The view is compared only for a line whose escape sequences are all SGR,
// the kind search adds and holds back, in valid UTF-8 with no control
// characters but "\n". The viewport and lipgloss treat anything else — an
// unterminated string sequence, a C1 control, a sequence left open at the
// end of the line, a "\v" or a "\r" — differently from ansi.Strip, which
// keeps every control character as text: they fold a "\r" ending one line
// into the "\n" joining it to the next, and drop some controls on render.
// Such a line's view differs from its text with no search at all, and an
// inserted escape can close what it left open or keep a "\r" from folding.
// The line on its own is checked for every input.
func FuzzSearch(f *testing.F) {
	for _, s := range [][2]string{
		{"hit", "a hit b"},
		{"key", "\x1b[32m\"key\": 1\x1b[m"},
		{"abc", "a\x1b[31mb\x1b[mc"},
		{"32m", "\x1b[32mx"},
		{"[", "a[b"},
		{"!x", "!x"},
		{"x*", "yyy"},
		{"語", "日本語\x1b[1m語"},
		{"e", "e\u0301\te"},
		{"a", "\x1b[3\ta2ma"},
		{"\xff", "\xe6a\xffa"},
		{"link", "\x1b]8;;u\x07link\x1b]8;;\x07"},
		{"b", "a\nb\r\nb"},
		{".", "\x1b["},
		// Found by fuzzing: an SGR that cuts short an unfinished escape; a
		// match ending inside a grapheme cluster; controls the viewport
		// treats differently from ansi.Strip; an unterminated SOS string.
		{"ABC", "A\x1b\x1b[mBC"},
		{"0", "0\u0a03"},
		{`\S`, "00000000000\v "},
		{".", "\r"},
		{"k.Y", "\x1bX\xf4000\xf4000k0Y000"},
	} {
		f.Add(s[0], s[1])
	}
	f.Fuzz(func(t *testing.T, expr, text string) {
		re := compileSearch(expr)
		for _, on := range []string{testOn, testCur} {
			if got := highlight(text, re, on); ansi.Strip(got) != ansi.Strip(text) {
				t.Fatalf("highlight(%q, %q) = %q: changed the text", text, expr, got)
			}
		}

		if !sgrOnly(text) {
			return
		}
		for _, wrap := range []bool{false, true} {
			plain, searched := New(), New()
			for _, m := range []*Model{plain, searched} {
				m.SetWrap(wrap)
				m.AppendLines([]string{text, "after"})
				m.Resize(12, 4)
				m.HandleScrollKey("g")
			}
			searched.SetSearch(expr)
			searched.NextMatch()
			searched.PrevMatch()
			plain.viewport.SetYOffset(searched.viewport.YOffset())
			plain.viewport.SetXOffset(searched.viewport.XOffset())
			if a, b := ansi.Strip(plain.View()), ansi.Strip(searched.View()); a != b {
				t.Fatalf("wrap=%v: searching %q changed the text of %q:\n%q\n%q", wrap, expr, text, a, b)
			}
		}

		filter := table.ParseFilter(expr)
		if filter.Kind() != table.FilterRegex || strings.HasPrefix(expr, "!") {
			return
		}
		m := New()
		m.AppendLine(text)
		m.SetSearch(expr)
		if want := filter.MatchesAny(ansi.Strip(text)); (m.Matches() == 1) != want {
			t.Fatalf("search %q on %q: %d matches; the filter keeps it: %v", expr, text, m.Matches(), want)
		}
	})
}

// sgrOnly reports whether every escape sequence in text is a complete SGR
// sequence, in valid UTF-8 with no control characters but "\n" and the
// escapes.
func sgrOnly(text string) bool {
	if !utf8.ValidString(text) {
		return false
	}
	for _, r := range text {
		if r != '\n' && r != ansi.ESC && unicode.IsControl(r) {
			return false
		}
	}
	sc := scan(text)
	return sc.ground[len(text)] && strings.Count(text, "\x1b") == len(sc.sgrs)
}
