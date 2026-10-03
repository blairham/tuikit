// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package theme

import (
	"image/color"
	"reflect"

	"charm.land/lipgloss/v2"
	"github.com/lucasb-eyer/go-colorful"
)

// chromaKept is the share of a color's chroma an inversion keeps at the
// least: the lightness moves before the color fades to gray.
const chromaKept = 0.5

// InvertColor turns a dark color light and a light one dark while keeping
// its hue, as k9s's --invert does: in OkLch, lightness L becomes 1-L, and
// the chroma is kept — at least half of it, moving the lightness toward
// the middle when the sRGB gamut has no room for that much at 1-L, and
// all of it when the gamut allows. A gray stays gray. nil and the
// terminal's default color are returned unchanged.
func InvertColor(c color.Color) color.Color {
	if c == nil || isDefault(c) {
		return c
	}
	col, ok := colorful.MakeColor(c)
	if !ok {
		return c
	}
	l, ch, h := col.OkLch()
	if ch < 0.01 {
		return lipgloss.Color(colorful.OkLch(1-l, 0, h).Clamped().Hex())
	}
	targetL := closestLightness(1-l, ch*chromaKept, h)
	return lipgloss.Color(colorful.OkLch(targetL, min(ch, maxChroma(targetL, h)), h).Clamped().Hex())
}

// maxChroma is the largest chroma at lightness l and hue h that is still
// an sRGB color, to a thousandth.
func maxChroma(l, h float64) float64 {
	lo, hi := 0.0, 0.4
	for hi-lo > 0.001 {
		mid := (lo + hi) / 2
		if colorful.OkLch(l, mid, h).IsValid() {
			lo = mid
		} else {
			hi = mid
		}
	}
	return lo
}

// closestLightness is target, or the nearest lightness that can carry
// chroma c at hue h: searched toward the middle first, where the gamut is
// widest, then past it.
func closestLightness(target, c, h float64) float64 {
	if maxChroma(target, h) >= c {
		return target
	}
	if target < 0.5 {
		for l := target; l <= 0.5; l += 0.01 {
			if maxChroma(l, h) >= c {
				return l
			}
		}
		for l := 0.51; l <= 0.95; l += 0.01 {
			if maxChroma(l, h) >= c {
				return l
			}
		}
		return target
	}
	for l := target; l >= 0.5; l -= 0.01 {
		if maxChroma(l, h) >= c {
			return l
		}
	}
	for l := 0.49; l >= 0.05; l -= 0.01 {
		if maxChroma(l, h) >= c {
			return l
		}
	}
	return target
}

var colorType = reflect.TypeFor[color.Color]()

// Inverted returns the theme with every color inverted by [InvertColor] —
// k9s's --invert — and its styles rebuilt. Colors left unset stay unset.
func (t Theme) Inverted() Theme {
	invertFields(reflect.ValueOf(&t).Elem())
	t.populateStyles()
	return t
}

// invertFields inverts every color field of v, a Theme or its
// StatusColors. The pre-built styles are left to populateStyles.
func invertFields(v reflect.Value) {
	for i := range v.NumField() {
		f := v.Field(i)
		switch f.Type() {
		case colorType:
			if c, ok := f.Interface().(color.Color); ok && c != nil {
				f.Set(reflect.ValueOf(InvertColor(c)))
			}
		case reflect.TypeFor[StatusColors]():
			invertFields(f)
		}
	}
}
