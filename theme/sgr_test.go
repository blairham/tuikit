package theme

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

const testBg = "\x1b[48;2;0;0;0m"

func TestReassertBackground(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, in, want string
	}{
		{
			name: "combined reset and color (systemd OK)",
			in:   "[\x1b[0;32m  OK  \x1b[0m] Started",
			want: "[\x1b[0;32m" + testBg + "  OK  \x1b[0m" + testBg + "] Started",
		},
		{name: "bare reset", in: "a\x1b[mb", want: "a\x1b[m" + testBg + "b"},
		{name: "zero reset", in: "a\x1b[0mb", want: "a\x1b[0m" + testBg + "b"},
		{name: "49 alone", in: "a\x1b[49mb", want: "a\x1b[49m" + testBg + "b"},
		{name: "49 mid list", in: "a\x1b[1;49;31mb", want: "a\x1b[1;49;31m" + testBg + "b"},
		{name: "empty param mid list", in: "a\x1b[1;;31mb", want: "a\x1b[1;;31m" + testBg + "b"},
		{name: "truecolor black fg is not a reset", in: "a\x1b[38;2;0;0;0mb", want: "a\x1b[38;2;0;0;0mb"},
		{name: "256 black bg is not a reset", in: "a\x1b[48;5;0mb", want: "a\x1b[48;5;0mb"},
		{name: "underline color zeros", in: "a\x1b[58;2;0;0;0mb", want: "a\x1b[58;2;0;0;0mb"},
		{name: "colon truecolor fg", in: "a\x1b[38:2::0:0:0mb", want: "a\x1b[38:2::0:0:0mb"},
		{name: "colon form then reset", in: "a\x1b[38:2::0:0:0;0mb", want: "a\x1b[38:2::0:0:0;0m" + testBg + "b"},
		{name: "underline off subparam", in: "a\x1b[4:0mb", want: "a\x1b[4:0mb"},
		{name: "reset then new bg", in: "a\x1b[0;48;5;1mb", want: "a\x1b[0;48;5;1mb"},
		{name: "fg only", in: "a\x1b[32mb", want: "a\x1b[32mb"},
		{name: "non-SGR CSI untouched", in: "a\x1b[0Kb\x1b[2J\x1b[0;0H", want: "a\x1b[0Kb\x1b[2J\x1b[0;0H"},
		{name: "private marker untouched", in: "a\x1b[?0mb", want: "a\x1b[?0mb"},
		{name: "reset at end of line", in: "a\x1b[m\nb\x1b[0m", want: "a\x1b[m\nb\x1b[0m"},
		{name: "reset at CRLF", in: "a\x1b[m\r\nb", want: "a\x1b[m\r\nb"},
		{name: "already followed by bg", in: "a\x1b[m" + testBg + "b", want: "a\x1b[m" + testBg + "b"},
		{name: "unterminated", in: "a\x1b[0", want: "a\x1b[0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := ReassertBackground(tc.in, testBg); got != tc.want {
				t.Errorf("ReassertBackground(%q)\n got %q\nwant %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestReassertBackground_EmptySeqIsNoop(t *testing.T) {
	t.Parallel()
	in := "a\x1b[0;32mb\x1b[m c"
	if got := ReassertBackground(in, ""); got != in {
		t.Errorf("got %q; want input unchanged", got)
	}
}

func TestBackgroundSeq(t *testing.T) {
	t.Parallel()
	got := BackgroundSeq(lipgloss.Color("#000000"))
	if !strings.HasPrefix(got, "\x1b[") || !strings.HasSuffix(got, "m") || !strings.Contains(got, "48;") {
		t.Errorf("BackgroundSeq(black) = %q; want a background SGR", got)
	}
}
