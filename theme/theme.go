package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Theme bundles every color and pre-built lipgloss style tuikit uses
// to render the chrome. Apps construct one (typically via [Default] or
// a project-specific constructor) and pass it into the chrome, table,
// viewfsm, and tail packages.
//
// PaintBackground governs whether every styled span explicitly paints
// Bg as its background. See the package doc for the trade-offs. The
// styles in [Theme] are pre-built to honor PaintBackground; if you
// build your own styles, use [Theme.On] so they follow the same rule.
//
// The pre-built styles are computed from the color fields when the
// theme is constructed. Assigning a color field (or PaintBackground)
// afterwards does not reach them on its own — call [Theme.Rebuild]
// once you are done adjusting fields:
//
//	t := theme.Default()
//	t.Logo = lipgloss.Color("#2496ED")
//	t.Accent = lipgloss.Color("#2496ED")
//	t.Rebuild() // LogoStyle and Title now use the new colors
//
//nolint:govet // field order favors readability over fieldalignment in this public API struct
type Theme struct {
	// Bg is the canvas color. With PaintBackground=true it's applied to
	// every styled span; with false, only the outer screen wrapper.
	Bg color.Color

	// Border is the unfocused border color: modals, secondary panes,
	// and the action-key shortcut color.
	Border color.Color

	// BorderFocus is the focused-frame border color — the main content
	// box ([Theme.TableBorder] and chrome.InjectBorderTitle's repainted
	// top line). k9s draws its focused frame in light sky blue and keeps
	// dodger blue for everything else. A nil BorderFocus (a Theme built
	// by hand) falls back to Border; see [Theme.FocusBorder].
	BorderFocus color.Color

	// Foreground accents. Apps recolor any of these when constructing
	// a custom theme. Changing one after [Default] or
	// [NoPaintBackground] requires [Theme.Rebuild] for the pre-built
	// styles below to pick it up.
	Accent     color.Color // primary highlight (cyan in Default)
	AccentAlt  color.Color // secondary highlight (fuchsia in Default)
	AccentBold color.Color // bold/numeric highlight (papaya whip in Default)
	Logo       color.Color // ASCII logo color
	Label      color.Color // info-panel label color
	Value      color.Color // info-panel value color
	Muted      color.Color // dim/separator text
	Prompt     color.Color // prompt-pill background
	PromptText color.Color // prompt-pill foreground (typically Bg)
	Filter     color.Color // filter prompt color
	// Command and filter bars, in k9s's prompt colors.
	CommandBorder color.Color // command-bar border (k9s: aqua)
	InputText     color.Color // text typed into the command and filter bars (k9s: cadetblue)
	Suggestion    color.Color // inline completion after the typed text (k9s: dodgerblue)
	Selection     color.Color // selected row bg (e.g. light sky blue)
	Mark          color.Color // marked-row text (k9s: palegreen)
	// Breadcrumb pills, k9s's frame.crumbs: black on aqua for the trail,
	// black on orange for the current view.
	BreadcrumbBg     color.Color // pill bg for the trail (k9s: aqua)
	BreadcrumbFg     color.Color // pill text (k9s: black)
	BreadcrumbActive color.Color // pill bg for the current view (k9s: orange)

	// Help overlay colors (k9s-style: muted-red title pill, green section
	// headers, cadet-blue description text, cyan border).
	HelpTitle   color.Color // help overlay border-title pill text
	HelpSection color.Color // help overlay section headers (RESOURCE, GENERAL, ...)
	HelpDesc    color.Color // help overlay description text
	HelpBorder  color.Color // help overlay outer border

	// Status colors.
	Status StatusColors

	// Pre-built styles. All honor PaintBackground. They are derived
	// from the color fields above at construction time and by
	// [Theme.Rebuild]; Rebuild overwrites any style assigned by hand.
	InfoLabel    lipgloss.Style
	InfoValue    lipgloss.Style
	ShortcutKey  lipgloss.Style // action keys (everything that isn't a view-switch digit)
	ShortcutView lipgloss.Style // k9s-style view-switch keys (<0>..<9>)
	ShortcutDesc lipgloss.Style
	LogoStyle    lipgloss.Style
	Title        lipgloss.Style
	Error        lipgloss.Style
	Success      lipgloss.Style
	MutedStyle   lipgloss.Style
	Footer       lipgloss.Style
	FilterStyle  lipgloss.Style
	PromptStyle  lipgloss.Style
	MarkStyle    lipgloss.Style // rows marked for a bulk action
	TableBorder  lipgloss.Style

	// PaintBackground controls whether tuikit paints Bg on every styled
	// span (true) or only the outer screen wrapper (false). See the
	// package doc for the trade-off.
	PaintBackground bool
}

// StatusColors groups foreground colors for the four conventional log
// levels / status indicators. Apps use them via [Theme.On] or directly.
type StatusColors struct {
	OK    color.Color
	Warn  color.Color
	Error color.Color
	Info  color.Color
}

// On returns a fresh [lipgloss.Style] with the given foreground color
// and the theme's PaintBackground rule applied. Use this anywhere you'd
// reach for lipgloss.NewStyle().Foreground(c) — it keeps the chrome
// consistent across themes.
func (t Theme) On(fg color.Color) lipgloss.Style {
	s := lipgloss.NewStyle().Foreground(fg)
	if t.PaintBackground {
		s = s.Background(t.Bg)
	}
	return s
}

// FocusBorder returns the color to draw the focused content border in:
// [Theme.BorderFocus], or [Theme.Border] when BorderFocus is nil (a
// Theme constructed by hand rather than from [Default]). Read the
// focused border color through this rather than the field, so a
// hand-built theme never renders a nil color.
func (t Theme) FocusBorder() color.Color {
	if t.BorderFocus != nil {
		return t.BorderFocus
	}
	return t.Border
}

// Default returns the canonical k9s-style theme: deep-black canvas,
// light-sky-blue focused border, dodger-blue unfocused borders, aqua/fuchsia accents, full background painting.
// This is the default look.
func Default() Theme {
	t := Theme{
		Bg:               lipgloss.Color("#000000"),
		Border:           lipgloss.Color("#1E90FF"), // DodgerBlue
		BorderFocus:      lipgloss.Color("#87CEFA"), // LightSkyBlue (k9s frame focusColor)
		Accent:           lipgloss.Color("#00FFFF"), // Aqua / Cyan
		AccentAlt:        lipgloss.Color("#FF00FF"), // Fuchsia
		AccentBold:       lipgloss.Color("#FFEFD5"), // PapayaWhip
		Logo:             lipgloss.Color("#FFA500"), // Orange
		Label:            lipgloss.Color("#FFA500"), // Orange
		Value:            lipgloss.Color("#FFFFFF"), // White
		Muted:            lipgloss.Color("#808080"), // Gray
		Prompt:           lipgloss.Color("#FFA500"), // Orange (bg of prompt pill)
		PromptText:       lipgloss.Color("#000000"), // Bg (fg of prompt pill)
		Filter:           lipgloss.Color("#2E8B57"), // SeaGreen (k9s-style)
		CommandBorder:    lipgloss.Color("#00FFFF"), // Aqua (k9s prompt border, command mode)
		InputText:        lipgloss.Color("#5F9EA0"), // CadetBlue (k9s prompt text)
		Suggestion:       lipgloss.Color("#1E90FF"), // DodgerBlue (k9s prompt suggestion)
		Selection:        lipgloss.Color("#87CEFA"), // LightSkyBlue
		Mark:             lipgloss.Color("#98FB98"), // PaleGreen (k9s table markColor)
		BreadcrumbBg:     lipgloss.Color("#00FFFF"), // Aqua (k9s crumbs bgColor)
		BreadcrumbFg:     lipgloss.Color("#000000"), // Black (k9s crumbs fgColor)
		BreadcrumbActive: lipgloss.Color("#FFA500"), // Orange (k9s crumbs activeColor)
		HelpTitle:        lipgloss.Color("#CD5C5C"), // IndianRed (k9s help title pill)
		HelpSection:      lipgloss.Color("#008000"), // Green (k9s help SectionColor)
		HelpDesc:         lipgloss.Color("#5F9EA0"), // CadetBlue (k9s help description text)
		HelpBorder:       lipgloss.Color("#00FFFF"), // Cyan (k9s help overlay border)
		Status: StatusColors{
			OK:    lipgloss.Color("#008000"), // Green
			Warn:  lipgloss.Color("#FFFF00"), // Yellow
			Error: lipgloss.Color("#FF0000"), // Red
			Info:  lipgloss.Color("#1E90FF"), // DodgerBlue
		},
		PaintBackground: true,
	}
	t.populateStyles()
	return t
}

// NoPaintBackground returns a theme with the same palette as [Default]
// but PaintBackground=false — letting the terminal's default background
// show through unstyled gaps. Use this when a table needs to override
// per-cell foregrounds without a background fight.
func NoPaintBackground() Theme {
	t := Default()
	t.PaintBackground = false
	t.populateStyles()
	return t
}

// Rebuild recomputes every pre-built style (InfoLabel, LogoStyle,
// Title, TableBorder, …) from the theme's current color fields and its
// current PaintBackground setting. Call it after changing any color on
// a theme returned by [Default] or [NoPaintBackground]; without it the
// pre-built styles keep the colors they were constructed with.
//
// Rebuild preserves the theme's background mode: a theme from
// [NoPaintBackground] stays unpainted, a theme from [Default] stays
// painted. Styles assigned by hand are overwritten, so override
// individual styles after calling Rebuild, not before.
func (t *Theme) Rebuild() {
	t.populateStyles()
}

// populateStyles fills in the pre-built styles based on the color
// fields. Constructors call this after assigning colors; theme variants
// can override individual styles afterward.
func (t *Theme) populateStyles() {
	t.InfoLabel = t.On(t.Label)
	t.InfoValue = t.On(t.Value).Bold(true)
	t.ShortcutKey = t.On(t.Border).Bold(true)
	t.ShortcutView = t.On(t.AccentAlt).Bold(true)
	t.ShortcutDesc = t.On(t.Muted)
	t.LogoStyle = t.On(t.Logo).Bold(true)
	t.Title = t.On(t.Accent).Bold(true)
	t.Error = t.On(t.Status.Error).Bold(true)
	t.Success = t.On(t.Status.OK).Bold(true)
	t.MutedStyle = t.On(t.Muted)
	t.Footer = t.On(t.Muted)
	t.FilterStyle = t.On(t.Filter).Bold(true)
	t.MarkStyle = t.On(t.Mark)
	t.PromptStyle = lipgloss.NewStyle().
		Background(t.Prompt).
		Foreground(t.PromptText).
		Bold(true).
		Padding(0, 1)
	t.TableBorder = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(t.FocusBorder())
	if t.PaintBackground {
		t.TableBorder = t.TableBorder.
			BorderBackground(t.Bg).
			Background(t.Bg)
	}
}
