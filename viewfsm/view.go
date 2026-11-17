// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package viewfsm

import tea "charm.land/bubbletea/v2"

// View is the content one [ViewID] stands for: something the app sizes
// to the chrome's inner box and draws into it. It is the method set every
// tuikit app's views already share (k4a's, cwlogs's), and that
// [tail.Model] and the chart widgets satisfy as they are.
//
// The [Router] still does not hold views. An app keeps its concrete
// views, typically in a map keyed by its own ViewID type, and uses this
// interface to resize, filter or draw whichever one the router says is
// active without a type switch:
//
//	views := map[ViewID]viewfsm.View{ViewLogGroups: groups, ViewTail: tl}
//	for _, v := range views {
//		v.Resize(c.ContentInnerSize(w, h, filtering, commanding, false, false))
//	}
//	content := views[ViewID(r.Active())].View()
//
// [tail.Model]: https://pkg.go.dev/github.com/blairham/tuikit/tail#Model
type View interface {
	// Resize sets the inner width and height the view draws into, as
	// chrome.Chrome.ContentInnerSize reports them.
	Resize(width, height int)
	// View renders the view's content, without the border or title the
	// chrome draws around it.
	View() string
}

// Filterable is a view the filter bar narrows. expr is the bar's text as
// typed; the view parses it (tuikit's table.ParseFilter and tail's filter
// read the same regex-and-! syntax), and "" clears the filter.
type Filterable interface {
	SetFilter(expr string)
}

// List is a [View] of rows: a resource list whose border title carries a
// count, which shows a placeholder while its first load is in flight, and
// whose table takes the navigation keys.
type List interface {
	View
	Filterable
	// Count is the number of rows shown, after the filter.
	Count() int
	// Loading reports that the rows have not arrived yet, so an empty list
	// means "wait", not "none".
	Loading() bool
	// UpdateTable forwards a message to the list's table: a key the app
	// did not consume, already passed through [TranslateNavKey].
	UpdateTable(msg tea.Msg) tea.Cmd
}
