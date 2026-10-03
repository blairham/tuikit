// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package chrome

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"github.com/blairham/tuikit/theme"
)

const (
	lightSkyBlueSGR = "38;2;135;206;250" // #87CEFA, Theme.BorderFocus
	dodgerBlueSGR   = "38;2;30;144;255"  // #1E90FF, Theme.Border
	cyanSGR         = "38;2;0;255;255"   // #00FFFF, Theme.HelpBorder
)

// sgrBefore returns the last SGR escape that precedes the first
// occurrence of glyph in line — the style the glyph is drawn in.
func sgrBefore(t *testing.T, line, glyph string) string {
	t.Helper()
	i := strings.Index(line, glyph)
	if i < 0 {
		t.Fatalf("glyph %q not found in %q", glyph, line)
	}
	j := strings.LastIndex(line[:i], "\x1b[")
	if j < 0 {
		t.Fatalf("no SGR before %q in %q", glyph, line)
	}
	return line[j:i]
}

// assertBoxColor checks that the top-left corner (the repainted title
// row) and the left side border of the box starting at row top are
// drawn in want — i.e. no color seam between title row and sides.
func assertBoxColor(t *testing.T, lines []string, top int, want string) {
	t.Helper()
	corner := sgrBefore(t, lines[top], "╭")
	side := sgrBefore(t, lines[top+1], "│")
	if !strings.Contains(corner, want) {
		t.Errorf("title-row corner drawn in %q, want %s", corner, want)
	}
	if !strings.Contains(side, want) {
		t.Errorf("side border drawn in %q, want %s", side, want)
	}
}

func TestBorderedContent_DefaultFocusColor(t *testing.T) {
	t.Parallel()
	c := New(Config{Theme: theme.Default()})
	box := c.BorderedContent("body", 40, 3)
	withTitle := InjectBorderTitle(box, c.Theme.Title.Render("streams"), c.Theme)
	lines := strings.Split(withTitle, "\n")
	assertBoxColor(t, lines, 0, lightSkyBlueSGR)
	bottom := sgrBefore(t, lines[len(lines)-1], "╰")
	if !strings.Contains(bottom, lightSkyBlueSGR) {
		t.Errorf("bottom border drawn in %q, want %s", bottom, lightSkyBlueSGR)
	}
	if strings.Contains(withTitle, dodgerBlueSGR) {
		t.Errorf("focused content box still carries unfocused %s: %q", dodgerBlueSGR, withTitle)
	}
}

func TestInjectBorderTitle_NilBorderFocusFallsBack(t *testing.T) {
	t.Parallel()
	th := theme.Theme{Border: lipgloss.Color("#1E90FF"), Title: lipgloss.NewStyle()}
	th.Rebuild()
	box := th.TableBorder.Width(40).Height(3).Render("body")
	lines := strings.Split(InjectBorderTitle(box, "streams", th), "\n")
	assertBoxColor(t, lines, 0, dodgerBlueSGR)
}

func TestRenderModal_TitleRowMatchesBorder(t *testing.T) {
	t.Parallel()
	c := New(Config{Theme: theme.Default()})
	m := NewModal(theme.Default())
	m.Open("Proceed?", ModalOpts{Title: "Confirm"})
	lines := strings.Split(c.renderModalOverlay(m, 120, 30), "\n")
	top := -1
	for i, l := range lines {
		if strings.Contains(l, "Confirm") {
			top = i
			break
		}
	}
	if top < 0 {
		t.Fatal("modal title row not found")
	}
	assertBoxColor(t, lines, top, dodgerBlueSGR)
}

func TestHelpOverlay_TitleRowMatchesBorder(t *testing.T) {
	t.Parallel()
	c := New(Config{Theme: theme.Default()})
	out := c.renderHelpOverlay(Frame{Width: 80, Height: 30})
	lines := strings.Split(out, "\n")
	assertBoxColor(t, lines, 0, cyanSGR)
}
