package chrome

import (
	"strings"
	"testing"

	"github.com/blairham/tuikit/theme"
)

// k9sInfo is the seven-row info panel from k9s's header (Context,
// Cluster, User, K9s Rev, K8s Rev, CPU, MEM).
func k9sInfo() []string {
	return []string{
		"Context:  kind-dev",
		"Cluster:  kind-dev",
		"User:     kind-dev-admin",
		"K9s Rev:  v0.40.10",
		"K8s Rev:  v1.33.1",
		"CPU:      12%",
		"MEM:      48%",
	}
}

// k9sViews is nine namespace hotkeys, <0>..<8>.
func k9sViews() []Shortcut {
	names := []string{"all", "default", "kube-system", "monitoring", "ingress", "cert-manager", "argocd", "vault", "logging"}
	out := make([]Shortcut, len(names))
	for i, n := range names {
		out[i] = Shortcut{Key: "<" + string(rune('0'+i)) + ">", Desc: n}
	}
	return out
}

// k9sActions is fourteen pod actions.
func k9sActions() []Shortcut {
	return []Shortcut{
		{"<a>", "Attach"}, {"<ctrl-d>", "Delete"}, {"<d>", "Describe"},
		{"<e>", "Edit"}, {"<?>", "Help"}, {"<ctrl-k>", "Kill"},
		{"<l>", "Logs"}, {"<p>", "Logs Previous"}, {"<shift-f>", "Port-Forward"},
		{"<z>", "Sanitize"}, {"<s>", "Shell"}, {"<o>", "Show Node"},
		{"<n>", "Show PortForward"}, {"<t>", "Trigger Cron"},
	}
}

// k9sLogo is a six-line logo drawn in '@' so its cells are
// distinguishable from shortcut text. Its lines are ragged (no trailing
// padding), as hand-written ASCII art usually is.
func k9sLogo() []string {
	return []string{
		" @@@@  @@@@@ @@@@@",
		"@@  @@ @@      @@",
		"@@  @@ @@@@    @@",
		"@@  @@ @@      @@",
		" @@@@  @@@@@   @@",
		"                 @@",
	}
}

const (
	k9sWidth  = 220
	k9sHeight = 45
)

// renderK9sHeader renders the k9s reference screen and returns the
// stripped lines of the whole frame.
func renderK9sHeader(t *testing.T, c Chrome) []string {
	t.Helper()
	innerW, innerH := c.ContentInnerSize(k9sWidth, k9sHeight, false, false, false, false)
	out := c.Render(Frame{
		Width:      k9sWidth,
		Height:     k9sHeight,
		InfoLines:  k9sInfo(),
		Shortcuts:  c.ShortcutGrid(k9sViews(), k9sActions()),
		Content:    c.BorderedContent("body", innerW+2, innerH),
		Breadcrumb: []Crumb{{Label: "pods", Leaf: true}},
	})
	lines := strings.Split(out, "\n")
	if len(lines) != k9sHeight {
		t.Fatalf("Render produced %d lines; want %d", len(lines), k9sHeight)
	}
	for i := range lines {
		lines[i] = stripANSI(lines[i])
	}
	return lines
}

func k9sChrome(logo []string) Chrome {
	return New(Config{Theme: theme.Default(), Logo: logo, InfoPanelRows: 7})
}

// widestInfoEnd is the column just past the widest info line, counting
// the 1-cell left inset.
func widestInfoEnd() int {
	w := 0
	for _, l := range k9sInfo() {
		w = max(w, len(l))
	}
	return 1 + w
}

// TestRender_LogoModeShortcutsFollowInfoPanel is issue #16: with a logo
// the shortcut block starts a small gap after the info panel (as it
// does without one), instead of being right-aligned against the logo.
func TestRender_LogoModeShortcutsFollowInfoPanel(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		logo []string
	}{{"logo", k9sLogo()}, {"logoless", nil}} {
		lines := renderK9sHeader(t, k9sChrome(tc.logo))
		col := strings.Index(lines[0], "<0>")
		end := widestInfoEnd()
		if col < end+1 || col > end+4 {
			t.Errorf("%s: <0> starts at col %d; want within a few cells after the info panel (ends at %d)\n%s",
				tc.name, col, end, strings.Join(lines[:7], "\n"))
		}
	}
}

// TestRender_LogoPinnedToRightEdge: the logo is pinned as a block — its
// widest line ends one cell short of the right edge (the trailing
// inset), shorter lines start in the same column rather than being
// right-justified one by one, and the logo columns of the rows past it
// are blank.
func TestRender_LogoPinnedToRightEdge(t *testing.T) {
	t.Parallel()
	lines := renderK9sHeader(t, k9sChrome(k9sLogo()))
	logo := k9sLogo()
	logoW := 0
	for _, l := range logo {
		logoW = max(logoW, len(l))
	}
	start := k9sWidth - 1 - logoW
	for i := range 7 {
		row := []rune(lines[i])
		seg := string(row[start : k9sWidth-1])
		want := ""
		if i < len(logo) {
			want = logo[i]
		}
		want = padRight(want, logoW)
		if seg != want {
			t.Errorf("row %d logo columns = %q; want %q", i, seg, want)
		}
		if row[k9sWidth-1] != ' ' {
			t.Errorf("row %d: right inset holds %q", i, row[k9sWidth-1])
		}
	}
}

// TestShortcutGrid_ColumnHeightIsShortcutRows is issue #17: columns wrap
// at ShortcutRows (6 by default, as in k9s), not at TopSectionRows(),
// even when a seven-row info panel makes the header seven rows tall.
func TestShortcutGrid_ColumnHeightIsShortcutRows(t *testing.T) {
	t.Parallel()
	c := k9sChrome(k9sLogo())
	if got := c.TopSectionRows(); got != 7 {
		t.Fatalf("TopSectionRows = %d; want 7", got)
	}
	grid := c.ShortcutGrid(k9sViews(), k9sActions())
	if len(grid) != 6 {
		t.Fatalf("ShortcutGrid rows = %d; want 6\n%s", len(grid), stripANSI(strings.Join(grid, "\n")))
	}
	lines := renderK9sHeader(t, c)
	zero := strings.Index(lines[0], "<0>")
	if got := strings.Index(lines[5], "<5>"); got != zero {
		t.Errorf("<5> at row 5 col %d; want col %d (last row of the first view column)", got, zero)
	}
	six := strings.Index(lines[0], "<6>")
	if six <= zero {
		t.Errorf("<6> at row 0 col %d; want a second column after <0> (col %d)", six, zero)
	}
	// Each action column holds six entries: <a> heads the first,
	// <l> (the seventh action) heads the second, <n> the third.
	for _, key := range []string{"<a>", "<l>", "<n>"} {
		if !strings.Contains(lines[0], key) {
			t.Errorf("row 0 missing column head %s: %q", key, lines[0])
		}
	}
	if strings.Contains(lines[6], "<") {
		t.Errorf("row 6 holds a shortcut; columns should stop at 6 rows: %q", lines[6])
	}
}

// TestShortcutGrid_ShortcutRowsCappedAtTopSection: a configured column
// height can never exceed the header reservation.
func TestShortcutGrid_ShortcutRowsCappedAtTopSection(t *testing.T) {
	t.Parallel()
	c := New(Config{ShortcutRows: 3})
	if got := len(c.ShortcutGrid(nil, k9sActions())); got != 3 {
		t.Errorf("ShortcutRows 3: grid rows = %d; want 3", got)
	}
	c = New(Config{ShortcutRows: 8})
	if got, top := len(c.ShortcutGrid(nil, k9sActions())), c.TopSectionRows(); got != top || top != 8 {
		t.Errorf("ShortcutRows 8: grid rows = %d, TopSectionRows = %d; want 8", got, top)
	}
}

// TestRender_ShortHeaderStillFillsReservation: a header with fewer rows
// of content than TopSectionRows() still occupies exactly that many, so
// the content box ContentInnerSize sized starts right below it and ends
// right above the footer. Raising the default ShortcutRows to 6 made a
// five-row grid the common case for this.
func TestRender_ShortHeaderStillFillsReservation(t *testing.T) {
	t.Parallel()
	for _, logo := range [][]string{nil, k9sLogo()[:3]} {
		c := New(Config{Logo: logo})
		const w, h = 160, 30
		_, innerH := c.ContentInnerSize(w, h, false, false, false, false)
		lines := strings.Split(stripANSI(c.Render(Frame{
			Width:      w,
			Height:     h,
			InfoLines:  []string{"Context: prod"},
			Shortcuts:  []string{"a", "b"},
			Content:    c.BorderedContent("body", w, innerH),
			Breadcrumb: []Crumb{{Label: "view", Leaf: true}},
		})), "\n")
		if len(lines) != h {
			t.Fatalf("logo=%d: %d lines; want %d", len(logo), len(lines), h)
		}
		if top := c.TopSectionRows(); !strings.Contains(lines[top], "╭") {
			t.Errorf("logo=%d: row %d = %q; want the content's top border", len(logo), top, lines[top])
		}
		if !strings.Contains(lines[h-3], "╰") {
			t.Errorf("logo=%d: row %d = %q; want the content's bottom border", len(logo), h-3, lines[h-3])
		}
	}
}
