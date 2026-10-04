// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

// Package chart draws small live charts for a dashboard in the style of
// k9s's pulses: [Sparkline]s of samples over time and [Gauge]s of a value
// against a max, laid out in a terminal grid by [Grid] or by the app's own
// lipgloss joins.
//
// Samples go into a [Series], a ring buffer that keeps the most recent
// ones; a [Sparkline] draws whatever its series hold each time it is
// rendered, so an app pushes a sample on each poll and renders as usual.
//
// Every widget is sized with Resize and draws, from View, exactly as many
// lines as its height, each exactly as many cells wide as its width, with
// escape sequences not counted. A grid of them lines up without
// measuring.
//
// # Samples that are not numbers
//
// A [Series] keeps whatever it is given; only drawing interprets it. A
// NaN, a negative value and -Inf draw as zero: an empty column. +Inf
// draws full height (or a full bar). Infinite and NaN values never set
// the scale: an auto-scaled [Sparkline] scales to the largest finite
// sample on screen. A value above a fixed max is drawn at the max. A
// positive value too small to reach one level still draws the lowest
// one, so a nonzero sample is never invisible.
//
// Titles and labels are drawn as plain text: escape sequences are
// removed and line breaks and tabs drawn as spaces, so a header is
// always one line of known width.
package chart
