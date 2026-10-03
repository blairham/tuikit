// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package table

import "testing"

// FuzzParseFilter feeds the filter bar's raw input to ParseFilter. Whatever
// the user types, parsing must not panic — invalid regex falls back to a
// literal — and a leading "!" must select exactly the rows the bare
// expression rejects. Only one "!" is special: "!" alone is a no-op and
// "!!x" negates the literal-or-regex "!x", so the complement is checked for
// expressions that do not already start with one.
func FuzzParseFilter(f *testing.F) {
	for _, s := range []string{"", "!", "nginx", "!nginx", "[", "!$", "(?i)a|b", `\`, "a{2,", "!!x"} {
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
