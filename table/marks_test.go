package table

import (
	"slices"
	"strings"
	"testing"

	"charm.land/bubbles/v2/table"
	"github.com/charmbracelet/x/ansi"

	"github.com/blairham/tuikit/theme"
)

func names(rows []table.Row) []string { return firstCol(rows) }

func rowsOf(ns ...string) []table.Row {
	rows := make([]table.Row, len(ns))
	for i, n := range ns {
		rows[i] = table.Row{n, "x"}
	}
	return rows
}

func TestMarks_ToggleSurvivesReorder(t *testing.T) {
	t.Parallel()
	m := NewMarks(nil)
	rows := rowsOf("a", "b", "c")
	if !m.HandleKey(KeyMarkToggle, rows, 1) {
		t.Fatal("space should be handled")
	}
	slices.Reverse(rows) // c b a — a re-sort
	rows = append(rows, table.Row{"d", "x"})
	if got := names(m.Selected(rows, 0)); !slices.Equal(got, []string{"b"}) {
		t.Errorf("mark did not follow its row: %v", got)
	}
	m.HandleKey(KeyMarkToggle, rows, 1) // b again
	if m.Len() != 0 {
		t.Error("toggling a marked row should unmark it")
	}
}

func TestMarks_Range(t *testing.T) {
	t.Parallel()
	m := NewMarks(nil)
	rows := rowsOf("a", "b", "c", "d", "e")
	m.HandleKey(KeyMarkToggle, rows, 3)
	m.HandleKey(KeyMarkRange, rows, 1)
	if got := names(m.Selected(rows, 0)); !slices.Equal(got, []string{"b", "c", "d"}) {
		t.Errorf("range upward = %v", got)
	}
	// The cursor is the new anchor: extend from b down to e.
	m.HandleKey(keyMarkRangeNUL, rows, 4)
	if got := names(m.Selected(rows, 0)); !slices.Equal(got, []string{"b", "c", "d", "e"}) {
		t.Errorf("range from new anchor = %v", got)
	}

	fresh := NewMarks(nil)
	fresh.MarkRange(rows, 2)
	if got := names(fresh.Selected(rows, 0)); !slices.Equal(got, []string{"c"}) {
		t.Errorf("range with no anchor should mark the cursor row: %v", got)
	}
}

func TestMarks_ClearAndSelectedFallback(t *testing.T) {
	t.Parallel()
	m := NewMarks(nil)
	rows := rowsOf("a", "b")
	m.HandleKey(KeyMarkToggle, rows, 0)
	m.HandleKey(KeyMarkToggle, rows, 1)
	if !m.HandleKey(KeyMarkClear, rows, 0) || m.Len() != 0 {
		t.Fatal("ctrl+\\ should clear every mark")
	}
	if got := names(m.Selected(rows, 1)); !slices.Equal(got, []string{"b"}) {
		t.Errorf("with no marks Selected is the cursor row: %v", got)
	}
	if m.Selected(nil, 0) != nil {
		t.Error("no rows, no selection")
	}
	if m.HandleKey("x", rows, 0) {
		t.Error("x is not a mark key")
	}
	if !m.HandleKey(KeyMarkToggle, rows, 9) || m.Len() != 0 {
		t.Error("space off the rows is handled and marks nothing")
	}
}

func TestMarks_PruneAndCustomKey(t *testing.T) {
	t.Parallel()
	m := NewMarks(func(r table.Row) string { return r[0] + "/" + r[1] })
	rows := []table.Row{{"ns1", "web"}, {"ns2", "web"}}
	m.Toggle(rows[0])
	if m.IsMarked(rows[1]) {
		t.Fatal("custom key should tell namespaces apart")
	}
	m.Prune(rows[1:])
	if m.Len() != 0 {
		t.Error("Prune should drop marks whose rows are gone")
	}
	// The anchor went with it: a range now marks only the cursor row.
	m.MarkRange(rows, 1)
	if m.Len() != 1 || !m.IsMarked(rows[1]) {
		t.Error("pruned anchor should not anchor a range")
	}
}

func TestMarks_Style(t *testing.T) {
	t.Parallel()
	th := theme.Default()
	m := NewMarks(nil)
	rows := []table.Row{{"\x1b[31ma\x1b[m", "1"}, {"b", "2"}}
	if got := m.Style(rows, th); &got[0] != &rows[0] {
		t.Error("with no marks Style should return rows as they are")
	}
	m.Toggle(rows[0])
	got := m.Style(rows, th)
	want := th.MarkStyle.Render("a")
	if got[0][0] != want || got[0][1] != th.MarkStyle.Render("1") {
		t.Errorf("marked row not restyled: %q", got[0])
	}
	if strings.Contains(got[0][0], "31m") {
		t.Error("the cell's own styling should be replaced")
	}
	if got[1][0] != "b" {
		t.Errorf("unmarked row changed: %q", got[1])
	}
	if ansi.Strip(rows[0][0]) != "a" || rows[0][0] == got[0][0] {
		t.Error("Style must not modify its input")
	}
	if !m.IsMarked(got[0]) {
		t.Error("the default key should still match a styled row")
	}
}

func TestMarks_UnmarkingTheAnchorDropsIt(t *testing.T) {
	t.Parallel()
	m := NewMarks(nil)
	rows := rowsOf("a", "b", "c", "d")
	m.HandleKey(KeyMarkToggle, rows, 1)
	m.HandleKey(KeyMarkToggle, rows, 1) // b unmarked: no anchor left
	m.HandleKey(KeyMarkRange, rows, 3)
	if got := names(m.Selected(rows, 0)); !slices.Equal(got, []string{"d"}) {
		t.Errorf("an unmarked row still anchored the range: %v", got)
	}
}
