// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package tail

import (
	"regexp"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
)

// sgrReset ends a highlight. What the line had set before it is replayed
// straight after, so the text past a match keeps its own styling.
const sgrReset = "\x1b[m"

// scanned is a styled line split the way [ansi.Strip] splits it: the text a
// search matches against, and where in the line each byte of it sits.
type scanned struct {
	// text is the line's printable bytes, exactly what ansi.Strip returns.
	text string
	// at[k] is the line offset of text's k-th byte.
	at []int
	// ground[i] reports whether the parser is in its ground state before the
	// line's byte i (i == len(line): after the last one) — not inside an
	// escape sequence, nor partway through a multi-byte rune — so an escape
	// sequence may be inserted there.
	ground []bool
	// sgrs are the [start, end) spans of the line's SGR sequences.
	sgrs [][2]int
}

// scan walks line with the state machine ansi.Strip uses, so a match on
// scanned.text is a match on exactly the text Strip would give, and every
// offset maps back into the styled line. Strip is mirrored rather than
// called because the mapping is what highlighting needs.
func scan(line string) scanned {
	sc := scanner{
		line:   line,
		buf:    make([]byte, 0, len(line)),
		seq:    -1,
		pstate: parser.GroundState,
	}
	sc.at = make([]int, 0, len(line))
	sc.ground = make([]bool, 0, len(line)+1)
	for i := range len(line) {
		sc.step(i)
	}
	sc.ground = append(sc.ground, sc.pstate == parser.GroundState)
	sc.text = string(sc.buf)
	return sc.scanned
}

// scanner is [scan]'s state between bytes.
type scanner struct {
	line string
	buf  []byte // text, as it is read
	scanned
	ri, rw int // bytes of the current multi-byte rune read, and its length
	seq    int // where the escape sequence being read began; -1 for none
	pstate parser.State
}

// step reads the line's byte i.
func (sc *scanner) step(i int) {
	ground := sc.pstate == parser.GroundState
	sc.ground = append(sc.ground, ground)
	if sc.pstate == parser.Utf8State {
		sc.keep(i)
		sc.ri++
		if sc.ri >= sc.rw {
			sc.pstate = parser.GroundState
			sc.ri, sc.rw = 0, 0
		}
		return
	}
	state, action := parser.Table.Transition(sc.pstate, sc.line[i])
	switch {
	case action == parser.CollectAction && state == parser.Utf8State:
		sc.rw = utf8ByteLen(sc.line[i])
		sc.ri++
		sc.keep(i)
	case action == parser.PrintAction, action == parser.ExecuteAction:
		sc.keep(i)
	}
	sc.track(i, ground, state)
	sc.pstate = state
}

// keep adds the line's byte i to the text.
func (sc *scanner) keep(i int) {
	sc.buf = append(sc.buf, sc.line[i])
	sc.at = append(sc.at, i)
}

// track follows escape sequences through byte i, which moves the parser to
// state, and records each SGR as it ends.
func (sc *scanner) track(i int, ground bool, state parser.State) {
	switch {
	case state == parser.EscapeState && sc.line[i] == ansi.ESC,
		ground && state != parser.GroundState && state != parser.Utf8State:
		sc.seq = i
	case !ground && state == parser.GroundState && sc.seq >= 0:
		// Only a sequence begun from the ground state stands alone: held
		// back, one that cut short an unfinished escape would leave that
		// escape to swallow what follows.
		if sc.ground[sc.seq] && isSGR(sc.line[sc.seq:i+1]) {
			sc.sgrs = append(sc.sgrs, [2]int{sc.seq, i + 1})
		}
		sc.seq = -1
	}
}

// clusterStarts marks, for each offset into text and its end, whether a
// grapheme cluster starts there, segmented as ansi.StringWidth segments it
// to measure the line.
func clusterStarts(text string) []bool {
	cut := make([]bool, len(text)+1)
	for i := 0; i < len(text); {
		cut[i] = true
		cluster, _ := ansi.FirstGraphemeCluster(text[i:], ansi.GraphemeWidth)
		i += max(len(cluster), 1)
	}
	cut[len(text)] = true
	return cut
}

// utf8ByteLen is the length ansi.Strip gives a rune from its first byte,
// -1 for a byte that cannot start one. Copied so [scan] groups bytes as
// Strip does, invalid UTF-8 included.
func utf8ByteLen(b byte) int {
	switch {
	case b <= 0x7f:
		return 1
	case b >= 0xc0 && b <= 0xdf:
		return 2
	case b >= 0xe0 && b <= 0xef:
		return 3
	case b >= 0xf0 && b <= 0xf7:
		return 4
	}
	return -1
}

// isSGR reports whether seq is a complete CSI ... m sequence carrying only
// numeric parameters: one that sets character attributes and nothing else.
func isSGR(seq string) bool {
	if len(seq) < 3 || seq[0] != ansi.ESC || seq[1] != '[' || seq[len(seq)-1] != 'm' {
		return false
	}
	for _, c := range []byte(seq[2 : len(seq)-1]) {
		if (c < '0' || c > '9') && c != ';' && c != ':' {
			return false
		}
	}
	return true
}

// isSGRReset reports whether the SGR seq resets every attribute.
func isSGRReset(seq string) bool {
	return strings.Trim(seq[2:len(seq)-1], "0") == ""
}

// sgrPrefix returns the escape sequence style opens its text with: its
// colors and attributes, with no padding or other layout. A style that
// adds text of its own around what it renders yields "".
func sgrPrefix(style lipgloss.Style) string {
	r := style.Render("x")
	i := strings.IndexByte(r, 'x')
	if i < 0 || ansi.Strip(r[:i]) != "" {
		return ""
	}
	return r[:i]
}

// highlight wraps every non-empty match of re in line's ANSI-stripped text
// in the escape sequence on. The line's own styling survives: its SGR
// sequences outside a match pass through, those inside one are held back so
// the highlight reads as one piece, and after each match the attributes
// are reset and every SGR the line has set so far is replayed, so the text
// that follows looks as it did. Only SGR sequences are ever added or held
// back, so stripping ANSI from the result gives the text of line back.
//
// A match whose edge falls where nothing can be inserted (inside an escape
// sequence, or partway through a multi-byte rune) is left unhighlighted.
func highlight(line string, re *regexp.Regexp, on string) string {
	if on == "" || re == nil {
		return line
	}
	sc := scan(line)
	spans := matchSpans(sc, re)
	if len(spans) == 0 {
		return line
	}
	w := weaver{line: line, sgrs: sc.sgrs, spans: spans, on: on}
	w.b.Grow(len(line) + len(spans)*(len(on)+len(sgrReset)))
	for i := 0; ; {
		w.edges(i)
		if i == len(line) {
			break
		}
		i = w.copy(i)
	}
	return w.b.String()
}

// matchSpans returns the line offsets [start, end) to highlight for re's
// matches in sc.text: each widened to whole grapheme clusters — an escape
// inside one would split it, and the terminal would draw it, and the
// viewport measure it, as two — and dropped when an edge falls where no
// escape can go.
func matchSpans(sc scanned, re *regexp.Regexp) [][2]int {
	cut := clusterStarts(sc.text)
	var spans [][2]int
	prevEnd := 0
	for _, r := range re.FindAllStringIndex(sc.text, -1) {
		lo, hi := r[0], r[1]
		for lo > 0 && !cut[lo] {
			lo--
		}
		for hi < len(sc.text) && !cut[hi] {
			hi++
		}
		lo = max(lo, prevEnd) // a cluster two matches share goes to the first
		if lo >= hi {
			continue
		}
		prevEnd = hi
		// The highlight opens on the match's first byte and closes right
		// after its last, so escapes either side of it stay outside.
		start, end := sc.at[lo], sc.at[hi-1]+1
		if sc.ground[start] && sc.ground[end] {
			spans = append(spans, [2]int{start, end})
		}
	}
	return spans
}

// weaver writes a line with highlights woven in.
type weaver struct {
	b      strings.Builder
	line   string
	on     string
	sgrs   [][2]int
	spans  [][2]int
	active []string // SGRs set since the line's last full reset
	mi, si int      // the next span and SGR
	in     bool     // inside a span
}

// edges closes the span ending at offset i and opens the one starting there.
func (w *weaver) edges(i int) {
	if w.in && i == w.spans[w.mi][1] {
		w.b.WriteString(sgrReset)
		for _, s := range w.active {
			w.b.WriteString(s)
		}
		w.in = false
		w.mi++
	}
	if !w.in && w.mi < len(w.spans) && i == w.spans[w.mi][0] {
		w.b.WriteString(w.on)
		w.in = true
	}
}

// copy writes what starts at offset i — an SGR, held back inside a span, or
// one byte — and returns the offset after it.
func (w *weaver) copy(i int) int {
	if w.si >= len(w.sgrs) || w.sgrs[w.si][0] != i {
		w.b.WriteByte(w.line[i])
		return i + 1
	}
	end := w.sgrs[w.si][1]
	w.si++
	s := w.line[i:end]
	if isSGRReset(s) {
		w.active = w.active[:0]
	} else {
		w.active = append(w.active, s)
	}
	if !w.in {
		w.b.WriteString(s)
	}
	return end
}
