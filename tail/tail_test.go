package tail

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"github.com/blairham/tuikit/theme"
)

func TestNew_FollowDefaultsOn(t *testing.T) {
	t.Parallel()
	m := New()
	if !m.Follow() {
		t.Error("New should default Follow=true")
	}
	if m.Ready() {
		t.Error("New should not be Ready until Resize")
	}
	if m.LineCount() != 0 || m.VisibleCount() != 0 {
		t.Error("New should have zero buffered lines")
	}
}

func TestAppendLine_BuffersAndFilters(t *testing.T) {
	t.Parallel()
	m := New()
	m.AppendLine("hello world")
	m.AppendLine("warn: thing")
	m.AppendLine("err: another")
	if m.LineCount() != 3 {
		t.Errorf("LineCount = %d; want 3", m.LineCount())
	}
	if m.VisibleCount() != 3 {
		t.Errorf("VisibleCount = %d; want 3", m.VisibleCount())
	}

	m.SetFilter("err|warn")
	if m.VisibleCount() != 2 {
		t.Errorf("after filter, VisibleCount = %d; want 2", m.VisibleCount())
	}
	if m.LineCount() != 3 {
		t.Errorf("LineCount should preserve all lines through filter, got %d", m.LineCount())
	}

	m.SetFilter("")
	if m.VisibleCount() != 3 {
		t.Errorf("clearing filter should restore all lines, got %d", m.VisibleCount())
	}
}

func TestSetFollow(t *testing.T) {
	t.Parallel()
	m := New()
	m.SetFollow(false)
	if m.Follow() {
		t.Error("SetFollow(false) should disable follow")
	}
	m.SetFollow(true)
	if !m.Follow() {
		t.Error("SetFollow(true) should enable follow")
	}
}

func TestHandleScrollKey_PreReadyIsNoop(t *testing.T) {
	t.Parallel()
	m := New()
	if m.HandleScrollKey("G") {
		t.Error("HandleScrollKey before Resize should return false")
	}
}

func TestHandleScrollKey_PostReady(t *testing.T) {
	t.Parallel()
	m := New()
	m.Resize(80, 24)
	cases := []struct {
		key        string
		wantFollow bool
		wantHandle bool
	}{
		{"ctrl+f", false, true},
		{"G", true, true},
		{"g", false, true},
		{"f", true, true}, // toggles from false → true
		{"f", false, true},
		{"x", false, false}, // unknown key
	}
	m.SetFollow(false) // start known
	for _, c := range cases {
		got := m.HandleScrollKey(c.key)
		if got != c.wantHandle {
			t.Errorf("HandleScrollKey(%q) handled=%v; want %v", c.key, got, c.wantHandle)
		}
		if c.wantHandle && m.Follow() != c.wantFollow {
			t.Errorf("HandleScrollKey(%q) Follow()=%v; want %v", c.key, m.Follow(), c.wantFollow)
		}
	}
}

func TestView_NotReadyReturnsEmpty(t *testing.T) {
	t.Parallel()
	m := New()
	if m.View() != "" {
		t.Error("View before Resize should return empty string")
	}
}

func TestView_BackgroundSurvivesCombinedReset(t *testing.T) {
	t.Parallel()
	m := New()
	m.SetBackground(lipgloss.Color("#000000"))
	m.Resize(40, 1)
	m.AppendLine("[\x1b[0;32m  OK  \x1b[0m] Started")
	bg := theme.BackgroundSeq(lipgloss.Color("#000000"))
	got := m.View()
	if !strings.Contains(got, "\x1b[0;32m"+bg+"  OK  ") {
		t.Errorf("background not re-asserted after combined reset \\x1b[0;32m: %q", got)
	}
}

// TestWrap pins soft-wrap, including a wrap set before the first Resize:
// the viewport is created lazily there and must pick the setting up.
func TestWrap(t *testing.T) {
	long := strings.Repeat("a", 30) + "TAIL"
	for _, tc := range []struct {
		name     string
		setFirst bool
		wrap     bool
	}{
		{name: "off", wrap: false},
		{name: "on after resize", wrap: true},
		{name: "on before resize", wrap: true, setFirst: true},
	} {
		m := New()
		if tc.setFirst {
			m.SetWrap(tc.wrap)
		}
		m.Resize(20, 5)
		if !tc.setFirst {
			m.SetWrap(tc.wrap)
		}
		m.AppendLine(long)
		if got := strings.Contains(m.View(), "TAIL"); got != tc.wrap {
			t.Errorf("%s: end of a long line visible = %v, want %v\n%s", tc.name, got, tc.wrap, m.View())
		}
		if m.Wrap() != tc.wrap {
			t.Errorf("%s: Wrap() = %v", tc.name, m.Wrap())
		}
	}
}

func TestClearKeepsStreaming(t *testing.T) {
	m := New()
	m.Resize(40, 5)
	m.AppendLines([]string{"one", "two"})
	m.Clear()
	if m.LineCount() != 0 || m.VisibleCount() != 0 || strings.Contains(m.View(), "two") {
		t.Fatalf("Clear left %d lines: %q", m.LineCount(), m.View())
	}
	m.AppendLine("three")
	if m.LineCount() != 1 || !strings.Contains(m.View(), "three") {
		t.Errorf("after Clear, a new line: count %d view %q", m.LineCount(), m.View())
	}
}

func TestVisibleLinesIsTheFilteredCopy(t *testing.T) {
	m := New()
	m.AppendLines([]string{"error: a", "info: b", "error: c"})
	m.SetFilter("error")
	got := m.VisibleLines()
	if strings.Join(got, "|") != "error: a|error: c" {
		t.Fatalf("VisibleLines = %q", got)
	}
	got[0] = "changed"
	if m.VisibleLines()[0] != "error: a" {
		t.Error("VisibleLines returned the buffer itself, not a copy")
	}
}

func TestMarkerPassesEveryFilter(t *testing.T) {
	// Filter set before the marker arrives: the append path.
	m := New()
	m.SetFilter("error")
	m.AppendLines([]string{"error: a", "info: b"})
	m.AppendMarker("── mark ──")
	m.AppendLines([]string{"info: c", "error: d"})
	if got := strings.Join(m.VisibleLines(), "|"); got != "error: a|── mark ──|error: d" {
		t.Errorf("filtered before the mark: %q", got)
	}

	// Filter set afterwards: the rebuild path, which also keeps its place.
	m.SetFilter("")
	if got := strings.Join(m.VisibleLines(), "|"); got != "error: a|info: b|── mark ──|info: c|error: d" {
		t.Errorf("unfiltered: %q", got)
	}
	m.SetFilter("info")
	if got := strings.Join(m.VisibleLines(), "|"); got != "info: b|── mark ──|info: c" {
		t.Errorf("filtered after the mark: %q", got)
	}
	if m.VisibleCount() != 3 || m.LineCount() != 5 {
		t.Errorf("counts: visible %d of %d, want 3 of 5", m.VisibleCount(), m.LineCount())
	}
}

func TestMarkerStaysPutAndClears(t *testing.T) {
	m := New()
	m.Resize(40, 10)
	m.AppendLine("new")
	m.AppendMarker("── mark ──")
	m.SetFilter("old")
	m.PrependLines([]string{"old one", "other"})
	if got := strings.Join(m.VisibleLines(), "|"); got != "old one|── mark ──" {
		t.Errorf("after prepending older lines: %q", got)
	}
	if !strings.Contains(m.View(), "── mark ──") {
		t.Errorf("the marker is not drawn: %q", m.View())
	}
	m.Clear()
	m.SetFilter("")
	if m.LineCount() != 0 || m.VisibleCount() != 0 {
		t.Errorf("Clear kept %d lines", m.LineCount())
	}
}

func numbered(from, to int) []string {
	out := make([]string, 0, to-from+1)
	for i := from; i <= to; i++ {
		out = append(out, fmt.Sprintf("line %02d", i))
	}
	return out
}

// TestMaxLinesDropsTheOldest: past the cap the oldest lines go, markers
// with them, and the visible buffer stays the filter over what is left.
func TestMaxLinesDropsTheOldest(t *testing.T) {
	m := New()
	m.SetMaxLines(4)
	m.AppendMarker("-- mark --")
	m.AppendLines(numbered(1, 5))
	if got := strings.Join(m.VisibleLines(), "|"); got != "line 02|line 03|line 04|line 05" || m.LineCount() != 4 {
		t.Errorf("capped buffer: %q (%d lines)", got, m.LineCount())
	}
	m.SetFilter("[24]")
	m.AppendLines(numbered(6, 8))
	if got := strings.Join(m.VisibleLines(), "|"); got != "" {
		t.Errorf("filtered after more drops: %q", got)
	}
	m.SetFilter("")
	if got := strings.Join(m.VisibleLines(), "|"); got != "line 05|line 06|line 07|line 08" {
		t.Errorf("unfiltered after drops: %q", got)
	}
}

// TestMaxLinesKeepsAScrolledViewStill: scrolled back, the lines on screen
// stay on screen as older ones are dropped above them.
func TestMaxLinesKeepsAScrolledViewStill(t *testing.T) {
	m := New()
	m.Resize(20, 2)
	m.SetMaxLines(10)
	m.AppendLines(numbered(1, 10))
	m.HandleScrollKey("g")
	m.HandleScrollKey("j")
	m.HandleScrollKey("j")
	before := strings.Split(m.View(), "\n")[0]
	m.AppendLines(numbered(11, 12))
	if after := strings.Split(m.View(), "\n")[0]; !strings.Contains(before, "line 03") || after != before {
		t.Errorf("top line moved from %q to %q", before, after)
	}
	if m.LineCount() != 10 {
		t.Errorf("%d lines kept, want 10", m.LineCount())
	}
}

// TestMaxLinesTrimsAtOnceAndPrependKeepsTheNewest: lowering the cap trims
// straight away; history prepended into a full buffer is what is dropped;
// 0 is no cap.
func TestMaxLinesTrimsAtOnceAndPrependKeepsTheNewest(t *testing.T) {
	m := New()
	m.AppendLines(numbered(1, 6))
	m.SetMaxLines(3)
	if got := strings.Join(m.VisibleLines(), "|"); got != "line 04|line 05|line 06" {
		t.Errorf("lowered cap: %q", got)
	}
	m.PrependLines(numbered(1, 3))
	if got := strings.Join(m.VisibleLines(), "|"); got != "line 04|line 05|line 06" {
		t.Errorf("prepend into a full buffer: %q", got)
	}
	m.SetMaxLines(0)
	m.AppendLines(numbered(7, 20))
	if m.LineCount() != 17 || m.MaxLines() != 0 {
		t.Errorf("uncapped: %d lines, cap %d", m.LineCount(), m.MaxLines())
	}
}

// TestReplaceLinesKeepsThePlace: a refetched document keeps the filter and,
// scrolled, the offset — clamped when it shrinks — while a following view
// stays at the bottom.
func TestReplaceLinesKeepsThePlace(t *testing.T) {
	m := New()
	m.Resize(20, 2)
	m.SetFollow(false)
	m.ReplaceLines(numbered(1, 10))
	m.HandleScrollKey("j")
	m.HandleScrollKey("j")
	m.HandleScrollKey("j")
	before := strings.Split(m.View(), "\n")[0]
	m.ReplaceLines(numbered(1, 10))
	if after := strings.Split(m.View(), "\n")[0]; !strings.Contains(before, "line 04") || after != before {
		t.Errorf("a refresh moved the view from %q to %q", before, after)
	}
	m.ReplaceLines(numbered(1, 3))
	if top := strings.Split(m.View(), "\n")[0]; !strings.Contains(top, "line 02") {
		t.Errorf("shorter content: top line %q, want the offset clamped to line 02", top)
	}

	m.SetFilter("line 0[13]")
	m.ReplaceLines(numbered(1, 5))
	if got := strings.Join(m.VisibleLines(), "|"); got != "line 01|line 03" {
		t.Errorf("filter after a refresh: %q", got)
	}

	f := New()
	f.Resize(20, 2)
	f.ReplaceLines(numbered(1, 5))
	f.ReplaceLines(numbered(1, 8))
	if last := strings.Split(f.View(), "\n")[1]; !strings.Contains(last, "line 08") {
		t.Errorf("a following view left the bottom: %q", last)
	}
}
