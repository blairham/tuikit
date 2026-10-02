package table

import (
	"testing"

	"charm.land/bubbles/v2/table"
)

func TestPlainText(t *testing.T) {
	t.Parallel()
	cols := []table.Column{{Title: "NAME", Width: 3}, {Title: "\x1b[1mAGE\x1b[m", Width: 3}, {Title: "ST", Width: 2}}
	rows := []table.Row{
		{"web-frontend", "5m", "\x1b[32mok\x1b[m"},
		{"db", "日本", ""},
		{"short"},
		{"a", "b", "c", "extra"},
	}
	want := "NAME          AGE   ST\n" +
		"web-frontend  5m    ok\n" +
		"db            日本\n" +
		"short\n" +
		"a             b     c\n"
	if got := PlainText(cols, rows); got != want {
		t.Errorf("PlainText =\n%s\nwant\n%s", got, want)
	}
}

func TestPlainText_NoRows(t *testing.T) {
	t.Parallel()
	if got := PlainText([]table.Column{{Title: "A"}, {Title: "B"}}, nil); got != "A  B\n" {
		t.Errorf("header only = %q", got)
	}
}
