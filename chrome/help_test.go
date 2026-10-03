// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package chrome

import (
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
