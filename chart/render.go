// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package chart

import (
	"image/color"
	"strings"
	"unicode"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/blairham/tuikit/theme"
)

const ellipsis = "…"

// canvas paints cells that carry no color of their own: the theme's
// background when it paints one, nothing otherwise.
type canvas struct {
	bg    color.Color
	paint bool
}

func newCanvas(t theme.Theme) canvas { return canvas{bg: t.Bg, paint: t.PaintBackground} }

// style is a style with fg and bg, either of which may be nil; a nil bg is
// the canvas.
func (c canvas) style(fg, bg color.Color) lipgloss.Style {
	s := lipgloss.NewStyle()
	if fg != nil {
		s = s.Foreground(fg)
	}
	switch {
	case bg != nil:
		s = s.Background(bg)
	case c.paint:
		s = s.Background(c.bg)
	}
	return s
}

// spaces is n blank cells on the canvas.
func (c canvas) spaces(n int) string {
	if n <= 0 {
		return ""
	}
	if !c.paint {
		return strings.Repeat(" ", n)
	}
	return c.style(nil, nil).Render(strings.Repeat(" ", n))
}

// cell is one drawn character cell.
type cell struct {
	fg, bg color.Color
	r      rune
}

// line renders cells, one style per run of cells that share their colors.
func (c canvas) line(cells []cell) string {
	var b strings.Builder
	for i := 0; i < len(cells); {
		j := i + 1
		for j < len(cells) && cells[j].fg == cells[i].fg && cells[j].bg == cells[i].bg {
			j++
		}
		run := make([]rune, 0, j-i)
		for _, x := range cells[i:j] {
			run = append(run, x.r)
		}
		if cells[i].fg == nil && cells[i].bg == nil && !c.paint {
			b.WriteString(string(run))
		} else {
			b.WriteString(c.style(cells[i].fg, cells[i].bg).Render(string(run)))
		}
		i = j
	}
	return b.String()
}

// plain is s as one line of plain text: escape sequences removed, invalid
// UTF-8 replaced, and every other control character drawn as a space.
func plain(s string) string {
	s = strings.ToValidUTF8(ansi.Strip(s), "�")
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, s)
}

// width is how many cells s is drawn in, escape sequences not counted.
func width(s string) int { return ansi.StringWidth(s) }

// cut shortens plain text to at most n cells, ending it in "…" when it
// had to be cut.
func cut(s string, n int) string {
	if n <= 0 {
		return ""
	}
	return ansi.Truncate(s, n, ellipsis)
}

// header is a one-line header w cells wide: title on the left in titleStyle,
// label on the right in labelStyle, the canvas between. The title gives
// way first when both do not fit.
func (c canvas) header(title, label string, titleStyle, labelStyle lipgloss.Style, w int) string {
	label = cut(label, w)
	room := w - width(label)
	if label != "" {
		room-- // keep a space between title and label
	}
	title = cut(title, room)
	var b strings.Builder
	if title != "" {
		b.WriteString(titleStyle.Render(title))
	}
	b.WriteString(c.spaces(w - width(title) - width(label)))
	if label != "" {
		b.WriteString(labelStyle.Render(label))
	}
	return b.String()
}

// fit makes s exactly w cells wide by h lines: longer lines are cut,
// shorter ones padded with the canvas, missing lines added blank and extra
// ones dropped.
func (c canvas) fit(s string, w, h int) []string {
	in := strings.Split(s, "\n")
	out := make([]string, h)
	for i := range out {
		line := ""
		if i < len(in) {
			line = in[i]
		}
		if cw := width(line); cw > w {
			line = ansi.Truncate(line, w, "")
		}
		out[i] = line + c.spaces(w-width(line))
	}
	return out
}
