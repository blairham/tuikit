// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package chrome

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/blairham/tuikit/theme"
)

// TestHelpSectionHeaderIsPlainGreen pins k9s's section heading: the
// theme's HelpSection green with no bold and no underline, and a
// TitleColor override still winning.
func TestHelpSectionHeaderIsPlainGreen(t *testing.T) {
	th := theme.Default()
	c := New(Config{Theme: th})
	sec := HelpSection{Title: "RESOURCE", Entries: []HelpEntry{{Key: "<a>", Desc: "Attach"}}}

	got := c.renderHelpSection(sec, 30)
	if want := th.On(th.HelpSection).Render("RESOURCE"); !strings.Contains(got, want) {
		t.Errorf("header not rendered plain in HelpSection:\nwant substring %q\ngot %q", want, got)
	}
	for _, styled := range []string{
		th.On(th.HelpSection).Bold(true).Render("RESOURCE"),
		th.On(th.HelpSection).Underline(true).Render("RESOURCE"),
	} {
		if strings.Contains(got, styled) {
			t.Errorf("header carries bold/underline: %q", got)
		}
	}

	sec.TitleColor = lipgloss.Color("#FF00FF")
	got = c.renderHelpSection(sec, 30)
	if want := th.On(sec.TitleColor).Render("RESOURCE"); !strings.Contains(got, want) {
		t.Errorf("TitleColor override ignored: %q", got)
	}
}

// TestNavigationHelpListsVimKeysOnly pins that the column advertises the
// vim spellings and never the arrows they translate to: the arrows work
// regardless, and listing both doubles the column.
func TestNavigationHelpListsVimKeysOnly(t *testing.T) {
	nav := NavigationHelp()
	keys := map[string]string{}
	for _, e := range nav.Entries {
		keys[e.Key] = e.Desc
		for _, arrow := range []string{"↑", "↓", "←", "→", helpKey("up"), helpKey("down"), helpKey("left"), helpKey("right"), "pgup", "pgdown"} {
			if strings.Contains(strings.ToLower(e.Key), arrow) {
				t.Errorf("navigation help lists an arrow/page key: %q", e.Key)
			}
		}
	}
	for k, desc := range map[string]string{
		helpKey("j"): "Down", helpKey("k"): "Up", helpKey("h"): "Left", helpKey("l"): "Right",
		helpKey("g"): "Goto Top", helpKey("shift-g"): "Goto Bottom", helpKey("ctrl-f"): "Page Down", helpKey("ctrl-b"): "Page Up",
	} {
		if keys[k] != desc {
			t.Errorf("navigation help %s = %q, want %q", k, keys[k], desc)
		}
	}
}

// TestHelpMatchesTheKeyConstants keeps the advertised keys and the
// constants apps bind in step: a constant renamed without its help entry
// would advertise a key nothing handles.
func TestHelpMatchesTheKeyConstants(t *testing.T) {
	all := map[string]bool{}
	for _, s := range []HelpSection{GeneralHelp(), NavigationHelp()} {
		if s.Title == "" || len(s.Entries) == 0 {
			t.Errorf("empty help section %+v", s)
		}
		for _, e := range s.Entries {
			all[e.Key] = true
		}
	}
	for constant, shown := range map[string]string{
		KeyBack: helpKey("q"), KeyReload: helpKey("ctrl-r"), KeyToggleCrumbs: helpKey("ctrl-g"), KeyToggleHeader: helpKey("ctrl-e"),
		KeyHistoryBack: helpKey("["), KeyHistoryForward: helpKey("]"), KeyLastView: helpKey("-"),
		KeyFieldNext: helpKey("tab"), KeyFieldPrev: helpKey("backtab"),
	} {
		if !all[shown] {
			t.Errorf("constant %q is not advertised as %s", constant, shown)
		}
		if want := strings.NewReplacer("ctrl+", "ctrl-", "shift+tab", "backtab").Replace(constant); "<"+want+">" != shown {
			t.Errorf("constant %q and help key %s disagree", constant, shown)
		}
	}
}

// TestHelpSectionNeverWraps pins one row per entry: the key column is the
// section's widest key plus a space, and a description too long for what
// is left ends in "…" instead of wrapping into the next entry's row or
// running into the next column.
func TestHelpSectionNeverWraps(t *testing.T) {
	c := New(Config{Theme: theme.Default()})
	sec := HelpSection{Title: "GENERAL", Entries: []HelpEntry{
		{Key: "<a>", Desc: "All"},
		{Key: "<backtab>", Desc: "Field Previous"},
		{Key: "<b>", Desc: "Browse a published port in the browser"},
	}}
	const colWidth = 25
	lines := strings.Split(strings.TrimRight(ansi.Strip(c.renderHelpSection(sec, colWidth)), " \n"), "\n")
	if len(lines) != 1+len(sec.Entries) {
		t.Fatalf("%d lines for a title and %d entries — an entry wrapped:\n%s",
			len(lines), len(sec.Entries), strings.Join(lines, "\n"))
	}
	if !strings.HasPrefix(lines[1], "<a>       All") { // "<backtab>" is 9 wide: keys pad to 10
		t.Errorf("key column not sized to the widest key: %q", lines[1])
	}
	if !strings.Contains(lines[2], "Field Previous") {
		t.Errorf("a description that fits was cut: %q", lines[2])
	}
	long := strings.TrimRight(lines[3], " ")
	if !strings.HasSuffix(long, "…") || !strings.HasPrefix(long, "<b>       Browse") {
		t.Errorf("long description not ended with …: %q", long)
	}
	for _, l := range lines {
		if w := lipgloss.Width(strings.TrimRight(l, " ")); w > colWidth-1 {
			t.Errorf("%q is %d wide; the last cell must stay clear before the next column", l, w)
		}
	}
}

// helpSection is a test section of n entries, keys and descriptions named
// after the section so a cut one is easy to find.
func helpSection(title string, n int) HelpSection {
	s := HelpSection{Title: title}
	for i := range n {
		s.Entries = append(
			s.Entries,
			HelpEntry{Key: fmt.Sprintf("<%s%d>", strings.ToLower(title[:1]), i), Desc: fmt.Sprintf("%s entry %d", title, i)},
		)
	}
	return s
}

// TestHelpOverlayFitsWithExtraSections: an app's PLUGINS and HOTKEYS
// sections beyond the four that fit 120 columns stack under the shortest
// columns instead of widening the row past the screen, every entry stays
// whole, and the overlay is no taller than the content area.
func TestHelpOverlayFitsWithExtraSections(t *testing.T) {
	t.Parallel()
	c := New(Config{Theme: theme.Default()})
	sections := []HelpSection{
		helpSection("RESOURCE", 18), helpSection("CONTAINER", 20), helpSection("GENERAL", 16),
		helpSection("NAVIGATION", 11), helpSection("PLUGINS", 3), helpSection("HOTKEYS", 2),
	}
	f := Frame{Width: 120, Height: 40, Help: HelpPanel{Sections: sections}}
	_, innerH := c.ContentInnerSize(f.Width, f.Height, false, false, false, false)
	out := ansi.Strip(c.renderHelpOverlay(f))
	lines := strings.Split(out, "\n")
	if len(lines) != innerH+2 {
		t.Errorf("overlay is %d lines, want the content area's %d", len(lines), innerH+2)
	}
	for _, l := range lines {
		if w := ansi.StringWidth(l); w > f.Width {
			t.Fatalf("a line is %d cells wide, past the %d-cell screen: %q", w, f.Width, l)
		}
	}
	for _, s := range sections {
		if !strings.Contains(out, s.Title) {
			t.Errorf("section %s is missing", s.Title)
		}
		for _, e := range s.Entries {
			found := false
			for _, l := range lines {
				if i := strings.Index(l, e.Key); i >= 0 && strings.Contains(l[i:], e.Desc) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("%s %q is cut or missing", e.Key, e.Desc)
			}
		}
	}
	// PLUGINS goes under the shortest column, NAVIGATION (11 entries).
	for _, l := range lines {
		if strings.Contains(l, "PLUGINS") && !strings.Contains(l, "<r") && !strings.Contains(l, "<c") {
			t.Errorf("PLUGINS is not stacked beside the taller columns: %q", l)
		}
	}
}

// TestHelpOverlayClampsToHeight: sections taller than the content area are
// cut to it rather than pushing the frame past the terminal.
func TestHelpOverlayClampsToHeight(t *testing.T) {
	t.Parallel()
	c := New(Config{Theme: theme.Default()})
	f := Frame{Width: 120, Height: 30, Help: HelpPanel{Sections: []HelpSection{helpSection("TALL", 60)}}}
	_, innerH := c.ContentInnerSize(f.Width, f.Height, false, false, false, false)
	if got := len(strings.Split(c.renderHelpOverlay(f), "\n")); got != innerH+2 {
		t.Errorf("overlay is %d lines, want %d", got, innerH+2)
	}
}

// TestHelpOverlayCountsTheBlankRow: a stacked section adds its blank row to
// its column's height, so the next one goes under the column that is
// really shortest. Two columns of 10 and 12 rows: X (2 rows) stacks under
// the first, making it 13 with the blank, so Y goes under the second.
func TestHelpOverlayCountsTheBlankRow(t *testing.T) {
	t.Parallel()
	c := New(Config{Theme: theme.Default()})
	f := Frame{Width: 60, Height: 50, Help: HelpPanel{Sections: []HelpSection{
		helpSection("ALPHA", 9), helpSection("BRAVO", 11), helpSection("XRAY", 1), helpSection("YANKEE", 1),
	}}}
	out := ansi.Strip(c.renderHelpOverlay(f))
	col := func(title string) int {
		for _, l := range strings.Split(out, "\n") {
			if i := strings.Index(l, title); i >= 0 {
				return ansi.StringWidth(l[:i])
			}
		}
		t.Fatalf("%s not drawn", title)
		return -1
	}
	if col("XRAY") != col("ALPHA") || col("YANKEE") != col("BRAVO") {
		t.Errorf("XRAY under col %d (ALPHA %d), YANKEE under col %d (BRAVO %d)",
			col("XRAY"), col("ALPHA"), col("YANKEE"), col("BRAVO"))
	}
}
