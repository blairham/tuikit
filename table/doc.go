// Package table provides themed style and keymap helpers for
// [bubbles/v2/table], plus the [FixSelectedRow] ANSI post-processor
// that solves bubbles-table's selection-paint bug, a [RowFilter]
// type for regex-based row filtering with negation and literal fallback,
// a [Sorter] that owns a table's sort column and direction, [Marks]
// for rows marked for a bulk action, and [PlainText] for saving a table.
package table
