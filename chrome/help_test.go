package chrome

import (
	"strings"
	"testing"
)

// TestNavigationHelpListsVimKeysOnly pins that the column advertises the
// vim spellings and never the arrows they translate to: the arrows work
// regardless, and listing both doubles the column.
func TestNavigationHelpListsVimKeysOnly(t *testing.T) {
	nav := NavigationHelp()
	keys := map[string]string{}
	for _, e := range nav.Entries {
		keys[e.Key] = e.Desc
		for _, arrow := range []string{"↑", "↓", "←", "→", helpKey("up"), helpKey("down"), helpKey("left"), helpKey("right"), "pgup", "pgdown"} {
			if strings.Contains(strings.ToLower(e.Key), arrow) {
				t.Errorf("navigation help lists an arrow/page key: %q", e.Key)
			}
		}
	}
	for k, desc := range map[string]string{
		helpKey("j"): "Down", helpKey("k"): "Up", helpKey("h"): "Left", helpKey("l"): "Right",
		helpKey("g"): "Goto Top", helpKey("shift-g"): "Goto Bottom", helpKey("ctrl-f"): "Page Down", helpKey("ctrl-b"): "Page Up",
	} {
		if keys[k] != desc {
			t.Errorf("navigation help %s = %q, want %q", k, keys[k], desc)
		}
	}
}

// TestHelpMatchesTheKeyConstants keeps the advertised keys and the
// constants apps bind in step: a constant renamed without its help entry
// would advertise a key nothing handles.
func TestHelpMatchesTheKeyConstants(t *testing.T) {
	all := map[string]bool{}
	for _, s := range []HelpSection{GeneralHelp(), NavigationHelp()} {
		if s.Title == "" || len(s.Entries) == 0 {
			t.Errorf("empty help section %+v", s)
		}
		for _, e := range s.Entries {
			all[e.Key] = true
		}
	}
	for constant, shown := range map[string]string{
		KeyBack: helpKey("q"), KeyReload: helpKey("ctrl-r"), KeyToggleCrumbs: helpKey("ctrl-g"), KeyToggleHeader: helpKey("ctrl-e"),
		KeyHistoryBack: helpKey("["), KeyHistoryForward: helpKey("]"), KeyLastView: helpKey("-"),
		KeyFieldNext: helpKey("tab"), KeyFieldPrev: helpKey("backtab"),
	} {
		if !all[shown] {
			t.Errorf("constant %q is not advertised as %s", constant, shown)
		}
		if want := strings.NewReplacer("ctrl+", "ctrl-", "shift+tab", "backtab").Replace(constant); "<"+want+">" != shown {
			t.Errorf("constant %q and help key %s disagree", constant, shown)
		}
	}
}
