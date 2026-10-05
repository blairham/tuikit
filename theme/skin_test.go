// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package theme

import (
	"image/color"
	"math"
	"reflect"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/lucasb-eyer/go-colorful"
	"go.yaml.in/yaml/v3"
)

func hex(c color.Color) string {
	if c == nil {
		return "<nil>"
	}
	if isDefault(c) {
		return "default"
	}
	cc, _ := colorful.MakeColor(c) //nolint:errcheck // test colors are opaque
	return cc.Hex()
}

func TestParseColor(t *testing.T) {
	for in, want := range map[string]string{
		"dodgerblue":    "#1e90ff",
		"DodgerBlue":    "#1e90ff",
		"#87cefa":       "#87cefa",
		"#abc":          "#aabbcc",
		"default":       "default",
		"-":             "default",
		"rebeccapurple": "#663399",
		"":              "<nil>",
	} {
		c, err := ParseColor(in)
		if err != nil || hex(c) != want {
			t.Errorf("ParseColor(%q) = %s, %v; want %s", in, hex(c), err, want)
		}
	}
	for _, bad := range []string{"bluish", "#12345", "#ggg"} {
		if _, err := ParseColor(bad); err == nil {
			t.Errorf("ParseColor(%q) accepted it", bad)
		}
	}
}

func TestInvertColor(t *testing.T) {
	for in, want := range map[string]string{"#000000": "#ffffff", "#ffffff": "#000000"} {
		if got := hex(InvertColor(lipgloss.Color(in))); got != want {
			t.Errorf("InvertColor(%s) = %s, want %s", in, got, want)
		}
	}
	// A gray stays gray; a hue survives, lightness flipped.
	gray, _ := colorful.MakeColor(InvertColor(lipgloss.Color("#333333"))) //nolint:errcheck // opaque
	if _, c, _ := gray.OkLch(); c > 0.001 {
		t.Errorf("gray inverted to a color: %s", gray.Hex())
	}
	in, _ := colorful.Hex("#1e90ff")
	out, _ := colorful.MakeColor(InvertColor(lipgloss.Color("#1e90ff"))) //nolint:errcheck // opaque
	l1, c1, h1 := in.OkLch()
	l2, c2, h2 := out.OkLch()
	if math.Abs(h1-h2) > 2 || l2 >= l1 || c2 < c1*chromaKept-0.01 {
		t.Errorf("dodgerblue (L %.2f C %.2f h %.0f) inverted to %s (L %.2f C %.2f h %.0f)", l1, c1, h1, out.Hex(), l2, c2, h2)
	}
	for _, c := range []color.Color{nil, lipgloss.NoColor{}} {
		if InvertColor(c) != c {
			t.Errorf("InvertColor changed %v", c)
		}
	}
}

// TestInvertedCoversEveryColor: every color a theme sets is inverted —
// found by reflection, so a field added later cannot be missed — the
// unset ones stay unset, and the styles are rebuilt from the new colors.
func TestInvertedCoversEveryColor(t *testing.T) {
	d := Default()
	inv := d.Inverted()
	var walk func(a, b reflect.Value, path string)
	walk = func(a, b reflect.Value, path string) {
		for i := range a.NumField() {
			name := path + a.Type().Field(i).Name
			fa, fb := a.Field(i), b.Field(i)
			switch fa.Type() {
			case colorType:
				switch {
				case fa.IsNil() != fb.IsNil():
					t.Errorf("%s: set %v, inverted set %v", name, !fa.IsNil(), !fb.IsNil())
				case !fa.IsNil() && hex(fa.Interface().(color.Color)) == hex(fb.Interface().(color.Color)): //nolint:forcetypeassert // color field
					t.Errorf("%s: not inverted (%s)", name, hex(fa.Interface().(color.Color))) //nolint:forcetypeassert // color field
				}
			case reflect.TypeFor[StatusColors]():
				walk(fa, fb, name+".")
			}
		}
	}
	walk(reflect.ValueOf(d), reflect.ValueOf(inv), "")
	if hex(inv.Bg) != "#ffffff" || hex(inv.ShortcutKey.GetForeground()) != hex(InvertColor(d.Border)) {
		t.Errorf("Bg %s, ShortcutKey %s: styles not rebuilt", hex(inv.Bg), hex(inv.ShortcutKey.GetForeground()))
	}
}

// skinYAML is a skin as k9s writes one, decoded the way an app would, so
// the yaml tags are checked against k9s's key names, not only the Go.
const skinYAML = `
k9s:
  body:
    bgColor: default
    logoColor: "#bd93f9"
  info:
    fgColor: hotpink
    sectionColor: white
  frame:
    border: {fgColor: "#44475a", focusColor: "#6272a4"}
    menu: {fgColor: "#f8f8f2", keyColor: "#ff79c6", numKeyColor: "#bd93f9"}
    crumbs: {fgColor: black, bgColor: "#8be9fd", activeColor: "#ffb86c"}
    status: {errorColor: "#ff5555", pendingColor: "#ffb86c", addColor: "#50fa7b"}
    title: {fgColor: "#8be9fd", highlightColor: "#ff79c6", counterColor: "#f1fa8c", filterColor: "#50fa7b"}
  prompt:
    fgColor: "#f8f8f2"
    suggestColor: "#6272a4"
    border: {command: "#ff79c6", default: "#ffffff"}
  help: {fgColor: "#f8f8f2", sectionColor: "#50fa7b"}
  views:
    table:
      fgColor: "#f8f8f2"
      cursorColor: "#111111"
      cursorFgColor: "#f8f8f2"
      cursorBgColor: "#44475a"
      markColor: "#ffb86c"
      header: {fgColor: "#f1fa8c"}
    logs: {fgColor: "#f8f8f2"}
    xray: {fgColor: "#ffffff", cursorColor: "#ff79c6", cursorTextColor: "#282a36", graphicColor: "#6272a4"}
    charts:
      defaultChartColors: ["#50fa7b", "#ff5555"]
`

func TestWithSkin(t *testing.T) {
	var file struct {
		K9s Skin `yaml:"k9s"`
	}
	if err := yaml.Unmarshal([]byte(skinYAML), &file); err != nil {
		t.Fatal(err)
	}
	got := Default().WithSkin(file.K9s)
	for name, pair := range map[string][2]color.Color{
		"Logo":                    {got.Logo, lipgloss.Color("#bd93f9")},
		"Label":                   {got.Label, lipgloss.Color("#ff69b4")},
		"Value":                   {got.Value, lipgloss.Color("#ffffff")},
		"Border":                  {got.Border, lipgloss.Color("#44475a")},
		"BorderFocus":             {got.BorderFocus, lipgloss.Color("#6272a4")},
		"MenuText":                {got.MenuText, lipgloss.Color("#f8f8f2")},
		"MenuKey":                 {got.MenuKey, lipgloss.Color("#ff79c6")},
		"MenuNumKey":              {got.MenuNumKey, lipgloss.Color("#bd93f9")},
		"BreadcrumbBg":            {got.BreadcrumbBg, lipgloss.Color("#8be9fd")},
		"BreadcrumbActive":        {got.BreadcrumbActive, lipgloss.Color("#ffb86c")},
		"Status.Error":            {got.Status.Error, lipgloss.Color("#ff5555")},
		"Status.Warn":             {got.Status.Warn, lipgloss.Color("#ffb86c")},
		"Status.Info":             {got.Status.Info, lipgloss.Color("#50fa7b")},
		"Accent":                  {got.Accent, lipgloss.Color("#8be9fd")},
		"AccentAlt":               {got.AccentAlt, lipgloss.Color("#ff79c6")},
		"AccentBold":              {got.AccentBold, lipgloss.Color("#f1fa8c")},
		"Filter (title wins)":     {got.Filter, lipgloss.Color("#50fa7b")},
		"CommandBorder":           {got.CommandBorder, lipgloss.Color("#ff79c6")},
		"InputText":               {got.InputText, lipgloss.Color("#f8f8f2")},
		"Suggestion":              {got.Suggestion, lipgloss.Color("#6272a4")},
		"HelpSection":             {got.HelpSection, lipgloss.Color("#50fa7b")},
		"Selection (Bg wins)":     {got.Selection, lipgloss.Color("#44475a")},
		"TableText":               {got.TableText, lipgloss.Color("#f8f8f2")},
		"SelectionText":           {got.SelectionText, lipgloss.Color("#f8f8f2")},
		"TableHeader":             {got.TableHeader, lipgloss.Color("#f1fa8c")},
		"Mark":                    {got.Mark, lipgloss.Color("#ffb86c")},
		"LogText":                 {got.LogText, lipgloss.Color("#f8f8f2")},
		"XrayText":                {got.XrayText, lipgloss.Color("#ffffff")},
		"XrayCursor":              {got.XrayCursor, lipgloss.Color("#ff79c6")},
		"XrayCursorText":          {got.XrayCursorText, lipgloss.Color("#282a36")},
		"XrayGraphic":             {got.XrayGraphic, lipgloss.Color("#6272a4")},
		"ChartPrimary":            {got.ChartPrimary, lipgloss.Color("#50fa7b")},
		"ChartSecondary":          {got.ChartSecondary, lipgloss.Color("#ff5555")},
		"ShortcutKey (rebuilt)":   {got.ShortcutKey.GetForeground(), lipgloss.Color("#ff79c6")},
		"Muted (not in the skin)": {got.Muted, Default().Muted},
	} {
		if hex(pair[0]) != hex(pair[1]) {
			t.Errorf("%s = %s, want %s", name, hex(pair[0]), hex(pair[1]))
		}
	}
	if got.PaintBackground || hex(got.Bg) != hex(Default().Bg) {
		t.Errorf("bgColor default: PaintBackground %v Bg %s", got.PaintBackground, hex(got.Bg))
	}
}

// TestSkinBadColorIsTerminalDefault: a color ParseColor cannot read is
// drawn in the terminal's own color, as k9s draws it — k9s's stock skin
// names "linegreen" — and Check names every such key.
func TestSkinBadColorIsTerminalDefault(t *testing.T) {
	var s Skin
	s.Frame.Menu.KeyColor = "bluish"
	s.Views.Charts.DefaultChartColors = []string{"linegreen", "#ff0000"}
	s.Views.Charts.ResourceColors = map[string][]string{"batch/v1/jobs": {"#00ff00", "nope"}}
	got := Default().WithSkin(s)
	if hex(got.MenuKey) != "default" || hex(got.ChartPrimary) != "default" || hex(got.ChartSecondary) != "#ff0000" {
		t.Errorf(
			"MenuKey %s, ChartPrimary %s, ChartSecondary %s",
			hex(got.MenuKey),
			hex(got.ChartPrimary),
			hex(got.ChartSecondary),
		)
	}
	err := s.Check()
	for _, key := range []string{
		"k9s.frame.menu.keyColor",
		"k9s.views.charts.defaultChartColors[0]",
		"k9s.views.charts.resourceColors.batch/v1/jobs[1]",
	} {
		if err == nil || !strings.Contains(err.Error(), key+": ") {
			t.Errorf("Check() = %v, want it to name %s", err, key)
		}
	}
	if err != nil && strings.Count(err.Error(), "\n") != 2 {
		t.Errorf("Check() names more than the three bad colors: %v", err)
	}
	if err := (Skin{}).Check(); err != nil {
		t.Errorf("an empty skin: %v", err)
	}
}

// allKeysYAML sets every key in k9s's skin schema (internal/config/
// styles.go), each to a distinct valid color, in every color form k9s
// reads, with the YAML anchors k9s's stock skins are written with.
const allKeysYAML = `
foreground: &fg "#c0c0c0"
k9s:
  body: {fgColor: *fg, bgColor: "-", logoColor: orange, logoColorMsg: white, logoColorInfo: green, logoColorWarn: yellow, logoColorError: red}
  prompt:
    fgColor: cadetblue
    bgColor: black
    suggestColor: dodgerblue
    border: {command: aqua, default: seagreen}
  help: {fgColor: cadetblue, bgColor: black, sectionColor: green, keyColor: dodgerblue, numKeyColor: fuchsia}
  dialog: {fgColor: dodgerblue, bgColor: black, buttonFgColor: black, buttonBgColor: dodgerblue, buttonFocusFgColor: white, buttonFocusBgColor: fuchsia, labelFgColor: fuchsia, fieldFgColor: dodgerblue}
  frame:
    title: {fgColor: aqua, bgColor: black, highlightColor: fuchsia, counterColor: papayawhip, filterColor: seagreen}
    border: {fgColor: dodgerblue, focusColor: lightskyblue}
    menu: {fgColor: white, fgStyle: dim, keyColor: dodgerblue, numKeyColor: fuchsia}
    crumbs: {fgColor: black, bgColor: aqua, activeColor: orange}
    status: {newColor: lightskyblue, modifyColor: greenyellow, addColor: dodgerblue, pendingColor: darkorange, errorColor: orangered, highlightColor: aqua, killColor: mediumpurple, completedColor: gray}
  info: {fgColor: orange, sectionColor: white, cpuColor: rebeccapurple, memColor: "#abc", k9sRevColor: default}
  views:
    table:
      fgColor: aqua
      bgColor: black
      cursorFgColor: black
      cursorBgColor: aqua
      cursorColor: aqua
      markColor: palegreen
      header: {fgColor: white, bgColor: black, sorterColor: aqua, selectedSortColumnColor: orange}
    xray: {fgColor: blue, bgColor: black, cursorColor: aqua, cursorTextColor: black, graphicColor: darkgoldenrod}
    charts:
      bgColor: default
      dialBgColor: black
      chartBgColor: black
      defaultDialColors: [palegreen, orangered]
      defaultChartColors: [palegreen, orangered]
      resourceColors:
        batch/v1/jobs: [palegreen, orangered]
        v1/pods: [aqua, fuchsia]
      focusFgColor: white
      focusBgColor: black
    yaml: {keyColor: steelblue, valueColor: papayawhip, colonColor: white}
    picker: {mainColor: white, focusColor: aqua, shortcutColor: fuchsia}
    logs:
      fgColor: white
      bgColor: black
      indicator: {fgColor: white, bgColor: black, toggleOnColor: limegreen, toggleOffColor: gray}
`

// TestSkinDecodesEveryK9sKey: a skin using every key k9s reads decodes
// with unknown keys refused, so no k9s key is missing from Skin, and every
// color in it reads.
func TestSkinDecodesEveryK9sKey(t *testing.T) {
	var file struct {
		Foreground string `yaml:"foreground"`
		K9s        Skin   `yaml:"k9s"`
	}
	dec := yaml.NewDecoder(strings.NewReader(allKeysYAML))
	dec.KnownFields(true)
	if err := dec.Decode(&file); err != nil {
		t.Fatal(err)
	}
	s := file.K9s
	if s.Body.FgColor != "#c0c0c0" || s.Frame.Menu.FgStyle != "dim" ||
		s.Views.Logs.Indicator.ToggleOnColor != "limegreen" ||
		len(s.Views.Charts.ResourceColors["v1/pods"]) != 2 ||
		s.Dialog.ButtonFocusBgColor != "fuchsia" {
		t.Errorf("decoded %+v", s)
	}
	if err := s.Check(); err != nil {
		t.Error(err)
	}
	var n int
	walkSkin(reflect.ValueOf(s), "k9s", func(_, value string) {
		if value != "" {
			n++
		}
	})
	if n != 91 { // the colors in allKeysYAML; fgStyle is not one
		t.Errorf("walkSkin saw %d set colors, want 91", n)
	}
	if got := Default().WithSkin(s); got.PaintBackground {
		t.Error(`body bgColor "-" left the background painted`)
	}
}

// TestOptionalColorsFallBack: a theme that sets none of the optional
// colors draws exactly as before they existed.
func TestOptionalColorsFallBack(t *testing.T) {
	d := Default()
	if hex(d.ShortcutKey.GetForeground()) != hex(d.Border) ||
		hex(d.ShortcutView.GetForeground()) != hex(d.AccentAlt) ||
		hex(d.ShortcutDesc.GetForeground()) != hex(d.Muted) ||
		hex(d.TableTextColor()) != hex(d.Selection) ||
		hex(d.TableHeaderColor()) != hex(d.Value) ||
		hex(d.LogTextColor()) != hex(d.Selection) ||
		hex(d.XrayTextColor()) != hex(d.TableTextColor()) ||
		hex(d.XrayCursorColor()) != hex(d.Selection) ||
		hex(d.XrayCursorTextColor()) != hex(d.SelectionTextColor()) ||
		hex(d.XrayGraphicColor()) != hex(d.Muted) {
		t.Error("an unset optional color does not fall back to the color used before it")
	}
}

// TestSkinChartColorsPartial: a skin naming one chart color sets the first
// and leaves the second at its default; Check names a bad one by index.
func TestSkinChartColorsPartial(t *testing.T) {
	var s Skin
	s.Views.Charts.DefaultChartColors = []string{"#123456"}
	got := Default().WithSkin(s)
	if hex(got.ChartPrimary) != "#123456" || hex(got.ChartSecondary) != hex(Default().ChartSecondary) {
		t.Errorf("one color: primary %s secondary %s", hex(got.ChartPrimary), hex(got.ChartSecondary))
	}
	s.Views.Charts.DefaultChartColors = []string{"#123456", "notacolor"}
	if err := s.Check(); err == nil || !strings.Contains(err.Error(), "defaultChartColors[1]") {
		t.Errorf("a bad second color: %v", err)
	}
}
