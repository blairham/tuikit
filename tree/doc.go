// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

// Package tree draws a navigable, collapsible tree in the style of k9s's
// xray view: a resource and everything under it, one node per row, with
// box-drawing guides and a ▸ / ▾ marker on the nodes that have children.
//
// Apps hand the content over as [Node] values via [Model.SetRoots], as
// often as they like: expansion state and the cursor are kept by
// [Node.ID], so a tree refreshed on a poll keeps the user's place. Keys go
// through [Model.HandleKey], which reports whether the tree used them, so
// enter and the app's own bindings fall through. [Model.SetFilter] takes a
// [table.ParseFilter] expression and keeps a node when it or any of its
// descendants matches, with the path down to every match opened.
package tree
