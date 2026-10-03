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

func TestParseFilter_Kind(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want FilterKind
	}{
		{"", FilterNone},
		{"!", FilterNone},
		{"-f", FilterNone},
		{"-f   ", FilterNone},
		{"-l", FilterNone},
		{"-l  , ,", FilterNone},
		{"nginx", FilterRegex},
		{"!nginx", FilterRegex},
		{"-foo", FilterRegex}, // no space after -f: still a regex
		{"-lib", FilterRegex},
		{"!-f x", FilterRegex}, // the prefix must come first
		{" -f x", FilterRegex},
		{"-f x", FilterFuzzy},
		{"-f\tx", FilterFuzzy},
		{"-l app=web", FilterLabel},
		{"-l =web", FilterLabel}, // malformed, but still a label filter
	}
	for _, tc := range tests {
		f := ParseFilter(tc.in)
		if got := f.Kind(); got != tc.want {
			t.Errorf("ParseFilter(%q).Kind() = %v; want %v", tc.in, got, tc.want)
		}
		if f.Empty() != (tc.want == FilterNone) {
			t.Errorf("ParseFilter(%q).Empty() = %v; want %v", tc.in, f.Empty(), tc.want == FilterNone)
		}
	}
}

func TestRowFilter_Fuzzy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		filter string
		fields []string
		want   bool
	}{
		{"-f ngx", []string{"nginx-7d9f"}, true},
		{"-f NGX", []string{"nginx-7d9f"}, true},  // case-insensitive, term side
		{"-f ngx", []string{"NGINX-7D9F"}, true},  // case-insensitive, field side
		{"-f xgn", []string{"nginx-7d9f"}, false}, // order matters
		{"-f nnn", []string{"nginx"}, false},      // each rune is used once
		{"-f nn", []string{"nginx"}, true},
		{"-f ngx", []string{"web", "nginx"}, true}, // any field
		{"-f ngx", []string{"ng", "x"}, false},     // within one field, not across
		{"-f ngx", nil, false},
		{"-f   ngx  ", []string{"nginx"}, true}, // term is trimmed
		{"-f n x", []string{"nginx"}, false},    // inner space is part of the term
		{"-f n x", []string{"nginx x"}, true},
		{"-f .*", []string{"nginx"}, false}, // not a regex
		{"-f .*", []string{"a.b*"}, true},
		{"-f ΣΣ", []string{"σας"}, true},      // Unicode folding: Σ folds to σ and to final ς
		{"-f k", []string{"K"}, true},         // Kelvin sign folds to k
		{"-f \xff", []string{"a\xfeb"}, true}, // stray bytes read as U+FFFD on both sides
		{"-f", []string{"anything"}, true},
	}
	for _, tc := range tests {
		f := ParseFilter(tc.filter)
		if got := f.MatchesAny(tc.fields...); got != tc.want {
			t.Errorf("ParseFilter(%q).MatchesAny(%q) = %v; want %v", tc.filter, tc.fields, got, tc.want)
		}
		if got := f.Match(tc.fields, map[string]string{"ngx": "ngx"}); got != tc.want {
			t.Errorf("ParseFilter(%q).Match(%q, labels) = %v; want %v", tc.filter, tc.fields, got, tc.want)
		}
	}
}

func TestRowFilter_Labels(t *testing.T) {
	t.Parallel()
	labels := map[string]string{"app": "web", "tier": "front", "empty": "", "eq": "a=b"}
	tests := []struct {
		filter string
		want   bool
	}{
		{"-l app=web", true},
		{"-l app==web", true},
		{"-l app=api", false},
		{"-l app=WEB", false}, // labels compare case-sensitively
		{"-l App=web", false},
		{"-l missing=web", false},
		{"-l app!=api", true},
		{"-l app!=web", false},
		{"-l missing!=web", true}, // absent counts as different
		{"-l app", true},
		{"-l missing", false},
		{"-l !missing", true},
		{"-l !app", false},
		{"-l empty", true},
		{"-l empty=", true},
		{"-l app=", false},
		{"-l missing=", false}, // absent is not equal to ""
		{"-l app=web,tier=front", true},
		{"-l app=web,tier=back", false}, // AND
		{"-l app=web,!missing,tier", true},
		{"-l  app = web ,  tier != back ", true}, // spaces trimmed
		{"-l ! missing", true},
		{"-l app=web,,", true},  // empty terms skipped
		{"-l ,", true},          // no terms: empty filter
		{"-l app=web=x", false}, // value is "web=x"
		{"-l eq=a=b", true},     // only the first "=" is the operator
		// Malformed: the whole filter matches nothing.
		{"-l =web", false},
		{"-l !", false},
		{"-l !=web", false},
		{"-l !app=web", false},
		{"-l app=web,=x", false},
		{"-l app in (web,api)", false},
		{"-l !!app", false},
		{"-l a b", false},
	}
	for _, tc := range tests {
		f := ParseFilter(tc.filter)
		if got := f.Match([]string{"app=web"}, labels); got != tc.want {
			t.Errorf("ParseFilter(%q).Match(labels) = %v; want %v", tc.filter, got, tc.want)
		}
	}
}

func TestRowFilter_LabelsWithoutLabels(t *testing.T) {
	t.Parallel()
	// A row with no labels: nil and an empty map agree.
	for _, labels := range []map[string]string{nil, {}} {
		for filter, want := range map[string]bool{
			"-l app":      false,
			"-l app=web":  false,
			"-l !app":     true,
			"-l app!=web": true,
			"-l =web":     false,
		} {
			if got := ParseFilter(filter).Match(nil, labels); got != want {
				t.Errorf("ParseFilter(%q).Match(nil, %#v) = %v; want %v", filter, labels, got, want)
			}
		}
	}
	// MatchesAny is given no labels, so a label filter matches nothing —
	// even one that a row without labels would pass.
	for _, filter := range []string{"-l app", "-l !app", "-l app!=web"} {
		if ParseFilter(filter).MatchesAny("app", "web", "app=web") {
			t.Errorf("ParseFilter(%q).MatchesAny matched; a label filter needs Match", filter)
		}
	}
}

func TestRowFilter_MatchLegacy(t *testing.T) {
	t.Parallel()
	// The plain form ignores labels: Match is MatchesAny on the fields.
	labels := map[string]string{"oms": "oms"}
	for filter, want := range map[string]bool{
		"oms":   true,
		"!oms":  false,
		"zzz":   false,
		"!zzz":  true,
		"[":     false,
		"":      true,
		"-l x":  false,
		"-f om": true,
	} {
		if got := ParseFilter(filter).Match([]string{"trading", "oms-1"}, labels); got != want {
			t.Errorf("ParseFilter(%q).Match = %v; want %v", filter, got, want)
		}
	}
}
