// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package viewfsm_test

import (
	"github.com/blairham/tuikit/chart"
	"github.com/blairham/tuikit/tail"
	"github.com/blairham/tuikit/tree"
	"github.com/blairham/tuikit/viewfsm"
)

// tuikit's own content widgets are Views as they stand, so an app can keep
// them in the same map as its list views. A method renamed on any of them
// breaks this file rather than the apps that rely on it.
var (
	_ viewfsm.View       = (*tail.Model)(nil)
	_ viewfsm.Filterable = (*tail.Model)(nil)
	_ viewfsm.View       = (*tree.Model)(nil)
	_ viewfsm.Filterable = (*tree.Model)(nil)
	_ viewfsm.View       = (*chart.Sparkline)(nil)
	_ viewfsm.View       = (*chart.Gauge)(nil)
	_ viewfsm.View       = (*chart.Grid)(nil)

	// chart.Widget predates View and has the same method set.
	_ chart.Widget = viewfsm.View(nil)
	_ viewfsm.View = chart.Widget(nil)
)
