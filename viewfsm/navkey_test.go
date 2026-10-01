package viewfsm

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// TestTranslateNavKey pins every vim spelling onto its arrow/page key, and
// that the keys it maps onto — and everything else — pass through as-is,
// so both spellings keep working.
func TestTranslateNavKey(t *testing.T) {
	press := func(s string) tea.KeyMsg {
		if len(s) == 1 {
			return tea.KeyPressMsg{Code: rune(s[0]), Text: s}
		}
		// "ctrl+x" → ctrl-modified x.
		return tea.KeyPressMsg{Code: rune(s[len(s)-1]), Mod: tea.ModCtrl}
	}
	for _, tc := range []struct {
		in   string
		want rune
	}{
		{in: "j", want: tea.KeyDown},
		{in: "k", want: tea.KeyUp},
		{in: "h", want: tea.KeyLeft},
		{in: "l", want: tea.KeyRight},
		{in: "g", want: tea.KeyHome},
		{in: "G", want: tea.KeyEnd},
		{in: navPageDownK9s, want: tea.KeyPgDown},
		{in: navPageUpK9s, want: tea.KeyPgUp},
		{in: navPageDownVim, want: tea.KeyPgDown},
		{in: navPageUpVim, want: tea.KeyPgUp},
	} {
		got, ok := TranslateNavKey(press(tc.in)).(tea.KeyPressMsg)
		if !ok || got.Code != tc.want || got.Mod != 0 {
			t.Errorf("TranslateNavKey(%s) = %+v, want code %v", tc.in, got, tc.want)
		}
	}

	for _, k := range []tea.KeyPressMsg{
		{Code: tea.KeyDown},
		{Code: tea.KeyUp},
		{Code: tea.KeyLeft},
		{Code: tea.KeyRight},
		{Code: tea.KeyPgDown},
		{Code: tea.KeyPgUp},
		{Code: tea.KeyHome},
		{Code: tea.KeyEnd},
		{Code: 'x', Text: "x"},
		{Code: 'L', Text: "L"},
	} {
		if got := TranslateNavKey(k); got != tea.Msg(k) {
			t.Errorf("TranslateNavKey(%s) rewrote a key it does not own: %+v", k.String(), got)
		}
	}
}
