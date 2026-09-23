// Command commandbar demonstrates [chrome.CommandBar] inside a minimal
// tuikit frame. Press `:` to open the bar, type a command, Enter to
// dispatch, Esc to cancel. Try `:q!` or `:quit` to exit, `:hello`, or
// anything else to see an "unknown command" error.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/blairham/tuikit/chrome"
	"github.com/blairham/tuikit/theme"
)

type model struct {
	bar    *chrome.CommandBar
	last   string
	chrome chrome.Chrome
	w, h   int
}

func newModel() model {
	t := theme.Default()
	return model{
		chrome: chrome.New(chrome.Config{Theme: t}),
		bar: chrome.NewCommandBar(t, chrome.CommandBarOpts{
			Placeholder: "try :hello or :q!",
			Suggestions: []string{"hello", "q!", "quit", "clear"},
		}),
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		return m, nil
	case tea.KeyPressMsg:
		if m.bar.Active() {
			handled, cmd := m.bar.Update(msg, m.dispatch)
			if handled {
				return m, cmd
			}
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case ":":
			return m, m.bar.Open()
		}
	}
	if m.bar.Active() {
		_, cmd := m.bar.Update(msg, nil)
		return m, cmd
	}
	return m, nil
}

func (m *model) dispatch(value string) (string, tea.Cmd) {
	switch strings.ToLower(value) {
	case "q!", "quit", "exit":
		return "", tea.Quit
	case "clear":
		m.last = ""
		return "", nil
	case "hello":
		m.last = "hello, world"
		return "", nil
	}
	return "unknown command: " + value, nil
}

func (m model) View() tea.View {
	if m.w == 0 || m.h == 0 {
		return tea.View{}
	}
	f := chrome.Frame{
		Width:      m.w,
		Height:     m.h,
		InfoLines:  []string{"CommandBar example", "Press : to open"},
		Shortcuts:  []string{m.chrome.ShortcutPair("<:>", "Command", "<q>", "Quit")},
		Content:    m.chrome.BorderedContent(centeredBody(m.last, m.w), m.w-2, m.h-12),
		Breadcrumb: []chrome.Crumb{{Label: "example", Leaf: true}},
	}
	if m.bar.Active() {
		f.Command = m.bar.Input()
	}
	if err := m.bar.Error(); err != "" {
		f.StatusBar = err
	}
	return tea.NewView(m.chrome.Render(f))
}

func centeredBody(line string, w int) string {
	if line == "" {
		line = "(type :hello to set the message, :clear to reset)"
	}
	pad := max((w-len(line))/2, 0)
	return strings.Repeat(" ", pad) + line
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
