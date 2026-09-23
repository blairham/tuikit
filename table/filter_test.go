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
