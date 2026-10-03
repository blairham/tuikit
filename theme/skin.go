// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package theme

import (
	"fmt"
	"image/color"
)

// Skin is k9s's skin schema — the part under the file's top-level "k9s:"
// key — as Go structs with yaml tags, so an app unmarshals a k9s skin
// file straight into one and a stock k9s skin works unchanged. tuikit
// reads no files and parses no YAML; the app does both. Colors are
// written as [ParseColor] reads them.
//
// [Theme.WithSkin] applies the colors tuikit draws with. A k9s key with
// no counterpart in tuikit's chrome — body.fgColor, the dialog, xray and
// chart colors, the status colors other than error, pending and add — is
// accepted and ignored.
type Skin struct {
	Body   SkinBody   `yaml:"body"`
	Prompt SkinPrompt `yaml:"prompt"`
	Help   SkinHelp   `yaml:"help"`
	Frame  SkinFrame  `yaml:"frame"`
	Info   SkinInfo   `yaml:"info"`
	Views  SkinViews  `yaml:"views"`
}

// SkinBody is k9s.body.
type SkinBody struct {
	FgColor   string `yaml:"fgColor"`
	BgColor   string `yaml:"bgColor"`
	LogoColor string `yaml:"logoColor"`
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

// SkinMenu is k9s.frame.menu, the header's shortcuts.
type SkinMenu struct {
	FgColor     string `yaml:"fgColor"`
	KeyColor    string `yaml:"keyColor"`
	NumKeyColor string `yaml:"numKeyColor"`
}

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
}

// SkinViews is k9s.views.
type SkinViews struct {
	Table SkinTable `yaml:"table"`
	Logs  SkinLogs  `yaml:"logs"`
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
	FgColor     string `yaml:"fgColor"`
	BgColor     string `yaml:"bgColor"`
	SorterColor string `yaml:"sorterColor"`
}

// SkinLogs is k9s.views.logs.
type SkinLogs struct {
	FgColor string `yaml:"fgColor"`
	BgColor string `yaml:"bgColor"`
}

// WithSkin returns the theme with every color the skin sets, and its
// styles rebuilt; colors the skin leaves out keep the theme's. A body
// bgColor of "default" stops the theme painting its background, so the
// terminal's own shows through. An unparsable color is an error naming
// its key, as "k9s.frame.menu.keyColor".
func (t Theme) WithSkin(s Skin) (Theme, error) {
	var err error
	set := func(dst *color.Color, key, value string) {
		if err != nil {
			return
		}
		c, perr := ParseColor(value)
		if perr != nil {
			err = fmt.Errorf("k9s.%s: %w", key, perr)
			return
		}
		if c != nil {
			*dst = c
		}
	}

	var bg color.Color
	set(&bg, "body.bgColor", s.Body.BgColor)
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
		dst        *color.Color
		key, value string
	}{
		{&t.Logo, "body.logoColor", s.Body.LogoColor},
		{&t.Label, "info.fgColor", s.Info.FgColor},
		{&t.Value, "info.sectionColor", s.Info.SectionColor},
		{&t.Border, "frame.border.fgColor", s.Frame.Border.FgColor},
		{&t.BorderFocus, "frame.border.focusColor", s.Frame.Border.FocusColor},
		{&t.MenuText, "frame.menu.fgColor", s.Frame.Menu.FgColor},
		{&t.MenuKey, "frame.menu.keyColor", s.Frame.Menu.KeyColor},
		{&t.MenuNumKey, "frame.menu.numKeyColor", s.Frame.Menu.NumKeyColor},
		{&t.BreadcrumbFg, "frame.crumbs.fgColor", s.Frame.Crumbs.FgColor},
		{&t.BreadcrumbBg, "frame.crumbs.bgColor", s.Frame.Crumbs.BgColor},
		{&t.BreadcrumbActive, "frame.crumbs.activeColor", s.Frame.Crumbs.ActiveColor},
		{&t.Status.Error, "frame.status.errorColor", s.Frame.Status.ErrorColor},
		{&t.Status.Warn, "frame.status.pendingColor", s.Frame.Status.PendingColor},
		{&t.Status.Info, "frame.status.addColor", s.Frame.Status.AddColor},
		{&t.InputText, "prompt.fgColor", s.Prompt.FgColor},
		{&t.Suggestion, "prompt.suggestColor", s.Prompt.SuggestColor},
		{&t.CommandBorder, "prompt.border.command", s.Prompt.Border.Command},
		{&t.Filter, "prompt.border.default", s.Prompt.Border.Default},
		{&t.Accent, "frame.title.fgColor", s.Frame.Title.FgColor},
		{&t.AccentAlt, "frame.title.highlightColor", s.Frame.Title.HighlightColor},
		{&t.AccentBold, "frame.title.counterColor", s.Frame.Title.CounterColor},
		{&t.Filter, "frame.title.filterColor", s.Frame.Title.FilterColor},
		{&t.HelpDesc, "help.fgColor", s.Help.FgColor},
		{&t.HelpSection, "help.sectionColor", s.Help.SectionColor},
		{&t.Selection, "views.table.cursorColor", s.Views.Table.CursorColor},
		{&t.Selection, "views.table.cursorBgColor", s.Views.Table.CursorBgColor},
		{&t.TableText, "views.table.fgColor", s.Views.Table.FgColor},
		{&t.SelectionText, "views.table.cursorFgColor", s.Views.Table.CursorFgColor},
		{&t.Mark, "views.table.markColor", s.Views.Table.MarkColor},
		{&t.TableHeader, "views.table.header.fgColor", s.Views.Table.Header.FgColor},
		{&t.LogText, "views.logs.fgColor", s.Views.Logs.FgColor},
	} {
		set(e.dst, e.key, e.value)
	}

	if err != nil {
		return Theme{}, err
	}
	t.populateStyles()
	return t, nil
}
