// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

// Package tail wraps a [bubbles/v2/viewport] with a follow toggle, a
// live filter and a search, suitable for log / event / message streams
// and for documents shown whole.
//
// Apps append lines via [Model.AppendLine] as they arrive; the viewport
// auto-scrolls when follow mode is on, the filter ([RowFilter]) hides
// non-matching lines — except markers ([Model.AppendMarker]), separators
// that every filter lets through — and standard scroll keys (ctrl-f/b, ctrl-d/u,
// g/G, j/k, f to toggle follow) are dispatched via
// [Model.HandleScrollKey].
//
// Search ([Model.SetSearch]) is k9s's describe-view search: it hides
// nothing, highlights every match in the lines the filter shows, styling
// included, and [Model.NextMatch] / [Model.PrevMatch] move between matching
// lines, with [Model.MatchIndex] and [Model.Matches] for a "3/17" counter.
package tail
