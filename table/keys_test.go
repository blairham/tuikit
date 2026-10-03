// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package table

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// The key constants are compared against KeyPressMsg.String(), so each must
// be the string bubbletea actually produces for its key.
func TestKeyConstantsMatchBubbletea(t *testing.T) {
	t.Parallel()
	cases := map[string]tea.KeyPressMsg{
		KeySortPrev:     {Code: tea.KeyLeft, Mod: tea.ModShift},
		KeySortNext:     {Code: tea.KeyRight, Mod: tea.ModShift},
		KeyMarkToggle:   {Code: tea.KeySpace},
		KeyMarkRange:    {Code: tea.KeySpace, Mod: tea.ModCtrl},
		KeyMarkClear:    {Code: '\\', Mod: tea.ModCtrl},
		keyMarkRangeNUL: {Code: '@', Mod: tea.ModCtrl},
	}
	for want, msg := range cases {
		if got := msg.String(); got != want {
			t.Errorf("constant %q, but bubbletea says %q", want, got)
		}
	}
}
