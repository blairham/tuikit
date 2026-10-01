package chrome

// GeneralHelp is the help overlay's GENERAL column for the shortcuts every
// k9s-style app shares, in k9s's wording. Apps append their own entries:
//
//	g := chrome.GeneralHelp()
//	g.Entries = append(g.Entries, chrome.HelpEntry{Key: helpKey("ctrl-s"), Desc: "Save"})
//
// Listing a key here is a promise the app binds it; drop any entry an app
// does not support before rendering.
func GeneralHelp() HelpSection {
	return HelpSection{
		Title: "GENERAL",
		Entries: []HelpEntry{
			{Key: helpKey(":cmd"), Desc: "Command mode"},
			{Key: helpKey("/term"), Desc: "Filter mode"},
			{Key: helpKey("?"), Desc: "Help"},
			{Key: helpKey("esc"), Desc: "Back/Clear"},
			{Key: helpKey("q"), Desc: "Back"},
			{Key: helpKey("ctrl-u"), Desc: "Command Clear"},
			{Key: helpKey("tab"), Desc: "Field Next"},
			{Key: helpKey("backtab"), Desc: "Field Previous"},
			{Key: helpKey("ctrl-r"), Desc: "Reload"},
			{Key: helpKey("ctrl-g"), Desc: "Toggle Crumbs"},
			{Key: helpKey("ctrl-e"), Desc: "Toggle Header"},
			{Key: helpKey(":q"), Desc: "Quit"},
		},
	}
}

// NavigationHelp is the help overlay's NAVIGATION column. It lists the vim
// keys [viewfsm.TranslateNavKey] maps, not the arrow, page and home/end keys
// they map onto: those work regardless, and listing both spellings would
// double the column for no information.
func NavigationHelp() HelpSection {
	return HelpSection{
		Title: "NAVIGATION",
		Entries: []HelpEntry{
			{Key: helpKey("j"), Desc: "Down"},
			{Key: helpKey("k"), Desc: "Up"},
			{Key: helpKey("h"), Desc: "Left"},
			{Key: helpKey("l"), Desc: "Right"},
			{Key: helpKey("g"), Desc: "Goto Top"},
			{Key: helpKey("shift-g"), Desc: "Goto Bottom"},
			{Key: helpKey("ctrl-f"), Desc: "Page Down"},
			{Key: helpKey("ctrl-b"), Desc: "Page Up"},
			{Key: helpKey("["), Desc: "History Back"},
			{Key: helpKey("]"), Desc: "History Forward"},
			{Key: helpKey("-"), Desc: "Last Used Command"},
		},
	}
}

// helpKey is a key as the help overlay shows it: "<j>", "<ctrl-g>".
func helpKey(k string) string { return "<" + k + ">" }
