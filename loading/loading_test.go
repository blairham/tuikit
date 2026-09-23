package loading

import (
	"strings"
	"testing"

	"github.com/blairham/tuikit/theme"
)

func TestTipRotation(t *testing.T) {
	t.Parallel()
	m := New(theme.Default(), []string{"first", "second"})

	if m.Tip() != "first" {
		t.Fatalf("initial tip = %q, want %q", m.Tip(), "first")
	}

	// Each rotateEvery ticks advances one tip.
	for range defaultRotateEvery {
		m.Update(TickMsg{})
	}
	if m.Tip() != "second" {
		t.Errorf("after %d ticks tip = %q, want %q", defaultRotateEvery, m.Tip(), "second")
	}

	// Wraps back around.
	for range defaultRotateEvery {
		m.Update(TickMsg{})
	}
	if m.Tip() != "first" {
		t.Errorf("after wrap tip = %q, want %q", m.Tip(), "first")
	}
}

func TestResetReturnsToFirstTip(t *testing.T) {
	t.Parallel()
	m := New(theme.Default(), []string{"a", "b", "c"})
	for range defaultRotateEvery {
		m.Update(TickMsg{})
	}
	if m.Tip() == "a" {
		t.Fatal("expected tip to have advanced before reset")
	}
	m.Reset()
	if m.Tip() != "a" {
		t.Errorf("after Reset tip = %q, want %q", m.Tip(), "a")
	}
}

func TestViewNoTips(t *testing.T) {
	t.Parallel()
	m := New(theme.Default(), nil)
	if m.Tip() != "" {
		t.Errorf("Tip() with no tips = %q, want empty", m.Tip())
	}
	// View should still render the spinner glyph, with no trailing tip text.
	if got := m.View(); strings.TrimSpace(got) == "" {
		t.Error("View() with no tips should still render the spinner")
	}
}

func TestCenteredFillsBox(t *testing.T) {
	t.Parallel()
	m := New(theme.Default(), []string{"loading"})
	out := m.Centered(40, 5)
	if lines := strings.Count(out, "\n") + 1; lines != 5 {
		t.Errorf("Centered height = %d lines, want 5", lines)
	}
}
