// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package theme

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
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
		"InfoLabel":     theme.InfoLabel.Render("Env:"),
		"InfoValue":     theme.InfoValue.Render("prd"),
		"ShortcutKey":   theme.ShortcutKey.Render("<enter>"),
		"ShortcutDesc":  theme.ShortcutDesc.Render("Tail"),
		"LogoStyle":     theme.LogoStyle.Render("X"),
		"Title":         theme.Title.Render("title"),
		"Error":         theme.Error.Render("oops"),
		"Footer":        theme.Footer.Render("breadcrumb"),
		"FilterStyle":   theme.FilterStyle.Render("/"),
		"PromptStyle":   theme.PromptStyle.Render("active"),
		"MarkStyle":     theme.MarkStyle.Render("pod-1"),
		"SearchMatch":   theme.SearchMatch.Render("hit"),
		"SearchCurrent": theme.SearchCurrent.Render("hit"),
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

// dockerBlueSGR is the truecolor foreground sequence lipgloss emits for
// #2496ED (36, 150, 237).
const dockerBlueSGR = "38;2;36;150;237"

func TestRebuild_RecolorsPrebuiltStyles(t *testing.T) {
	t.Parallel()
	for name, ctor := range map[string]func() Theme{
		"Default":           Default,
		"NoPaintBackground": NoPaintBackground,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			th := ctor()
			th.Logo = lipgloss.Color("#2496ED")
			th.Accent = lipgloss.Color("#2496ED")

			// The contract: assigning a field alone does not reach the
			// pre-built styles.
			if got := th.LogoStyle.Render("X"); strings.Contains(got, dockerBlueSGR) {
				t.Fatalf("LogoStyle picked up the new color without Rebuild: %q", got)
			}

			th.Rebuild()
			for style, got := range map[string]string{
				"LogoStyle": th.LogoStyle.Render("X"),
				"Title":     th.Title.Render("title"),
			} {
				if !strings.Contains(got, dockerBlueSGR) {
					t.Errorf("%s after Rebuild = %q, want foreground %s", style, got, dockerBlueSGR)
				}
			}
		})
	}
}

func TestRebuild_PreservesBackgroundMode(t *testing.T) {
	t.Parallel()

	painted := Default()
	painted.Logo = lipgloss.Color("#2496ED")
	painted.Rebuild()
	if got := painted.LogoStyle.Render("X"); !strings.Contains(got, "48;") {
		t.Errorf("Default after Rebuild lost its background: %q", got)
	}

	bare := NoPaintBackground()
	bare.Logo = lipgloss.Color("#2496ED")
	bare.Rebuild()
	if bare.PaintBackground {
		t.Error("Rebuild flipped PaintBackground on a NoPaintBackground theme")
	}
	for style, got := range map[string]string{
		"LogoStyle":   bare.LogoStyle.Render("X"),
		"TableBorder": bare.TableBorder.Render("X"),
	} {
		if strings.Contains(got, "48;") {
			t.Errorf("NoPaintBackground %s after Rebuild paints a background: %q", style, got)
		}
	}
}

// lightSkyBlueSGR and dodgerBlueSGR are the truecolor foreground
// sequences for #87CEFA (k9s's frame focusColor) and #1E90FF.
const (
	lightSkyBlueSGR = "38;2;135;206;250"
	dodgerBlueSGR   = "38;2;30;144;255"
)

func TestDefault_TableBorderIsFocusColor(t *testing.T) {
	t.Parallel()
	for name, ctor := range map[string]func() Theme{
		"Default":           Default,
		"NoPaintBackground": NoPaintBackground,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := ctor().TableBorder.Render("X")
			if !strings.Contains(got, lightSkyBlueSGR) {
				t.Errorf("TableBorder = %q, want border foreground %s", got, lightSkyBlueSGR)
			}
			if strings.Contains(got, dodgerBlueSGR) {
				t.Errorf("TableBorder = %q still carries the unfocused %s", got, dodgerBlueSGR)
			}
		})
	}
}

func TestFocusBorder_NilFallsBackToBorder(t *testing.T) {
	t.Parallel()
	th := Theme{Border: lipgloss.Color("#1E90FF")}
	if th.FocusBorder() != th.Border {
		t.Errorf("FocusBorder() = %v, want Border %v when BorderFocus is nil", th.FocusBorder(), th.Border)
	}
	th.Rebuild()
	got := th.TableBorder.Render("X")
	if !strings.Contains(got, dodgerBlueSGR) {
		t.Errorf("TableBorder with nil BorderFocus = %q, want Border foreground %s", got, dodgerBlueSGR)
	}
}

func TestRebuild_PicksUpBorderFocus(t *testing.T) {
	t.Parallel()
	th := Default()
	th.BorderFocus = lipgloss.Color("#2496ED")
	if got := th.TableBorder.Render("X"); strings.Contains(got, dockerBlueSGR) {
		t.Fatalf("TableBorder picked up BorderFocus without Rebuild: %q", got)
	}
	th.Rebuild()
	if got := th.TableBorder.Render("X"); !strings.Contains(got, dockerBlueSGR) {
		t.Errorf("TableBorder after Rebuild = %q, want foreground %s", got, dockerBlueSGR)
	}
}
