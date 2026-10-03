// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package table

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// RowFilter is a compiled, optionally negated case-insensitive regex
// filter. A leading "!" negates the match: rows passing the underlying
// regex are excluded.
//
// Invalid regex syntax falls back to a literal substring match (with
// the input quoted via [regexp.QuoteMeta]), so users can type "$" or
// "[" without crashing the filter. Bytes that are not valid UTF-8 become
// U+FFFD first: the regexp package rejects them in a pattern, and reads
// them as U+FFFD in the text it matches, so a filter pasted from a field
// with a stray byte still matches that field.
type RowFilter struct {
	re     *regexp.Regexp
	negate bool
}

// ParseFilter compiles a filter expression. An empty string returns an
// empty filter ([RowFilter.Empty] returns true).
func ParseFilter(s string) RowFilter {
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
	return RowFilter{re: re, negate: negate}
}

// Empty reports whether the filter is a no-op (everything passes).
func (f RowFilter) Empty() bool { return f.re == nil }

// MatchesAny returns true if any field passes the filter. For negated
// filters, returns true only when no field matches.
func (f RowFilter) MatchesAny(fields ...string) bool {
	if f.re == nil {
		return true
	}
	for _, field := range fields {
		if f.re.MatchString(field) {
			return !f.negate
		}
	}
	return f.negate
}
