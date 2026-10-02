package tail

import (
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
