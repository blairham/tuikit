package chrome

import (
	"fmt"
	"strings"
	"testing"

	"github.com/blairham/tuikit/theme"
)

// stripANSI removes CSI escape sequences so tests can reason about the
// visible cells of a rendered line.
func stripANSI(s string) string {
	var sb strings.Builder
	for j := 0; j < len(s); j++ {
		if s[j] == 0x1b && j+1 < len(s) && s[j+1] == '[' {
			j += 2
			for j < len(s) && (s[j] < 0x40 || s[j] > 0x7E) {
				j++
			}
			continue
		}
		sb.WriteByte(s[j])
	}
	return sb.String()
}

func twelveActions() []Shortcut {
	out := make([]Shortcut, 12)
	for i := range out {
		out[i] = Shortcut{Key: fmt.Sprintf("<%c>", 'a'+i), Desc: fmt.Sprintf("Action%02d", i)}
	}
	return out
}

func sixRowLogo() []string {
	out := make([]string, 6)
	for i := range out {
		out[i] = strings.Repeat("#", 20)
	}
	return out
}

// assertFrameFits checks the issue #4 repro invariants: the rendered
// frame is exactly height lines, and the columns the logo occupies on
// every top-section row hold nothing but logo glyphs or blanks.
func assertFrameFits(t *testing.T, c Chrome, out string, width, height int, wantShortcut string) {
	t.Helper()
	got := strings.Split(out, "\n")
	if len(got) != height {
		t.Fatalf("Render produced %d lines; want %d (frame overflowed the terminal)", len(got), height)
	}
	top := c.TopSectionRows()
	logoW := 20
	logoStart := width - 1 - logoW // trailing 1-col inset after the logo
	sawShortcut := false
	for i := range top {
		line := []rune(stripANSI(got[i]))
		if strings.Contains(string(line), wantShortcut) {
			sawShortcut = true
		}
		if len(line) < width {
			t.Fatalf("row %d is %d cells; want %d", i, len(line), width)
		}
		for col := logoStart; col < width-1; col++ {
			if r := line[col]; r != '#' && r != ' ' {
				t.Errorf("row %d col %d = %q: shortcut text under the logo: %q", i, col, r, string(line))
				break
			}
		}
	}
	if !sawShortcut {
		t.Errorf("no top-section row contains %q", wantShortcut)
	}
}

// TestRender_TooManyShortcutRowsStayInFrame is the issue #4 repro: a
// 6-line logo with 12 pre-rendered shortcut rows at 160x40 must still
// render exactly 40 lines, with no shortcut drawn under the logo.
func TestRender_TooManyShortcutRowsStayInFrame(t *testing.T) {
	t.Parallel()
	const width, height = 160, 40
	c := New(Config{Theme: theme.Default(), Logo: sixRowLogo()})

	rows := make([]string, 0, 12)
	for _, a := range twelveActions() {
		rows = append(rows, c.Shortcut(a.Key, a.Desc))
	}
	innerW, innerH := c.ContentInnerSize(width, height, false, false, false, false)
	frame := Frame{
		Width:      width,
		Height:     height,
		InfoLines:  []string{"Context: prod"},
		Shortcuts:  rows,
		Content:    c.BorderedContent("body", innerW+2, innerH),
		Breadcrumb: []Crumb{{Label: "view", Leaf: true}},
	}
	assertFrameFits(t, c, c.Render(frame), width, height, "Action00")
}

// TestShortcutGrid_WrapsIntoColumns checks that a grid never grows taller
// than the header's reservation: 12 actions under a 6-row top section
// wrap into a second action column (k9s behavior) instead of 12 rows.
func TestShortcutGrid_WrapsIntoColumns(t *testing.T) {
	t.Parallel()
	const width, height = 160, 40
	c := New(Config{Theme: theme.Default(), Logo: sixRowLogo()})

	views := []Shortcut{{Key: "<1>", Desc: "Pods"}, {Key: "<2>", Desc: "Nodes"}}
	grid := c.ShortcutGrid(views, twelveActions())
	if len(grid) != c.TopSectionRows() {
		t.Fatalf("ShortcutGrid rows = %d; want %d (wrapped at TopSectionRows)", len(grid), c.TopSectionRows())
	}
	// Row 0 carries view 0, action 0 and action 6 (first of the second
	// action column); every action must appear exactly once.
	row0 := stripANSI(grid[0])
	for _, want := range []string{"<1>", "Pods", "Action00", "Action06"} {
		if !strings.Contains(row0, want) {
			t.Errorf("row 0 missing %q: %q", want, row0)
		}
	}
	all := stripANSI(strings.Join(grid, "\n"))
	for _, a := range twelveActions() {
		if n := strings.Count(all, a.Desc); n != 1 {
			t.Errorf("%s appears %d times; want 1", a.Desc, n)
		}
	}

	innerW, innerH := c.ContentInnerSize(width, height, false, false, false, false)
	frame := Frame{
		Width:      width,
		Height:     height,
		InfoLines:  []string{"Context: prod"},
		Shortcuts:  grid,
		Content:    c.BorderedContent("body", innerW+2, innerH),
		Breadcrumb: []Crumb{{Label: "view", Leaf: true}},
	}
	assertFrameFits(t, c, c.Render(frame), width, height, "Action11")
}

// TestShortcutPair_DescriptionsNeverTouchNextKey is issue #5: a
// ten-character description filled its column exactly, rendering
// "Containers<a>".
func TestShortcutPair_DescriptionsNeverTouchNextKey(t *testing.T) {
	t.Parallel()
	c := New(Config{})
	pair := stripANSI(c.ShortcutPair("<1>", "Containers", "<a>", "Attach"))
	if strings.Contains(pair, "Containers<a>") || !strings.Contains(pair, "Containers ") {
		t.Errorf("ShortcutPair runs columns together: %q", pair)
	}
	long := stripANSI(c.Shortcut("<ctrl-shift-d>", "Delete"))
	if !strings.Contains(long, "<ctrl-shift-d> ") {
		t.Errorf("Shortcut runs an over-wide key into its description: %q", long)
	}

	grid := c.ShortcutGrid(
		[]Shortcut{{Key: "<1>", Desc: "Pods"}, {Key: "<2>", Desc: "Containers"}},
		[]Shortcut{{Key: "<a>", Desc: "Attach"}, {Key: "<l>", Desc: "Logs"}},
	)
	// Columns stay aligned: the action key starts at the same visible
	// column on every row, past the longest view description.
	first := strings.Index(stripANSI(grid[0]), "<a>")
	second := strings.Index(stripANSI(grid[1]), "<l>")
	if first != second {
		t.Errorf("action column misaligned: <a> at %d, <l> at %d\n%s", first, second, stripANSI(strings.Join(grid, "\n")))
	}
	if strings.Contains(stripANSI(grid[1]), "Containers<l>") {
		t.Errorf("ShortcutGrid runs columns together: %q", stripANSI(grid[1]))
	}
}

// TestRender_ShortcutRowsPastLogoAlignLeft covers the second half of
// issue #4: with a logo shorter than the shortcut rows, the rows past
// the logo must start in the same column as the rows beside it rather
// than right-aligning under the logo.
func TestRender_ShortcutRowsPastLogoAlignLeft(t *testing.T) {
	t.Parallel()
	const width, height = 160, 40
	c := New(Config{Theme: theme.Default(), Logo: sixRowLogo()[:3]})
	if c.TopSectionRows() != 6 {
		t.Fatalf("TopSectionRows = %d; want 6", c.TopSectionRows())
	}
	rows := []string{
		c.ShortcutPair("<1>", "Pods", "<a>", "Attach"),
		c.ShortcutPair("<2>", "Nodes", "<l>", "Logs"),
		c.ShortcutPair("<3>", "Jobs", "<d>", "Describe"),
		c.ShortcutPair("<4>", "Events", "<e>", "Edit"),
		c.ShortcutPair("<5>", "Secrets", "<s>", "Shell"),
	}
	innerW, innerH := c.ContentInnerSize(width, height, false, false, false, false)
	out := c.Render(Frame{
		Width:      width,
		Height:     height,
		InfoLines:  []string{"Context: prod"},
		Shortcuts:  rows,
		Content:    c.BorderedContent("body", innerW+2, innerH),
		Breadcrumb: []Crumb{{Label: "view", Leaf: true}},
	})
	assertFrameFits(t, c, out, width, height, "<5>")
	lines := strings.Split(out, "\n")
	want := strings.Index(stripANSI(lines[0]), "<1>")
	for i, key := range []string{"<1>", "<2>", "<3>", "<4>", "<5>"} {
		if got := strings.Index(stripANSI(lines[i]), key); got != want {
			t.Errorf("row %d: %s at col %d; want %d (aligned with the rows beside the logo)", i, key, got, want)
		}
	}
}

func TestNew_ShortcutWidthsConfigurable(t *testing.T) {
	t.Parallel()
	c := New(Config{ShortcutKeyWidth: 5, ShortcutDescWidth: 14})
	if c.ShortcutKeyWidth != 5 || c.ShortcutDescWidth != 14 {
		t.Fatalf("widths not stored: key=%d desc=%d", c.ShortcutKeyWidth, c.ShortcutDescWidth)
	}
	got := stripANSI(c.ShortcutPair("<1>", "Pods", "<a>", "Attach"))
	if want := "<1>  Pods          <a>  Attach"; got != want {
		t.Errorf("ShortcutPair = %q; want %q", got, want)
	}
	// Defaults are unchanged for apps that set neither.
	got = stripANSI(New(Config{}).ShortcutPair("<1>", "Pods", "<a>", "Attach"))
	if want := "<1>      Pods      <a>      Attach"; got != want {
		t.Errorf("default ShortcutPair = %q; want %q", got, want)
	}
}

// TestRender_InfoPanelOverflowStaysInFrame: the top section is clamped
// to TopSectionRows whatever overflows it, so InfoLines past
// InfoPanelRows cannot push the footer off the terminal either.
func TestRender_InfoPanelOverflowStaysInFrame(t *testing.T) {
	t.Parallel()
	const width, height = 120, 30
	c := New(Config{})
	info := make([]string, 9)
	for i := range info {
		info[i] = fmt.Sprintf("Line%d: value", i)
	}
	innerW, innerH := c.ContentInnerSize(width, height, false, false, false, false)
	out := c.Render(Frame{
		Width:      width,
		Height:     height,
		InfoLines:  info,
		Content:    c.BorderedContent("body", innerW+2, innerH),
		Breadcrumb: []Crumb{{Label: "view", Leaf: true}},
	})
	if n := len(strings.Split(out, "\n")); n != height {
		t.Errorf("Render produced %d lines; want %d", n, height)
	}
}
