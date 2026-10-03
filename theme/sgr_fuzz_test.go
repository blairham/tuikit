// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package theme

import (
	"strings"
	"testing"
)

// FuzzReassertBackground feeds it arbitrary styled text, as a log line from
// a container would be. It may only ever insert the background sequence —
// every other byte passes through untouched — and a second pass must change
// nothing, since a reset already followed by the background is left alone.
func FuzzReassertBackground(f *testing.F) {
	for _, s := range []string{
		"", "plain", "\x1b[m", "a\x1b[0mb", "a\x1b[32;0mb\n", "\x1b[38;2;0;0;0mx\x1b[mY",
		"\x1b[38:2::1:2:3;49mz", "\x1b[", "\x1b[1", "\x1b[mq\r\n", "\x1b[0K\x1b[m!",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, in string) {
		out := ReassertBackground(in, testBg)
		if strings.ReplaceAll(out, testBg, "") != strings.ReplaceAll(in, testBg, "") {
			t.Fatalf("ReassertBackground(%q) = %q: changed more than inserting the background", in, out)
		}
		if again := ReassertBackground(out, testBg); again != out {
			t.Fatalf("not idempotent on %q:\n once  %q\n twice %q", in, out, again)
		}
	})
}
