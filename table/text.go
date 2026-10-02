package table

import (
	"strings"

	"charm.land/bubbles/v2/table"
	"github.com/charmbracelet/x/ansi"
)

// PlainText renders a header line and every row as plain text: ANSI
// stripped, each column padded to its widest cell (title included) and
// separated by two spaces, trailing spaces trimmed. Unlike the table's own
// View, it includes every row rather than the ones scrolled into view and
// never truncates a cell, which is what a saved view wants (see
// chrome.SaveDump). Cells beyond the last column are ignored.
func PlainText(cols []table.Column, rows []table.Row) string {
	widths := make([]int, len(cols))
	for i, c := range cols {
		widths[i] = ansi.StringWidth(ansi.Strip(c.Title))
	}
	for _, r := range rows {
		for i := range widths {
			widths[i] = max(widths[i], ansi.StringWidth(cell(r, i)))
		}
	}

	var b strings.Builder
	line := func(field func(i int) string) {
		var l strings.Builder
		for i, w := range widths {
			s := field(i)
			l.WriteString(s)
			if i < len(widths)-1 {
				l.WriteString(strings.Repeat(" ", w-ansi.StringWidth(s)+2))
			}
		}
		b.WriteString(strings.TrimRight(l.String(), " "))
		b.WriteByte('\n')
	}
	line(func(i int) string { return ansi.Strip(cols[i].Title) })
	for _, r := range rows {
		line(func(i int) string { return cell(r, i) })
	}
	return b.String()
}
