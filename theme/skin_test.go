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
		"dodgerblue": "#1e90ff",
		"DodgerBlue": "#1e90ff",
		"#87cefa":    "#87cefa",
		"#abc":       "#aabbcc",
		"default":    "default",
		"":           "<nil>",
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
    xray: {fgColor: "#ffffff"}
`

func TestWithSkin(t *testing.T) {
	var file struct {
		K9s Skin `yaml:"k9s"`
	}
	if err := yaml.Unmarshal([]byte(skinYAML), &file); err != nil {
		t.Fatal(err)
	}
	got, err := Default().WithSkin(file.K9s)
	if err != nil {
		t.Fatal(err)
	}
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

func TestWithSkinNamesABadColor(t *testing.T) {
	var s Skin
	s.Frame.Menu.KeyColor = "bluish"
	if _, err := Default().WithSkin(s); err == nil || !strings.Contains(err.Error(), "k9s.frame.menu.keyColor") {
		t.Errorf("err = %v, want it to name k9s.frame.menu.keyColor", err)
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
		hex(d.LogTextColor()) != hex(d.Selection) {
		t.Error("an unset optional color does not fall back to the color used before it")
	}
}
