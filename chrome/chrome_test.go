package chrome

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"

	"github.com/blairham/tuikit/theme"
)

func TestNew_FillsDefaults(t *testing.T) {
	t.Parallel()
	c := New(Config{})
	if c.InfoLabelWidth != 56 {
		t.Errorf("InfoLabelWidth = %d; want 56", c.InfoLabelWidth)
	}
	if c.ShortcutColumnWidth != 42 {
		t.Errorf("ShortcutColumnWidth = %d; want 42", c.ShortcutColumnWidth)
	}
	if c.MinLogoWidth != 134 {
		t.Errorf("MinLogoWidth = %d; want 134", c.MinLogoWidth)
	}
	if c.Theme.Bg == nil {
		t.Error("Theme should default to theme.Default()")
	}
}

func TestNew_HonorsCustomConfig(t *testing.T) {
	t.Parallel()
	c := New(Config{
		Theme:               theme.NoPaintBackground(),
		Logo:                []string{"x", "y"},
		InfoLabelWidth:      40,
		ShortcutColumnWidth: 30,
		MinLogoWidth:        100,
	})
	if c.InfoLabelWidth != 40 || c.ShortcutColumnWidth != 30 || c.MinLogoWidth != 100 {
		t.Error("Config overrides not honored")
	}
	if c.Theme.PaintBackground {
		t.Error("Custom theme's PaintBackground flag lost")
	}
	if len(c.Logo) != 2 {
		t.Error("Logo not stored")
	}
}

func TestContentInnerSize(t *testing.T) {
	t.Parallel()
	// Default chrome: no logo, default rows. Top = max(4 info, 5 shortcut) = 5.
	// Reserved = 5 + 2 + 2 = 9 (footer reserves a row of bottom padding so
	// the breadcrumb pill mirrors the bordered content's left/right margin).
	c := New(Config{})
	cases := []struct {
		name                string
		w, h                int
		flt, cmd, cnf, stat bool
		wantW, wantH        int
	}{
		{"basic", 150, 40, false, false, false, false, 146, 31},
		{"with filter", 150, 40, true, false, false, false, 146, 28},
		{"with confirm", 150, 40, false, false, true, false, 146, 28},
		{"all bars", 150, 40, true, true, true, true, 146, 21},
		{"floor at 1", 150, 5, false, false, false, false, 146, 1},
		{"width floor", 8, 40, false, false, false, false, 10, 31},
	}
	for _, tc := range cases {
		gotW, gotH := c.ContentInnerSize(tc.w, tc.h, tc.flt, tc.cmd, tc.cnf, tc.stat)
		if gotW != tc.wantW || gotH != tc.wantH {
			t.Errorf("%s: ContentInnerSize(%d,%d,%v,%v,%v,%v) = (%d,%d); want (%d,%d)",
				tc.name, tc.w, tc.h, tc.flt, tc.cmd, tc.cnf, tc.stat, gotW, gotH, tc.wantW, tc.wantH)
		}
	}
}

func TestTopSectionRows(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		cfg  Config
		want int
	}{
		{"defaults: max(4 info, 5 shortcut) = 5", Config{}, 5},
		{"5-row logo matches default shortcut count", Config{Logo: lines(5)}, 5},
		{"9-row logo dominates the right column", Config{Logo: lines(9)}, 9},
		{"tall info panel dominates", Config{InfoPanelRows: 8}, 8},
		{"explicit shortcut row count beats logo + info", Config{ShortcutRows: 7, Logo: lines(3)}, 7},
	}
	for _, tc := range cases {
		got := New(tc.cfg).TopSectionRows()
		if got != tc.want {
			t.Errorf("%s: TopSectionRows = %d; want %d", tc.name, got, tc.want)
		}
	}
}

func lines(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = "x"
	}
	return out
}

func TestShortcut_RendersWithStyles(t *testing.T) {
	t.Parallel()
	c := New(Config{})
	out := c.Shortcut("<enter>", "Tail")
	if !strings.Contains(out, "<enter>") || !strings.Contains(out, "Tail") {
		t.Errorf("Shortcut output missing expected content: %q", out)
	}
	if !strings.Contains(out, "\x1b[") {
		t.Errorf("Shortcut output has no ANSI styling: %q", out)
	}
}

// TestShortcut_ViewVsActionKeyColoring verifies that view keys ("<N>")
// pick up ShortcutView and action keys ("<enter>", "<ctrl-d>", etc.)
// pick up ShortcutKey — the k9s convention.
func TestShortcut_ViewVsActionKeyColoring(t *testing.T) {
	t.Parallel()
	th := theme.Default()
	c := New(Config{Theme: th})

	viewKey := c.Shortcut("<1>", "Topics")
	actionKey := c.Shortcut("<enter>", "View")

	if !strings.Contains(viewKey, th.ShortcutView.Render("")[:5]) &&
		viewKey == actionKey {
		t.Errorf("view and action keys rendered identically — expected different ANSI: view=%q action=%q", viewKey, actionKey)
	}

	cases := []struct {
		key    string
		isView bool
	}{
		{"<0>", true},
		{"<1>", true},
		{"<10>", true},
		{"<enter>", false},
		{"<ctrl-d>", false},
		{"<:q>", false},
		{"<a>", false},
		{"", false},
		{"<>", false},
	}
	for _, tc := range cases {
		if got := isViewKey(tc.key); got != tc.isView {
			t.Errorf("isViewKey(%q) = %v; want %v", tc.key, got, tc.isView)
		}
	}
}

// TestShortcutGrid verifies that views fill column 1 and actions
// column 2, that the row count matches the longer list, and that
// each rendered row contains the expected key + description text.
func TestShortcutGrid(t *testing.T) {
	t.Parallel()
	c := New(Config{Theme: theme.Default()})

	views := []Shortcut{
		{Key: "<1>", Desc: "Topics"},
		{Key: "<2>", Desc: "Groups"},
		{Key: "<3>", Desc: "Cluster"},
	}
	actions := []Shortcut{
		{Key: "<enter>", Desc: "Messages"},
		{Key: "<o>", Desc: "Overview"},
		{Key: "<p>", Desc: "Produce"},
		{Key: "<ctrl-d>", Desc: "Delete"},
		{Key: "<r>", Desc: "Refresh"},
	}

	rows := c.ShortcutGrid(views, actions)
	if len(rows) != len(actions) {
		t.Fatalf("ShortcutGrid rows = %d; want %d (max of view/action counts)", len(rows), len(actions))
	}

	// Row 0: view 1 + action 0
	if !strings.Contains(rows[0], "<1>") || !strings.Contains(rows[0], "Topics") {
		t.Errorf("row 0 missing view 0: %q", rows[0])
	}
	if !strings.Contains(rows[0], "<enter>") || !strings.Contains(rows[0], "Messages") {
		t.Errorf("row 0 missing action 0: %q", rows[0])
	}

	// Row 3: views exhausted (empty cell), action 3
	if strings.Contains(rows[3], "<3>") || strings.Contains(rows[3], "<1>") {
		t.Errorf("row 3 should have no view cell: %q", rows[3])
	}
	if !strings.Contains(rows[3], "<ctrl-d>") {
		t.Errorf("row 3 missing action 3: %q", rows[3])
	}

	// Row 4: views still exhausted, action 4 present
	if !strings.Contains(rows[4], "Refresh") {
		t.Errorf("row 4 missing action 4: %q", rows[4])
	}
}

func TestVersionLine(t *testing.T) {
	t.Parallel()
	c := New(Config{})

	t.Run("empty current returns empty", func(t *testing.T) {
		t.Parallel()
		if got := c.VersionLine("Rev:", "", "v1.0.0"); got != "" {
			t.Errorf("VersionLine with empty current = %q; want empty", got)
		}
	})

	t.Run("no latest renders label + current only", func(t *testing.T) {
		t.Parallel()
		out := c.VersionLine("Rev:", "v0.4.2", "")
		if !strings.Contains(out, "Rev:") || !strings.Contains(out, "v0.4.2") {
			t.Errorf("VersionLine output missing label/current: %q", out)
		}
		if strings.Contains(out, "⚡") {
			t.Errorf("VersionLine should not include lightning bolt when latest is empty: %q", out)
		}
	})

	t.Run("latest equal to current suppresses bolt", func(t *testing.T) {
		t.Parallel()
		out := c.VersionLine("Rev:", "v0.4.2", "v0.4.2")
		if strings.Contains(out, "⚡") {
			t.Errorf("VersionLine should not show bolt when latest == current: %q", out)
		}
	})

	t.Run("newer latest appends lightning bolt + version", func(t *testing.T) {
		t.Parallel()
		out := c.VersionLine("Rev:", "v0.4.2", "v0.5.0")
		if !strings.Contains(out, "⚡") {
			t.Errorf("VersionLine missing lightning bolt for newer version: %q", out)
		}
		if !strings.Contains(out, "v0.5.0") {
			t.Errorf("VersionLine missing latest version: %q", out)
		}
	})
}

// TestRenderTopSection_LogolessShortcutsLeftAligned verifies that in
// logoless mode shortcuts sit immediately to the right of the info
// panel (k9s convention) — every shortcut row starts in the same
// column, and that column is at the info-panel's right edge (no large
// gap pushing shortcuts toward the right). Rows do NOT hug the chrome's
// right edge; the unused right space is left blank for the chrome bg.
func TestRenderTopSection_LogolessShortcutsLeftAligned(t *testing.T) {
	t.Parallel()
	th := theme.Default()
	th.PaintBackground = false
	c := New(Config{Theme: th}) // no Logo → logoless path

	const width = 200
	frame := Frame{
		Width:  width,
		Height: 20,
		InfoLines: []string{
			"Context: production-us",
			"Auth:    iam",
		},
		Shortcuts: []string{
			c.ShortcutPair("<enter>", "Switch", "<esc>", "Back"),
			c.ShortcutPair("<ctrl-f>", "PgDn", "<ctrl-b>", "PgUp"),
			c.ShortcutPair("<1>", "Topics", "<2>", "Groups"),
			c.ShortcutPair("<3>", "Cluster", "<4>", "ACLs"),
			c.ShortcutPair("<r>", "Refresh", "<:q>", "Quit"),
		},
	}

	out := c.renderTopSection(frame)
	rows := strings.Split(strings.TrimRight(out, "\n"), "\n")

	// Left inset: InfoLines start at col 1 (PaddingLeft(1)).
	if got := rows[0]; len(got) < 2 || got[0] != ' ' || got[1] == ' ' {
		t.Errorf("InfoLines row 0: want 1-col left inset before non-space; got %q", got[:min(20, len(got))])
	}

	// Shortcut rows occupy rows[len(InfoLines):]. Every row's first
	// visible col must match — left-aligned inside the right block.
	shortcutRows := rows[len(frame.InfoLines):]
	wantFirstCol := -1
	for i, row := range shortcutRows {
		firstVisCol := firstVisibleCol(row)
		if i == 0 {
			wantFirstCol = firstVisCol
			continue
		}
		if firstVisCol != wantFirstCol {
			t.Errorf(
				"shortcut row %d: firstVisibleCol = %d; want %d (left-aligned)",
				i+len(frame.InfoLines),
				firstVisCol,
				wantFirstCol,
			)
		}
	}

	// Shortcuts sit at the info-panel boundary, NOT at the right edge.
	// Expected start col = InfoLabelWidth (the left block's width).
	if wantFirstCol != c.InfoLabelWidth {
		t.Errorf(
			"shortcut start col = %d; want %d (immediately right of info panel)",
			wantFirstCol,
			c.InfoLabelWidth,
		)
	}

	// And no shortcut row hugs the chrome's right edge — the rightmost
	// visible col should be well short of width-2.
	rightmost := -1
	for _, row := range shortcutRows {
		if lv := lastVisibleCol(row); lv > rightmost {
			rightmost = lv
		}
	}
	if rightmost >= width-2 {
		t.Errorf(
			"widest shortcut row lastVisibleCol = %d; want < %d (left-aligned, not right-pinned)",
			rightmost,
			width-2,
		)
	}
}

// lastVisibleCol returns the 0-indexed visible column of the rightmost
// non-space byte in line, treating ANSI CSI escapes (\x1b[ ... <final>)
// as zero-width. The final byte of a CSI sequence is in 0x40..0x7E and
// is preceded by 0..N parameter bytes (0x30..0x3F).
func lastVisibleCol(line string) int {
	last := -1
	col := 0
	for j := 0; j < len(line); j++ {
		b := line[j]
		if b == 0x1b && j+1 < len(line) && line[j+1] == '[' {
			j += 2
			for j < len(line) {
				c := line[j]
				if c >= 0x40 && c <= 0x7E {
					break
				}
				j++
			}
			continue
		}
		if b != ' ' {
			last = col
		}
		col++
	}
	return last
}

// TestRenderTopSection_TrailingInsetPaintsBackground verifies that the
// 1-col right inset on each top-section row carries the chrome bg when
// PaintBackground=true. Without explicit bg on that cell, the inner
// styled content (LogoStyle / shortcut padder) emits a reset escape
// before the trailing space and the outer styledBlock's bg is not
// re-applied — the raw cell falls back to the terminal's default
// background, showing as a gray sliver on the right edge of every
// top-section row.
func TestRenderTopSection_TrailingInsetPaintsBackground(t *testing.T) {
	t.Parallel()
	c := New(Config{
		Theme: theme.Default(), // PaintBackground=true
		Logo: []string{
			"   _____      _| | ___   __ _ ___ ",
			"  / __\\ \\ /\\ / / |/ _ \\ / _` / __|",
			" | (__ \\ V  V /| | (_) | (_| \\__ \\",
			"  \\___| \\_/\\_/ |_|\\___/ \\__, |___/",
			"                        |___/     ",
		},
	})

	frame := Frame{
		Width:     200,
		Height:    20,
		InfoLines: []string{"Env: prd"},
		Shortcuts: []string{
			c.ShortcutPair("<enter>", "Streams", "</>", "Filter"),
			c.ShortcutPair("<j/k>", "Move", "<r>", "Refresh"),
			c.ShortcutPair("<1>", "LogGroups", "<2>", "Streams"),
			c.ShortcutPair("<3>", "Tail", "<?>", "Help"),
			c.ShortcutPair("<:>", "Command", "<q>", "Quit"),
		},
	}
	out := c.renderTopSection(frame)
	rows := strings.Split(strings.TrimRight(out, "\n"), "\n")

	// Each row's last visible cell (the 1-col right inset) must sit
	// inside a bg-painting span — i.e. the byte sequence immediately
	// before the line's final ANSI reset must include a 48; background
	// SGR. Searching for "48;" between the last reset-prelude and the
	// last visible character is fragile; instead, check that the line
	// does NOT end with a bare space + reset(s), which is the broken
	// shape we saw before the fix.
	brokenSuffix := "\x1b[m \x1b[m"
	for i, row := range rows[:len(frame.Shortcuts)] {
		if strings.HasSuffix(row, brokenSuffix) || strings.HasSuffix(row, brokenSuffix+"\x1b[m") {
			t.Errorf("top-section row %d ends with unpainted trailing inset; got %q", i, lastEsc(row, 24))
		}
	}
}

func lastEsc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}

// firstVisibleCol returns the 0-indexed visible column of the leftmost
// non-space byte in line, applying the same CSI-skipping convention.
func firstVisibleCol(line string) int {
	col := 0
	for j := 0; j < len(line); j++ {
		b := line[j]
		if b == 0x1b && j+1 < len(line) && line[j+1] == '[' {
			j += 2
			for j < len(line) {
				c := line[j]
				if c >= 0x40 && c <= 0x7E {
					break
				}
				j++
			}
			continue
		}
		if b != ' ' {
			return col
		}
		col++
	}
	return -1
}

func TestRender_ZeroDimensionsReturnsEmpty(t *testing.T) {
	t.Parallel()
	c := New(Config{})
	if got := c.Render(Frame{}); got != "" {
		t.Errorf("Render with zero dims = %q; want empty", got)
	}
}

func TestRenderBars_SpanFullFrameWidth(t *testing.T) {
	t.Parallel()
	// lipgloss/v2 Width includes borders in the rendered cell count.
	// Each bar must render exactly `width` cells wide so the right
	// edge lines up with the top section, status bar, and footer —
	// otherwise the bordered filter/command/confirm bars float 2
	// cells short of the frame on the right.
	c := New(Config{})
	in := textinput.New()
	const width = 120
	cases := []struct {
		name string
		out  string
	}{
		{"filter", c.renderFilterBar(&in, width)},
		{"command", c.renderCommandBar(&in, width)},
		{"confirm", c.renderConfirmBar("Delete topic?", width)},
	}
	for _, tc := range cases {
		first := strings.SplitN(tc.out, "\n", 2)[0]
		if w := lipgloss.Width(first); w != width {
			t.Errorf("%s bar width = %d; want %d", tc.name, w, width)
		}
	}
}

func TestRenderFooter_FillsTwoRowReservation(t *testing.T) {
	t.Parallel()
	// ContentInnerSize reserves 2 rows for the footer so the breadcrumb
	// pill sits one row above the bottom of the terminal — mirroring the
	// 1-col gap between the bordered content and the screen edges.
	c := New(Config{})
	out := c.renderFooter([]Crumb{{Label: "LogGroups"}}, 120)
	if h := lipgloss.Height(out); h != 2 {
		t.Errorf("renderFooter height = %d; want 2", h)
	}
}

func TestRenderStatusBar_FillsOneRowReservation(t *testing.T) {
	t.Parallel()
	// ContentInnerSize reserves 1 row for the status bar. The rendered
	// output must fill exactly 1 row — anything taller would push the
	// footer down off the terminal, anything shorter would let the
	// outer `screen.Height(f.Height)` wrapper pad blank rows at the
	// bottom (visible as wasted space below the footer pill).
	c := New(Config{})
	for _, msg := range []string{"x", "short error", "a slightly longer error message that fits on one line"} {
		out := c.renderStatusBar(msg, LevelError, 120)
		if h := lipgloss.Height(out); h != 1 {
			t.Errorf("renderStatusBar(%q) height = %d; want 1", msg, h)
		}
	}
}

func TestInjectBorderTitle(t *testing.T) {
	t.Parallel()
	c := New(Config{})
	box := c.Theme.TableBorder.Width(40).Height(3).Render("body")
	withTitle := InjectBorderTitle(box, c.Theme.Title.Render("streams"), c.Theme)
	if !strings.Contains(withTitle, "streams") {
		t.Errorf("title not injected: %q", withTitle)
	}
	// Should still have the bottom-row border characters.
	if !strings.Contains(withTitle, "╰") || !strings.Contains(withTitle, "╯") {
		t.Errorf("bottom border lost during injection: %q", withTitle)
	}
}
