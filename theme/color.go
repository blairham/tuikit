// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package theme

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
)

// defaultColor is how a skin names the terminal's own color.
const defaultColor = "default"

// ParseColor reads a skin color as k9s writes one: a CSS color name
// ("dodgerblue", any case), "#rrggbb" or "#rgb", or "default" (also
// "transparent") for the terminal's own color, which comes back as
// lipgloss.NoColor. An empty string is not set and returns nil.
func ParseColor(s string) (color.Color, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "":
		return nil, nil //nolint:nilnil // unset is not an error and has no color
	case defaultColor, "transparent":
		return lipgloss.NoColor{}, nil
	}
	if hex, ok := strings.CutPrefix(s, "#"); ok {
		if len(hex) == 3 {
			hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
		}
		if len(hex) != 6 {
			return nil, fmt.Errorf("color %q: want #rrggbb or #rgb", s)
		}
		if _, err := strconv.ParseUint(hex, 16, 32); err != nil {
			return nil, fmt.Errorf("color %q: not hexadecimal", s)
		}
		return lipgloss.Color("#" + strings.ToUpper(hex)), nil
	}
	if v, ok := colorNames[s]; ok {
		return lipgloss.Color(fmt.Sprintf("#%06X", v)), nil
	}
	return nil, fmt.Errorf("color %q: not a color name, #rrggbb or default", s)
}

// isDefault reports whether c is the terminal's own color.
func isDefault(c color.Color) bool {
	_, ok := c.(lipgloss.NoColor)
	return ok
}

// or is c, or fallback when c is unset.
func or(c, fallback color.Color) color.Color {
	if c == nil {
		return fallback
	}
	return c
}
