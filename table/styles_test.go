package table

import (
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/blairham/tuikit/theme"
)

func TestTruncate(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in     string
		want   string
		maxLen int
	}{
		{"hello", "hello", 5},       // exact fit, no truncate
		{"hello world", "hell…", 5}, // truncated with ellipsis
		{"x", "x", 1},               // len equals maxLen, no truncate
		{"xy", "…", 1},              // maxLen=1 + needs truncate → ellipsis only
		{"", "", 5},
		{"hello", "hello", 0},  // maxLen<1 → no limit
		{"hello", "hello", -1}, // maxLen<1 → no limit
		{"x", "x", 0},
		// #3: a byte cut used to land inside "→" and tear it.
		{"5433→5432/tcp 5433→5432/tcp", "5433→5432/tcp 5433→54…", 22},
		{"a→b", "a→b", 3}, // 3 cells (5 bytes) fits in 3
		{"→→→→", "→…", 2}, // cut by cell, not byte
		{"日本語", "日本語", 6}, // 3 wide runes = 6 cells, fits exactly
		{"日本語", "日…", 4},  // 日(2)+…(1); 本 would overrun
		{"日本語", "日…", 3},
		{"日本語", "…", 2}, // no room for a wide rune beside the ellipsis
		{"日本語", "…", 1},
		{"\x1b[31mhello world\x1b[0m", "\x1b[31mhell…\x1b[0m", 5}, // SGR kept, not counted
	}
	for _, c := range cases {
		got := Truncate(c.in, c.maxLen)
		if got != c.want {
			t.Errorf("Truncate(%q, %d) = %q; want %q", c.in, c.maxLen, got, c.want)
		}
	}
}

// TestTruncateInvariants checks that for every cut point the result is
// valid UTF-8, never wider than maxLen cells, and that strings already
// within budget come back unchanged.
func TestTruncateInvariants(t *testing.T) {
	t.Parallel()
	inputs := []string{
		"5433→5432/tcp 5433→5432/tcp",
		"日本語のテキスト",
		"mixed ascii 和 wide 字 runes",
		"emoji 🚀🚀🚀 rocket",
		"plain ascii only",
		"é́ combining",
	}
	for _, in := range inputs {
		w := ansi.StringWidth(in)
		for maxLen := 1; maxLen <= w+2; maxLen++ {
			got := Truncate(in, maxLen)
			if !utf8.ValidString(got) {
				t.Errorf("Truncate(%q, %d) = %q: invalid UTF-8", in, maxLen, got)
			}
			if gw := ansi.StringWidth(got); gw > maxLen {
				t.Errorf("Truncate(%q, %d) = %q: width %d > %d", in, maxLen, got, gw, maxLen)
			}
			if w <= maxLen && got != in {
				t.Errorf("Truncate(%q, %d) = %q: fits (width %d) but was changed", in, maxLen, got, w)
			}
			if w > maxLen && !strings.HasSuffix(got, "…") {
				t.Errorf("Truncate(%q, %d) = %q: cut without ellipsis", in, maxLen, got)
			}
		}
	}
}

// TestKeyMap pins #9: j/k are deliberately unbound on LineUp/LineDown
// (viewfsm.TranslateNavKey owns them), the arrows stay bound, and the
// help text does not advertise keys that do nothing.
func TestKeyMap(t *testing.T) {
	t.Parallel()
	km := KeyMap()
	if got, want := km.LineUp.Keys(), []string{"up"}; !slices.Equal(got, want) {
		t.Errorf("LineUp keys = %v; want %v", got, want)
	}
	if got, want := km.LineDown.Keys(), []string{"down"}; !slices.Equal(got, want) {
		t.Errorf("LineDown keys = %v; want %v", got, want)
	}
	if h := km.LineUp.Help(); h.Key != "↑" || h.Desc != "up" {
		t.Errorf("LineUp help = %q/%q; want \"↑\"/\"up\"", h.Key, h.Desc)
	}
	if h := km.LineDown.Help(); h.Key != "↓" || h.Desc != "down" {
		t.Errorf("LineDown help = %q/%q; want \"↓\"/\"down\"", h.Key, h.Desc)
	}
	// The rest of the default keymap is untouched.
	if !slices.Contains(km.GotoTop.Keys(), "g") || !slices.Contains(km.GotoBottom.Keys(), "G") {
		t.Errorf("GotoTop/GotoBottom lost g/G: %v / %v", km.GotoTop.Keys(), km.GotoBottom.Keys())
	}
}

func TestFitHeight(t *testing.T) {
	t.Parallel()
	cases := []struct{ rows, maxH, want int }{
		{0, 10, 1},   // floor at 1
		{3, 10, 4},   // rows + header
		{20, 10, 10}, // clamped to maxH
		{5, 0, 6},    // no max → desired
	}
	for _, c := range cases {
		got := FitHeight(c.rows, c.maxH)
		if got != c.want {
			t.Errorf("FitHeight(%d, %d) = %d; want %d", c.rows, c.maxH, got, c.want)
		}
	}
}

func TestPaintModeFor(t *testing.T) {
	t.Parallel()
	if PaintModeFor(theme.Default()) != PaintModeFull {
		t.Error("Default theme should use PaintModeFull")
	}
	if PaintModeFor(theme.NoPaintBackground()) != PaintModeNone {
		t.Error("NoPaintBackground theme should use PaintModeNone")
	}
}

func TestFixSelectedRow_NonePassesNonSelectedLinesUnchanged(t *testing.T) {
	t.Parallel()
	// Lines without the Selected-bg marker pass through untouched
	// under PaintModeNone — the chrome doesn't paint cell backgrounds.
	in := "\x1b[38;2;255;0;0msome\x1b[m \x1b[1;38;2;0;255;0mthing\x1b[m"
	if got := FixSelectedRow(in, PaintModeNone); got != in {
		t.Errorf("non-selected line should pass through PaintModeNone; got %q", got)
	}
}

func TestFixSelectedRow_NoneStillFixesSelectedLines(t *testing.T) {
	t.Parallel()
	// A row containing the selected-bg marker hits fixSelectedLine
	// under either paint mode. Without this, bubbles table's per-cell
	// `\x1b[m` resets clear the row-level Selected background after
	// column 1 — harmless under PaintModeFull (covered by its
	// fallback) but the only fix for PaintModeNone themes.
	in := "\x1b[1;48;2;135;206;250m" +
		"\x1b[38;2;100;100;100mcell-a\x1b[m" +
		" " +
		"\x1b[38;2;100;100;100mcell-b\x1b[m"

	got := FixSelectedRow(in, PaintModeNone)
	if got == in {
		t.Fatal("PaintModeNone should rewrite mid-row resets on selected lines")
	}
	// The mid-row reset before "cell-b" should be followed by a
	// re-application of the selected bg + black fg so the highlight
	// continues across the row.
	reapply := "\x1b[m\x1b[1;38;2;0;0;0;" + selectedBgMarker + "m"
	if !strings.Contains(got, reapply) {
		t.Errorf("rewritten line missing mid-row reapply sequence; got %q", got)
	}
	// Cell foregrounds should be rewritten to black so text stays
	// legible on the highlight regardless of the original color.
	if strings.Contains(got, "38;2;100;100;100") {
		t.Errorf("original cell foreground should be rewritten to black; got %q", got)
	}
	// The trailing reset is preserved so the row doesn't bleed style
	// onto subsequent lines.
	if !strings.HasSuffix(got, "\x1b[m") {
		t.Errorf("rewritten line should end with a clean reset; got %q", got)
	}
}

func TestFixSelectedRow_FullStillPaintsNonSelected(t *testing.T) {
	t.Parallel()
	// Under PaintModeFull, non-selected lines containing multiple
	// cell-resets get the black-bg paint pass so cell padding stays
	// opaque between cells. Only the trailing reset is left alone to
	// keep line endings clean.
	in := "\x1b[38;2;255;0;0mcell-a\x1b[m \x1b[38;2;0;255;0mcell-b\x1b[m"
	got := FixSelectedRow(in, PaintModeFull)
	if got == in {
		t.Error("PaintModeFull should rewrite mid-row resets on non-selected lines")
	}
	if !strings.Contains(got, blackBgMarker) {
		t.Errorf("non-selected line under PaintModeFull should reapply the black bg; got %q", got)
	}
}

// TestStylesUseTheSkinsTableColors: a theme's TableText and TableHeader
// color the cells and header; unset, they are the colors used before.
func TestStylesUseTheSkinsTableColors(t *testing.T) {
	th := theme.Default()
	th.TableText = lipgloss.Color("#f8f8f2")
	th.TableHeader = lipgloss.Color("#f1fa8c")
	s := Styles(th)
	if s.Cell.GetForeground() != th.TableText || s.Header.GetForeground() != th.TableHeader {
		t.Errorf("cell %v header %v, want the skin's", s.Cell.GetForeground(), s.Header.GetForeground())
	}
	d := Styles(theme.Default())
	if d.Cell.GetForeground() != theme.Default().Selection || d.Header.GetForeground() != theme.Default().Value {
		t.Error("unset table colors changed the default look")
	}
}
