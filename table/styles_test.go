package table

import (
	"strings"
	"testing"

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
	}
	for _, c := range cases {
		got := Truncate(c.in, c.maxLen)
		if got != c.want {
			t.Errorf("Truncate(%q, %d) = %q; want %q", c.in, c.maxLen, got, c.want)
		}
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
