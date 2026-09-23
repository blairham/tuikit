// Package viewfsm provides the [View] interface and [Router] that
// drive a tuikit app's view stack: drill-in / pop / jump-to-view-by-key,
// global key dispatch (/, :, ?, esc, q, 1..N, enter, r, j/k/g/G/ctrl-d/u),
// and breadcrumb emission for the chrome footer.
//
// The router supports both fixed shallow drill stacks and
// arbitrarily-deep stacks with detail / form views. Apps register
// views, set a few key bindings, and let the router translate keys
// to view actions.
package viewfsm
