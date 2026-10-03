// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package table

import "testing"

func TestParseFilter_Empty(t *testing.T) {
	t.Parallel()
	for _, in := range []string{"", "!"} {
		f := ParseFilter(in)
		if !f.Empty() {
			t.Errorf("ParseFilter(%q) should be Empty()", in)
		}
	}
}

func TestRowFilter_MatchesAny(t *testing.T) {
	t.Parallel()
	tests := []struct {
		filter string
		fields []string
		want   bool
	}{
		{"oms", []string{"trading", "oms-7bb44f869-zwbst"}, true},
		{"OMS", []string{"trading", "oms-7bb44f869-zwbst"}, true}, // case-insensitive
		{"missing", []string{"trading", "oms-7bb44f869-zwbst"}, false},
		{"!oms", []string{"trading", "oms-7bb44f869-zwbst"}, false}, // negated
		{"!missing", []string{"trading", "oms-7bb44f869-zwbst"}, true},
		{"trad.*", []string{"trading", "x"}, true},
		{"[", []string{"foo[bar"}, true}, // invalid regex → literal → matches
		{"", []string{"anything"}, true},
	}
	for _, tc := range tests {
		f := ParseFilter(tc.filter)
		got := f.MatchesAny(tc.fields...)
		if got != tc.want {
			t.Errorf("ParseFilter(%q).MatchesAny(%v) = %v; want %v", tc.filter, tc.fields, got, tc.want)
		}
	}
}

func TestParseFilterInvalidUTF8(t *testing.T) {
	t.Parallel()
	// Each used to panic inside the literal fallback's MustCompile.
	for _, expr := range []string{"\xff", "!\xd6", "a\x99b", "[\xff"} {
		f := ParseFilter(expr)
		if f.Empty() {
			t.Fatalf("ParseFilter(%q) is empty", expr)
		}
	}
	// A stray byte in the filter matches the same stray byte in a field,
	// and nothing else.
	f := ParseFilter("pod-\xff")
	if !f.MatchesAny("pod-\xff-1") {
		t.Error(`"pod-\xff" does not match the field it came from`)
	}
	if f.MatchesAny("pod-a") {
		t.Error(`"pod-\xff" matches "pod-a"`)
	}
}
