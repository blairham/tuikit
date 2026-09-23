// Package tail wraps a [bubbles/v2/viewport] with a follow toggle and a
// live regex filter, suitable for log / event / message streams.
//
// Apps append lines via [Model.AppendLine] as they arrive; the viewport
// auto-scrolls when follow mode is on, the filter ([RowFilter]) hides
// non-matching lines, and standard scroll keys (ctrl-f/b, ctrl-d/u,
// g/G, j/k, f to toggle follow) are dispatched via
// [Model.HandleScrollKey].
package tail
