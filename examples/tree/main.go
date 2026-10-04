// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

// Command tree demonstrates [tree.Model]: a resource and everything under
// it, k9s xray style. Move with the arrows (or h/j/k/l), space to open or
// close a node, x to open or close everything, enter to show the selected
// node's path, q to quit.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/blairham/tuikit/theme"
	"github.com/blairham/tuikit/tree"
)

type model struct {
	tree   *tree.Model
	status string
	open   bool
}

func sample() []*tree.Node {
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	pod := func(id string) *tree.Node {
		return &tree.Node{
			ID:    "pod/" + id,
			Label: "pod/" + id + " " + green.Render("Running"),
			Children: []*tree.Node{
				{ID: "co/" + id + "/web", Label: "container/web"},
				{ID: "co/" + id + "/proxy", Label: "container/proxy"},
			},
		}
	}
	return []*tree.Node{
		{ID: "deploy/web", Label: "deploy/web", Children: []*tree.Node{
			{ID: "rs/web-7d9f", Label: "rs/web-7d9f", Children: []*tree.Node{pod("web-7d9f-x1"), pod("web-7d9f-y2")}},
			{ID: "svc/web", Label: "svc/web"},
		}},
		{ID: "cm/settings", Label: "cm/settings"},
	}
}

func newModel() model {
	t := tree.New(theme.Default())
	t.SetDefaultDepth(2)
	t.SetRoots(sample())
	return model{tree: t}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.tree.Resize(msg.Width, msg.Height-1)
	case tea.KeyPressMsg:
		if m.tree.HandleKey(msg.String()) {
			return m, nil
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "x":
			m.open = !m.open
			if m.open {
				m.tree.ExpandAll()
			} else {
				m.tree.CollapseAll()
			}
		case "enter":
			var ids []string
			for _, n := range m.tree.SelectedPath() {
				ids = append(ids, n.ID)
			}
			m.status = strings.Join(ids, " > ")
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	v := tea.NewView(m.tree.View() + "\n" + m.status)
	v.AltScreen = true
	return v
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
