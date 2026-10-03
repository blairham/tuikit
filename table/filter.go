// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package table

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// FilterKind names the mode a [RowFilter] was parsed in.
type FilterKind int

const (
	// FilterNone is the empty filter: everything passes.
	FilterNone FilterKind = iota
	// FilterRegex is the plain form: a case-insensitive regex, "!" to negate.
	FilterRegex
	// FilterFuzzy is "-f term": a case-insensitive subsequence match.
	FilterFuzzy
	// FilterLabel is "-l selector": a label selector over a row's labels.
	FilterLabel
)

// labelOp is the comparison one label-selector term makes.
type labelOp int

const (
	labelEquals labelOp = iota
	labelNotEquals
	labelExists
	labelNotExists
)

// labelTerm is one comma-separated term of a "-l" selector.
type labelTerm struct {
	key   string
	value string
	op    labelOp
}

// RowFilter is a parsed filter-bar expression, in one of three modes,
// chosen as k9s chooses them:
//
//   - "-f term" is fuzzy. A field matches when the term's characters
//     appear in it in order, case-insensitively — "ngx" matches
//     "nginx-7d9f". Spaces around the term are trimmed; "-f" with no term
//     is an empty filter.
//   - "-l selector" is a label selector: comma-separated terms, all of
//     which must hold. "k=v" and "k==v" need label k to equal v; "k!=v"
//     needs it to differ or be absent; "k" needs it to exist; "!k" needs
//     it absent. Spaces around terms, keys and values are trimmed, and
//     keys and values compare case-sensitively. Empty terms (a stray
//     comma) are skipped, and a selector with no terms is an empty
//     filter. A malformed term — no key, or a key holding a space, "="
//     or "!", as in "=v", "!" or "app in (a" — makes the whole filter
//     match nothing, rather than quietly widening it.
//   - Anything else is the plain form: a case-insensitive regex. A leading
//     "!" negates the match: rows passing the underlying regex are
//     excluded.
//
// A prefix counts only when "-f" or "-l" is followed by a space or ends
// the input, so "-foo" is still the regex "-foo". The prefix must come
// first: "!-f x" is the negated regex "-f x".
//
// In the plain form, invalid regex syntax falls back to a literal
// substring match (with the input quoted via [regexp.QuoteMeta]), so
// users can type "$" or "[" without crashing the filter. Bytes that are
// not valid UTF-8 become U+FFFD first: the regexp package rejects them in
// a pattern, and reads them as U+FFFD in the text it matches, so a filter
// pasted from a field with a stray byte still matches that field. Fuzzy
// matching reads stray bytes as U+FFFD on both sides the same way.
type RowFilter struct {
	re        *regexp.Regexp
	fuzzy     []rune
	terms     []labelTerm
	kind      FilterKind
	negate    bool
	malformed bool
}

// ParseFilter compiles a filter expression. An empty string returns an
// empty filter ([RowFilter.Empty] returns true).
func ParseFilter(s string) RowFilter {
	if rest, ok := cutMode(s, "-f"); ok {
		return parseFuzzy(rest)
	}
	if rest, ok := cutMode(s, "-l"); ok {
		return parseLabels(rest)
	}
	return parseRegex(s)
}

// cutMode strips a mode prefix that is followed by whitespace or ends s.
func cutMode(s, prefix string) (string, bool) {
	rest, ok := strings.CutPrefix(s, prefix)
	if !ok {
		return "", false
	}
	if rest == "" {
		return rest, true
	}
	r, _ := utf8.DecodeRuneInString(rest)
	if !unicode.IsSpace(r) {
		return "", false
	}
	return rest, true
}

func parseRegex(s string) RowFilter {
	if s == "" {
		return RowFilter{}
	}
	negate := strings.HasPrefix(s, "!")
	if negate {
		s = s[1:]
	}
	if s == "" {
		return RowFilter{}
	}
	s = strings.ToValidUTF8(s, string(utf8.RuneError))
	re, err := regexp.Compile("(?i)" + s)
	if err != nil {
		re = regexp.MustCompile("(?i)" + regexp.QuoteMeta(s))
	}
	return RowFilter{kind: FilterRegex, re: re, negate: negate}
}

func parseFuzzy(s string) RowFilter {
	s = strings.TrimSpace(s)
	if s == "" {
		return RowFilter{}
	}
	// Ranging over a string yields U+FFFD for each invalid byte, which is
	// how the field side is read too.
	term := make([]rune, 0, len(s))
	for _, r := range s {
		term = append(term, r)
	}
	return RowFilter{kind: FilterFuzzy, fuzzy: term}
}

func parseLabels(s string) RowFilter {
	var terms []labelTerm
	malformed := false
	for raw := range strings.SplitSeq(s, ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		t, ok := parseLabelTerm(raw)
		if !ok {
			malformed = true
			continue
		}
		terms = append(terms, t)
	}
	if !malformed && len(terms) == 0 {
		return RowFilter{}
	}
	return RowFilter{kind: FilterLabel, terms: terms, malformed: malformed}
}

// parseLabelTerm parses one trimmed, non-empty selector term.
func parseLabelTerm(raw string) (labelTerm, bool) {
	var t labelTerm
	if i := strings.IndexByte(raw, '='); i >= 0 {
		key, value := raw[:i], raw[i+1:]
		switch {
		case strings.HasSuffix(key, "!"):
			key, t.op = key[:len(key)-1], labelNotEquals
		case strings.HasPrefix(value, "="):
			value, t.op = value[1:], labelEquals
		default:
			t.op = labelEquals
		}
		t.key, t.value = strings.TrimSpace(key), strings.TrimSpace(value)
	} else if key, ok := strings.CutPrefix(raw, "!"); ok {
		t.key, t.op = strings.TrimSpace(key), labelNotExists
	} else {
		t.key, t.op = raw, labelExists
	}
	return t, validLabelKey(t.key)
}

// validLabelKey reports whether k can name a label in a selector term.
func validLabelKey(k string) bool {
	return k != "" && !strings.ContainsFunc(k, func(r rune) bool {
		return r == '=' || r == '!' || unicode.IsSpace(r)
	})
}

// Empty reports whether the filter is a no-op (everything passes).
func (f RowFilter) Empty() bool { return f.kind == FilterNone }

// Kind reports which mode the filter was parsed in. An app whose rows have
// no labels can check for [FilterLabel] to explain why nothing matches.
func (f RowFilter) Kind() FilterKind { return f.kind }

// MatchesAny returns true if any field passes the filter. For negated
// filters, returns true only when no field matches.
//
// A label filter needs labels, which MatchesAny is not given, so it
// matches nothing here; apps whose rows carry labels call [RowFilter.Match].
func (f RowFilter) MatchesAny(fields ...string) bool {
	switch f.kind {
	case FilterNone:
		return true
	case FilterFuzzy:
		for _, field := range fields {
			if fuzzyMatch(f.fuzzy, field) {
				return true
			}
		}
		return false
	case FilterLabel:
		return false
	}
	for _, field := range fields {
		if f.re.MatchString(field) {
			return !f.negate
		}
	}
	return f.negate
}

// Match reports whether a row passes the filter. Regex and fuzzy filters
// test the fields, as [RowFilter.MatchesAny] does, and ignore labels; a
// label filter tests the labels and ignores the fields. A nil labels map
// is a row with no labels, so "!k" and "k!=v" pass it and "k" does not.
func (f RowFilter) Match(fields []string, labels map[string]string) bool {
	if f.kind != FilterLabel {
		return f.MatchesAny(fields...)
	}
	if f.malformed {
		return false
	}
	for _, t := range f.terms {
		v, ok := labels[t.key]
		var pass bool
		switch t.op {
		case labelEquals:
			pass = ok && v == t.value
		case labelNotEquals:
			pass = !ok || v != t.value
		case labelExists:
			pass = ok
		case labelNotExists:
			pass = !ok
		}
		if !pass {
			return false
		}
	}
	return true
}

// fuzzyMatch reports whether term's runes appear in field in order,
// ignoring case.
func fuzzyMatch(term []rune, field string) bool {
	i := 0
	for _, r := range field {
		if i == len(term) {
			break
		}
		if foldEqual(r, term[i]) {
			i++
		}
	}
	return i == len(term)
}

// foldEqual reports whether a and b are equal under simple Unicode case
// folding, as [strings.EqualFold] compares runes.
func foldEqual(a, b rune) bool {
	if a == b {
		return true
	}
	for r := unicode.SimpleFold(a); r != a; r = unicode.SimpleFold(r) {
		if r == b {
			return true
		}
	}
	return false
}
