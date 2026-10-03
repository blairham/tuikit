// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package theme

import (
	"image/color"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
)

// BackgroundSeq returns the raw SGR escape that sets c as the background,
// derived from lipgloss so it matches what the theme's own styles emit.
// It returns "" if lipgloss produced no escape for c.
func BackgroundSeq(c color.Color) string {
	sample := lipgloss.NewStyle().Background(c).Render(" ")
	if i := strings.IndexByte(sample, 'm'); strings.HasPrefix(sample, "\x1b[") && i > 0 {
		return sample[:i+1]
	}
	return ""
}

// ReassertBackground writes bgSeq immediately after every SGR sequence in s
// that leaves the background at the terminal default, so a painted canvas
// survives the resets that end inner styled spans.
//
// A sequence leaves the background at default when, read left to right, its
// last background-affecting parameter is a reset (an empty list, an empty or
// "0" parameter) or 49. Combined resets such as "\x1b[0;32m" count; the
// arguments of extended colors (38/48/58 followed by ;5;n or ;2;r;g;b) are
// skipped so the zeros of a black color are not read as resets. A
// colon-separated group ("38:2::r:g:b") is one parameter.
//
// A reset that ends a line (followed by "\n" or the end of s) is left alone so
// the background never bleeds past the line's right edge or onto the next
// line. A reset already followed by bgSeq is left alone too.
func ReassertBackground(s, bgSeq string) string {
	if bgSeq == "" || !strings.Contains(s, "\x1b[") {
		return s
	}
	var sb strings.Builder
	sb.Grow(len(s) + len(s)/8)
	i := 0
	for {
		j := strings.Index(s[i:], "\x1b[")
		if j < 0 {
			sb.WriteString(s[i:])
			return sb.String()
		}
		start := i + j
		end, final := csiEnd(s, start+2)
		if end < 0 {
			sb.WriteString(s[i:])
			return sb.String()
		}
		sb.WriteString(s[i:end])
		i = end
		if final != 'm' || !sgrLeavesDefaultBg(s[start+2:end-1]) {
			continue
		}
		rest := s[i:]
		if rest == "" || rest[0] == '\n' || (rest[0] == '\r' && strings.HasPrefix(rest, "\r\n")) ||
			strings.HasPrefix(rest, bgSeq) {
			continue
		}
		sb.WriteString(bgSeq)
	}
}

// csiEnd scans a CSI sequence whose parameter bytes start at p and returns
// the index just past its final byte, and the final byte. It returns -1 if
// the sequence is unterminated or malformed.
func csiEnd(s string, p int) (end int, final byte) {
	for k := p; k < len(s); k++ {
		c := s[k]
		switch {
		case c >= 0x20 && c <= 0x3f: // parameter and intermediate bytes
		case c >= 0x40 && c <= 0x7e:
			return k + 1, c
		default:
			return -1, 0
		}
	}
	return -1, 0
}

// sgrLeavesDefaultBg reports whether an SGR parameter string leaves the
// background at the terminal default.
func sgrLeavesDefaultBg(params string) bool {
	if params == "" {
		return true
	}
	if !isPlainSGR(params) {
		return false
	}
	fields := strings.Split(params, ";")
	defaultBg := false
	for k := 0; k < len(fields); k++ {
		code, colon, ok := sgrCode(fields[k])
		if !ok {
			continue
		}
		switch code {
		case 0, 49:
			defaultBg = true
		case 38, 48, 58:
			if code == 48 {
				defaultBg = false
			}
			if !colon {
				k += extendedColorArgs(fields[k+1:])
			}
		}
	}
	return defaultBg
}

// isPlainSGR reports whether params holds only digits and separators.
// Anything else (a private marker or an intermediate byte) is not a plain
// SGR and is left alone.
func isPlainSGR(params string) bool {
	for k := 0; k < len(params); k++ {
		if c := params[k]; (c < '0' || c > '9') && c != ';' && c != ':' {
			return false
		}
	}
	return true
}

// sgrCode returns the code of one ';'-separated field. A colon group
// ("38:2::r:g:b") carries its own arguments and its first subfield is the
// code; colon reports that form. An empty field is 0.
func sgrCode(field string) (code int, colon, ok bool) {
	head, _, colon := strings.Cut(field, ":")
	if head == "" {
		return 0, colon, true
	}
	n, err := strconv.Atoi(head)
	if err != nil {
		return 0, colon, false
	}
	return n, colon, true
}

// extendedColorArgs returns how many of the fields following a 38/48/58
// belong to it: "5;n" is two, "2;r;g;b" is four.
func extendedColorArgs(rest []string) int {
	if len(rest) == 0 {
		return 0
	}
	switch rest[0] {
	case "5":
		return 2
	case "2":
		return 4
	}
	return 0
}
