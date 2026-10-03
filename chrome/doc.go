// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

// Package chrome renders the visual frame around a tuikit app: top
// section (info-panel + shortcut grid + ASCII logo), bordered content
// with an injected title, breadcrumb footer, status bar, filter and
// command bars, and the help overlay.
//
// The filter and command bars remember what the user submits: up and
// down recall earlier entries, k9s-style, and History / SetHistory let
// an app persist and seed them. Up/down belong to history in both bars,
// so the command bar cycles its suggestions on ctrl+n / ctrl+p; tab and
// → accept one.
//
// Apps assemble a [Frame] each tick and call [Chrome.Render]. The chrome
// owns the layout, the [Theme] applies the colors, and the app owns the
// content. No control inversion — apps still drive their own
// [tea.Program] update loop.
package chrome
