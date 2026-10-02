package table

import (
	"slices"
	"testing"

	"charm.land/bubbles/v2/table"
)

func firstCol(rows []table.Row) []string {
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = r[0]
	}
	return out
}

func TestSorter_SortNaturalAndDirection(t *testing.T) {
	t.Parallel()
	rows := []table.Row{{"10"}, {"9"}, {"beta"}, {"Alpha"}, {"100"}}
	s := NewSorter(0, false)
	s.Sort(rows)
	// Numbers compare numerically ("10" after "9"); text case-insensitively.
	if got, want := firstCol(rows), []string{"9", "10", "100", "Alpha", "beta"}; !slices.Equal(got, want) {
		t.Fatalf("ascending = %v, want %v", got, want)
	}
	idx := func(v string) int { return slices.Index(firstCol(rows), v) }
	if idx("9") > idx("10") || idx("10") > idx("100") {
		t.Errorf("numeric cells out of order: %v", firstCol(rows))
	}
	if idx("Alpha") > idx("beta") {
		t.Errorf("text should compare case-insensitively: %v", firstCol(rows))
	}

	s.Reverse()
	s.Sort(rows)
	if idx("100") > idx("10") || idx("beta") > idx("Alpha") {
		t.Errorf("descending sort not reversed: %v", firstCol(rows))
	}
}

func TestSorter_SortStableAndStripsANSI(t *testing.T) {
	t.Parallel()
	rows := []table.Row{
		{"b", "1"},
		{"\x1b[31ma\x1b[m", "2"},
		{"b", "3"},
		{"a", "4"},
	}
	NewSorter(0, false).Sort(rows)
	order := make([]string, 0, len(rows))
	for _, r := range rows {
		order = append(order, r[1])
	}
	if want := []string{"2", "4", "1", "3"}; !slices.Equal(order, want) {
		t.Errorf("order = %v, want %v (ties keep input order; ANSI ignored)", order, want)
	}
}

func TestSorter_SortShortRowsAndCustomCompare(t *testing.T) {
	t.Parallel()
	rows := []table.Row{{"x", "b"}, {"y"}, {"z", "a"}}
	s := NewSorter(1, false)
	s.Sort(rows)
	if got := firstCol(rows); !slices.Equal(got, []string{"y", "z", "x"}) {
		t.Errorf("short row should sort as empty: %v", got)
	}

	var sawCol int
	s.SetCompare(func(col int, a, b string) int {
		sawCol = col
		return len(a) - len(b)
	})
	rows = []table.Row{{"", "ccc"}, {"", "a"}, {"", "bb"}}
	s.Sort(rows)
	if sawCol != 1 || rows[0][1] != "a" || rows[2][1] != "ccc" {
		t.Errorf("custom compare not used: col=%d rows=%v", sawCol, rows)
	}
}

func TestSorter_HandleKey(t *testing.T) {
	t.Parallel()
	s := NewSorter(0, true)
	if s.HandleKey("left", 3) {
		t.Error("a plain arrow is not a sort key")
	}
	if !s.HandleKey(KeySortPrev, 3) || s.Column() != 0 || !s.Descending() {
		t.Errorf("shift+← at the first column should be a handled no-op: col=%d desc=%v", s.Column(), s.Descending())
	}
	if !s.HandleKey(KeySortNext, 3) || s.Column() != 1 || s.Descending() {
		t.Errorf("shift+→ should move to column 1 ascending: col=%d desc=%v", s.Column(), s.Descending())
	}
	s.HandleKey(KeySortNext, 3)
	s.HandleKey(KeySortNext, 3)
	if s.Column() != 2 {
		t.Errorf("shift+→ should stop at the last column, got %d", s.Column())
	}
	if !s.HandleKey(KeySortNext, 0) || s.Column() != 2 {
		t.Error("with no columns a sort key is handled and changes nothing")
	}
}

func TestSorter_SortBy(t *testing.T) {
	t.Parallel()
	s := NewSorter(0, false)
	s.SortBy(2)
	if s.Column() != 2 || s.Descending() {
		t.Fatalf("SortBy(new) should select it ascending: col=%d desc=%v", s.Column(), s.Descending())
	}
	s.SortBy(2)
	if !s.Descending() {
		t.Error("SortBy(active) should reverse the direction")
	}
	s.SortBy(-1)
	if s.Column() != 2 {
		t.Error("SortBy(-1) should be ignored")
	}
}

func TestSorter_Columns(t *testing.T) {
	t.Parallel()
	cols := []table.Column{{Title: "NAME", Width: 10}, {Title: "AGE", Width: 5}}
	s := NewSorter(1, false)
	got := s.Columns(cols)
	if got[1].Title != "AGE"+SortAscIndicator || got[0].Title != "NAME" {
		t.Errorf("ascending header = %q %q", got[0].Title, got[1].Title)
	}
	if cols[1].Title != "AGE" {
		t.Error("Columns must not modify its input")
	}
	s.Reverse()
	if got := s.Columns(cols); got[1].Title != "AGE"+SortDescIndicator {
		t.Errorf("descending header = %q", got[1].Title)
	}
	if got := NewSorter(5, false).Columns(cols); got[0].Title != "NAME" || got[1].Title != "AGE" {
		t.Error("an out-of-range sort column should decorate nothing")
	}
}

func TestSorter_ZeroValue(t *testing.T) {
	t.Parallel()
	var s Sorter
	rows := []table.Row{{"b"}, {"a"}}
	s.Sort(rows)
	if rows[0][0] != "a" {
		t.Errorf("zero Sorter should sort column 0 ascending: %v", rows)
	}
}
