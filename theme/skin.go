// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package theme

import (
	"errors"
	"fmt"
	"image/color"
	"reflect"
	"slices"

	"charm.land/lipgloss/v2"
)

// Skin is k9s's skin schema — the part under the file's top-level "k9s:"
// key — as Go structs with yaml tags, field for field with k9s's own, so
// an app unmarshals a k9s skin file straight into one, strict decoding
// included, and every stock k9s skin works unchanged. tuikit reads no
// files and parses no YAML; the app does both. Colors are written as
// [ParseColor] reads them.
//
// [Theme.WithSkin] applies the colors tuikit draws with. A key with no
// counterpart in tuikit — body.fgColor, the dialog, the yaml, picker and
// log indicator colors, the dial and per-resource chart colors, the
// status colors other than error, pending and add — is decoded and
// ignored.
type Skin struct {
	Body   SkinBody   `yaml:"body"`
	Prompt SkinPrompt `yaml:"prompt"`
	Help   SkinHelp   `yaml:"help"`
	Frame  SkinFrame  `yaml:"frame"`
	Info   SkinInfo   `yaml:"info"`
	Views  SkinViews  `yaml:"views"`
	Dialog SkinDialog `yaml:"dialog"`
}

// SkinBody is k9s.body. The LogoColor variants are the logo's colors
// while k9s flashes a message, an info, a warning or an error.
type SkinBody struct {
	FgColor        string `yaml:"fgColor"`
	BgColor        string `yaml:"bgColor"`
	LogoColor      string `yaml:"logoColor"`
	LogoColorMsg   string `yaml:"logoColorMsg"`
	LogoColorInfo  string `yaml:"logoColorInfo"`
	LogoColorWarn  string `yaml:"logoColorWarn"`
	LogoColorError string `yaml:"logoColorError"`
}

// SkinPrompt is k9s.prompt, the command and filter bars.
type SkinPrompt struct {
	FgColor      string           `yaml:"fgColor"`
	BgColor      string           `yaml:"bgColor"`
	SuggestColor string           `yaml:"suggestColor"`
	Border       SkinPromptBorder `yaml:"border"`
}

// SkinPromptBorder is k9s.prompt.border.
type SkinPromptBorder struct {
	Command string `yaml:"command"`
	Default string `yaml:"default"`
}

// SkinHelp is k9s.help.
type SkinHelp struct {
	FgColor      string `yaml:"fgColor"`
	BgColor      string `yaml:"bgColor"`
	SectionColor string `yaml:"sectionColor"`
	KeyColor     string `yaml:"keyColor"`
	NumKeyColor  string `yaml:"numKeyColor"`
}

// SkinDialog is k9s.dialog, the confirm and input dialogs.
type SkinDialog struct {
	FgColor            string `yaml:"fgColor"`
	BgColor            string `yaml:"bgColor"`
	ButtonFgColor      string `yaml:"buttonFgColor"`
	ButtonBgColor      string `yaml:"buttonBgColor"`
	ButtonFocusFgColor string `yaml:"buttonFocusFgColor"`
	ButtonFocusBgColor string `yaml:"buttonFocusBgColor"`
	LabelFgColor       string `yaml:"labelFgColor"`
	FieldFgColor       string `yaml:"fieldFgColor"`
}

// SkinFrame is k9s.frame.
type SkinFrame struct {
	Title  SkinTitle  `yaml:"title"`
	Border SkinBorder `yaml:"border"`
	Menu   SkinMenu   `yaml:"menu"`
	Crumbs SkinCrumbs `yaml:"crumbs"`
	Status SkinStatus `yaml:"status"`
}

// SkinTitle is k9s.frame.title, the content box's border title.
type SkinTitle struct {
	FgColor        string `yaml:"fgColor"`
	BgColor        string `yaml:"bgColor"`
	HighlightColor string `yaml:"highlightColor"`
	CounterColor   string `yaml:"counterColor"`
	FilterColor    string `yaml:"filterColor"`
}

// SkinBorder is k9s.frame.border.
type SkinBorder struct {
	FgColor    string `yaml:"fgColor"`
	FocusColor string `yaml:"focusColor"`
}

// SkinMenu is k9s.frame.menu, the header's shortcuts. FgStyle is a text
// style, not a color: "normal", "bold" or "dim".
type SkinMenu struct {
	FgColor     string        `yaml:"fgColor"`
	FgStyle     SkinTextStyle `yaml:"fgStyle"`
	KeyColor    string        `yaml:"keyColor"`
	NumKeyColor string        `yaml:"numKeyColor"`
}

// SkinTextStyle is a k9s text style: "normal", "bold" or "dim".
type SkinTextStyle string

// SkinCrumbs is k9s.frame.crumbs.
type SkinCrumbs struct {
	FgColor     string `yaml:"fgColor"`
	BgColor     string `yaml:"bgColor"`
	ActiveColor string `yaml:"activeColor"`
}

// SkinStatus is k9s.frame.status.
type SkinStatus struct {
	NewColor       string `yaml:"newColor"`
	ModifyColor    string `yaml:"modifyColor"`
	AddColor       string `yaml:"addColor"`
	PendingColor   string `yaml:"pendingColor"`
	ErrorColor     string `yaml:"errorColor"`
	HighlightColor string `yaml:"highlightColor"`
	KillColor      string `yaml:"killColor"`
	CompletedColor string `yaml:"completedColor"`
}

// SkinInfo is k9s.info, the header's info panel.
type SkinInfo struct {
	FgColor      string `yaml:"fgColor"`
	SectionColor string `yaml:"sectionColor"`
	CPUColor     string `yaml:"cpuColor"`
	MEMColor     string `yaml:"memColor"`
	K9sRevColor  string `yaml:"k9sRevColor"`
}

// SkinViews is k9s.views.
type SkinViews struct {
	Table  SkinTable  `yaml:"table"`
	Xray   SkinXray   `yaml:"xray"`
	Charts SkinCharts `yaml:"charts"`
	Yaml   SkinYaml   `yaml:"yaml"`
	Picker SkinPicker `yaml:"picker"`
	Logs   SkinLogs   `yaml:"logs"`
}

// SkinCharts is k9s.views.charts. DefaultChartColors' first two entries
// become ChartPrimary and ChartSecondary; the dial, background, focus and
// per-resource colors have no counterpart yet.
type SkinCharts struct {
	BgColor            string              `yaml:"bgColor"`
	DialBgColor        string              `yaml:"dialBgColor"`
	ChartBgColor       string              `yaml:"chartBgColor"`
	DefaultDialColors  []string            `yaml:"defaultDialColors"`
	DefaultChartColors []string            `yaml:"defaultChartColors"`
	ResourceColors     map[string][]string `yaml:"resourceColors"`
	FocusFgColor       string              `yaml:"focusFgColor"`
	FocusBgColor       string              `yaml:"focusBgColor"`
}

// SkinTable is k9s.views.table. CursorColor is the older name of
// CursorBgColor; a skin that sets both gets CursorBgColor.
type SkinTable struct {
	FgColor       string          `yaml:"fgColor"`
	BgColor       string          `yaml:"bgColor"`
	CursorFgColor string          `yaml:"cursorFgColor"`
	CursorBgColor string          `yaml:"cursorBgColor"`
	CursorColor   string          `yaml:"cursorColor"`
	MarkColor     string          `yaml:"markColor"`
	Header        SkinTableHeader `yaml:"header"`
}

// SkinTableHeader is k9s.views.table.header.
type SkinTableHeader struct {
	FgColor                 string `yaml:"fgColor"`
	BgColor                 string `yaml:"bgColor"`
	SorterColor             string `yaml:"sorterColor"`
	SelectedSortColumnColor string `yaml:"selectedSortColumnColor"`
}

// SkinXray is k9s.views.xray, which the tree package draws.
type SkinXray struct {
	FgColor         string `yaml:"fgColor"`
	BgColor         string `yaml:"bgColor"`
	CursorColor     string `yaml:"cursorColor"`
	CursorTextColor string `yaml:"cursorTextColor"`
	GraphicColor    string `yaml:"graphicColor"`
}

// SkinYaml is k9s.views.yaml, the resource describe and yaml views.
type SkinYaml struct {
	KeyColor   string `yaml:"keyColor"`
	ValueColor string `yaml:"valueColor"`
	ColonColor string `yaml:"colonColor"`
}

// SkinPicker is k9s.views.picker, the container picker.
type SkinPicker struct {
	MainColor     string `yaml:"mainColor"`
	FocusColor    string `yaml:"focusColor"`
	ShortcutColor string `yaml:"shortcutColor"`
}

// SkinLogs is k9s.views.logs.
type SkinLogs struct {
	FgColor   string           `yaml:"fgColor"`
	BgColor   string           `yaml:"bgColor"`
	Indicator SkinLogIndicator `yaml:"indicator"`
}

// SkinLogIndicator is k9s.views.logs.indicator, the log view's toggles.
type SkinLogIndicator struct {
	FgColor        string `yaml:"fgColor"`
	BgColor        string `yaml:"bgColor"`
	ToggleOnColor  string `yaml:"toggleOnColor"`
	ToggleOffColor string `yaml:"toggleOffColor"`
}

// Check reports every color in the skin that [ParseColor] cannot read,
// each named by its key, as "k9s.frame.menu.keyColor", joined into one
// error; nil when every color reads. [Theme.WithSkin] draws such a color
// in the terminal's own, as k9s does — k9s's stock skin itself names
// one, "linegreen" — so Check is for an app that wants to warn about a
// skin, not a gate a skin must pass.
func (s Skin) Check() error {
	var errs []error
	walkSkin(reflect.ValueOf(s), "k9s", func(key, value string) {
		if _, err := ParseColor(value); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", key, err))
		}
	})
	return errors.Join(errs...)
}

// walkSkin calls f with the key and value of every color under v, in
// field order: struct fields by yaml tag, list entries by index, and
// resource colors by sorted resource name.
func walkSkin(v reflect.Value, key string, f func(key, value string)) {
	switch v.Kind() {
	case reflect.String:
		if v.Type() == reflect.TypeFor[string]() {
			f(key, v.String())
		}
	case reflect.Struct:
		for i := range v.NumField() {
			walkSkin(v.Field(i), key+"."+v.Type().Field(i).Tag.Get("yaml"), f)
		}
	case reflect.Slice:
		for i := range v.Len() {
			walkSkin(v.Index(i), fmt.Sprintf("%s[%d]", key, i), f)
		}
	case reflect.Map:
		names := make([]string, 0, v.Len())
		for _, k := range v.MapKeys() {
			names = append(names, k.String())
		}
		slices.Sort(names)
		for _, name := range names {
			walkSkin(v.MapIndex(reflect.ValueOf(name)), key+"."+name, f)
		}
	default:
	}
}

// WithSkin returns the theme with every color the skin sets, and its
// styles rebuilt; colors the skin leaves out keep the theme's. A body
// bgColor of "default" stops the theme painting its background, so the
// terminal's own shows through. A color [ParseColor] cannot read is drawn
// in the terminal's own color, as k9s draws it; [Skin.Check] names them.
func (t Theme) WithSkin(s Skin) Theme {
	set := func(dst *color.Color, value string) {
		c, err := ParseColor(value)
		if err != nil {
			c = lipgloss.NoColor{}
		}
		if c != nil {
			*dst = c
		}
	}

	var bg color.Color
	set(&bg, s.Body.BgColor)
	switch {
	case bg == nil:
	case isDefault(bg):
		t.PaintBackground = false
	default:
		t.Bg = bg
	}
	// Later entries win: the title's filter color over the prompt's
	// border, cursorBgColor over its older name cursorColor.
	for _, e := range []struct {
		dst   *color.Color
		value string
	}{
		{&t.Logo, s.Body.LogoColor},
		{&t.Label, s.Info.FgColor},
		{&t.Value, s.Info.SectionColor},
		{&t.Border, s.Frame.Border.FgColor},
		{&t.BorderFocus, s.Frame.Border.FocusColor},
		{&t.MenuText, s.Frame.Menu.FgColor},
		{&t.MenuKey, s.Frame.Menu.KeyColor},
		{&t.MenuNumKey, s.Frame.Menu.NumKeyColor},
		{&t.BreadcrumbFg, s.Frame.Crumbs.FgColor},
		{&t.BreadcrumbBg, s.Frame.Crumbs.BgColor},
		{&t.BreadcrumbActive, s.Frame.Crumbs.ActiveColor},
		{&t.Status.Error, s.Frame.Status.ErrorColor},
		{&t.Status.Warn, s.Frame.Status.PendingColor},
		{&t.Status.Info, s.Frame.Status.AddColor},
		{&t.InputText, s.Prompt.FgColor},
		{&t.Suggestion, s.Prompt.SuggestColor},
		{&t.CommandBorder, s.Prompt.Border.Command},
		{&t.Filter, s.Prompt.Border.Default},
		{&t.Accent, s.Frame.Title.FgColor},
		{&t.AccentAlt, s.Frame.Title.HighlightColor},
		{&t.AccentBold, s.Frame.Title.CounterColor},
		{&t.Filter, s.Frame.Title.FilterColor},
		{&t.HelpDesc, s.Help.FgColor},
		{&t.HelpSection, s.Help.SectionColor},
		{&t.Selection, s.Views.Table.CursorColor},
		{&t.Selection, s.Views.Table.CursorBgColor},
		{&t.TableText, s.Views.Table.FgColor},
		{&t.SelectionText, s.Views.Table.CursorFgColor},
		{&t.Mark, s.Views.Table.MarkColor},
		{&t.TableHeader, s.Views.Table.Header.FgColor},
		{&t.XrayText, s.Views.Xray.FgColor},
		{&t.XrayCursor, s.Views.Xray.CursorColor},
		{&t.XrayCursorText, s.Views.Xray.CursorTextColor},
		{&t.XrayGraphic, s.Views.Xray.GraphicColor},
		{&t.LogText, s.Views.Logs.FgColor},
	} {
		set(e.dst, e.value)
	}
	for i, dst := range []*color.Color{&t.ChartPrimary, &t.ChartSecondary} {
		if i < len(s.Views.Charts.DefaultChartColors) {
			set(dst, s.Views.Charts.DefaultChartColors[i])
		}
	}

	t.populateStyles()
	return t
}
