package theme

import (
	"strings"
	"testing"
)

func TestDefault_PaintsBackground(t *testing.T) {
	t.Parallel()
	theme := Default()
	if !theme.PaintBackground {
		t.Error("Default theme should set PaintBackground=true")
	}
	// On() with PaintBackground=true should emit a Background ANSI code.
	rendered := theme.On(theme.Value).Render("x")
	if !strings.Contains(rendered, "48;") {
		t.Errorf("expected background ANSI in %q", rendered)
	}
}

func TestNoPaintBackground_OmitsBackground(t *testing.T) {
	t.Parallel()
	theme := NoPaintBackground()
	if theme.PaintBackground {
		t.Error("NoPaintBackground theme should set PaintBackground=false")
	}
	rendered := theme.On(theme.Value).Render("x")
	if strings.Contains(rendered, "48;") {
		t.Errorf("did not expect background ANSI in %q", rendered)
	}
}

func TestDefault_PrebuiltStylesPopulated(t *testing.T) {
	t.Parallel()
	theme := Default()
	// Spot-check that styles render without producing empty strings,
	// which would indicate a missing color assignment.
	cases := map[string]string{
		"InfoLabel":    theme.InfoLabel.Render("Env:"),
		"InfoValue":    theme.InfoValue.Render("prd"),
		"ShortcutKey":  theme.ShortcutKey.Render("<enter>"),
		"ShortcutDesc": theme.ShortcutDesc.Render("Tail"),
		"LogoStyle":    theme.LogoStyle.Render("X"),
		"Title":        theme.Title.Render("title"),
		"Error":        theme.Error.Render("oops"),
		"Footer":       theme.Footer.Render("breadcrumb"),
		"FilterStyle":  theme.FilterStyle.Render("/"),
		"PromptStyle":  theme.PromptStyle.Render("active"),
	}
	for name, rendered := range cases {
		if rendered == "" {
			t.Errorf("%s rendered empty", name)
		}
		if !strings.Contains(rendered, "\x1b[") {
			t.Errorf("%s has no ANSI escapes: %q", name, rendered)
		}
	}
}

func TestOn_AppliesForeground(t *testing.T) {
	t.Parallel()
	theme := Default()
	rendered := theme.On(theme.Accent).Render("test")
	if !strings.Contains(rendered, "38;") {
		t.Errorf("expected foreground ANSI in %q", rendered)
	}
}
