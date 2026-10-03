// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package table

import (
	"strings"
	"testing"
)

// FuzzParseFilter feeds the filter bar's raw input to ParseFilter. Whatever
// the user types, parsing must not panic — invalid regex falls back to a
// literal — and a leading "!" must select exactly the rows the bare
// expression rejects. Only one "!" is special: "!" alone is a no-op and
// "!!x" negates the literal-or-regex "!x", so the complement is checked for
// expressions that do not already start with one. "-f" and "-l" select
// other modes, which have no "!" form, so "!-f x" is always the plain
// regex and the complement is checked only where expr is plain too.
func FuzzParseFilter(f *testing.F) {
	for _, s := range []string{"", "!", "nginx", "!nginx", "[", "!$", "(?i)a|b", `\`, "a{2,", "!!x", "-f x", "-l a=b", "-l", "-foo"} {
		f.Add(s, "nginx-7d9f")
	}
	f.Fuzz(func(t *testing.T, expr, field string) {
		plain := ParseFilter(expr)
		negated := ParseFilter("!" + expr)
		if expr == "" {
			if !plain.Empty() || !negated.Empty() {
				t.Fatalf("empty expression %q gave a non-empty filter", expr)
			}
			return
		}
		if expr[0] == '!' {
			return
		}
		if negated.Kind() != FilterRegex {
			t.Fatalf("!%q parsed as kind %v; a leading \"!\" is always the plain form", expr, negated.Kind())
		}
		if plain.Kind() != FilterRegex {
			return
		}
		// With no fields nothing can match, which pins the complement to
		// the right way round.
		if plain.MatchesAny() || !negated.MatchesAny() {
			t.Fatalf("%q with no fields: plain %v, negated %v; want false, true",
				expr, plain.MatchesAny(), negated.MatchesAny())
		}
		if negated.MatchesAny(field) == plain.MatchesAny(field) {
			t.Fatalf("!%q and %q agree on %q: negation must be the complement", expr, expr, field)
		}
	})
}

// FuzzParseFilterModes feeds arbitrary "-f" and "-l" input. Parsing and
// matching must not panic; fuzzy matching a field against itself always
// succeeds; and "-l k=v" matches a row labeled exactly k: v, for any key
// and value a selector can spell.
func FuzzParseFilterModes(f *testing.F) {
	for _, s := range []string{"", "x", "app=web", "a!=b,c", "!a", "=", "!", ",,", "a==", "Σ", "\xff", " k = v "} {
		f.Add(s, "web", "nginx-7d9f")
	}
	f.Fuzz(func(t *testing.T, expr, value, field string) {
		labels := map[string]string{expr: value}
		for _, mode := range []string{"-f ", "-l ", "-f", "-l"} {
			flt := ParseFilter(mode + expr)
			flt.MatchesAny(field)
			flt.Match([]string{field}, labels)
			flt.Match(nil, nil)
		}

		if self := ParseFilter("-f " + field); !self.MatchesAny(field) {
			t.Fatalf("-f %q does not match itself", field)
		}

		key := strings.TrimSpace(expr)
		if !validLabelKey(key) || strings.Contains(key, ",") ||
			strings.TrimSpace(value) != value || strings.Contains(value, ",") ||
			strings.HasPrefix(value, "=") {
			return
		}
		sel := ParseFilter("-l " + key + "=" + value)
		if sel.Kind() != FilterLabel {
			t.Fatalf("-l %s=%s parsed as kind %v", key, value, sel.Kind())
		}
		if !sel.Match(nil, map[string]string{key: value}) {
			t.Fatalf("-l %s=%s does not match {%q: %q}", key, value, key, value)
		}
		if sel.MatchesAny(key, value, key+"="+value) {
			t.Fatalf("-l %s=%s matched fields without labels", key, value)
		}
	})
}
