// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

// Package table provides themed style and keymap helpers for
// [bubbles/v2/table], plus the [FixSelectedRow] ANSI post-processor
// that solves bubbles-table's selection-paint bug, a [RowFilter]
// for k9s's filter-bar modes — a regex with "!" negation and a literal
// fallback, "-f" for a fuzzy match and "-l" for a label selector —
// a [Sorter] that owns a table's sort column and direction, [Marks]
// for rows marked for a bulk action, and [PlainText] for saving a table.
package table
