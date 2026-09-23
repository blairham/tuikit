package tail

import "testing"

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
