// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

// Command chart demonstrates the chart package: a k9s-pulses-style
// dashboard of live sparklines and gauges over made-up data, laid out by
// [chart.Grid]. It samples twice a second; q quits.
package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/blairham/tuikit/chart"
	"github.com/blairham/tuikit/theme"
)

const samples = 256 // enough history for a wide terminal

type tickMsg struct{}

func tick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg { return tickMsg{} })
}

type model struct {
	cpu, mem, net *chart.Series
	cpuChart      *chart.Sparkline
	netChart      *chart.Sparkline
	pods, nodes   *chart.Gauge
	grid          *chart.Grid
	step          float64
}

func newModel() *model {
	t := theme.Default()
	m := &model{
		cpu: chart.NewSeries(samples),
		mem: chart.NewSeries(samples),
		net: chart.NewSeries(samples),
	}
	m.cpuChart = chart.NewSparkline(t, m.cpu, m.mem)
	m.cpuChart.SetTitle("CPU / MEM")
	m.cpuChart.SetMax(100)
	m.netChart = chart.NewSparkline(t, m.net)
	m.netChart.SetTitle("Network")
	m.pods = chart.NewGauge(t)
	m.pods.SetTitle("Pods")
	m.pods.SetLowThresholds(1, 0.5)
	m.nodes = chart.NewGauge(t)
	m.nodes.SetTitle("CPU requests")
	m.nodes.SetHighThresholds(0.7, 0.9)
	m.grid = chart.NewGrid(t, 2, m.cpuChart, m.netChart, m.pods, m.nodes)
	m.sample()
	return m
}

func (m *model) Init() tea.Cmd { return tick() }

// sample pushes one made-up reading into every chart.
func (m *model) sample() {
	m.step++
	cpu := 50 + 35*math.Sin(m.step/9) + rand.Float64()*10
	mem := 40 + 10*math.Sin(m.step/23)
	m.cpu.Push(cpu)
	m.mem.Push(mem)
	m.net.Push(math.Max(0, 200*math.Sin(m.step/5)) + rand.Float64()*40)
	m.cpuChart.SetLabel(fmt.Sprintf("%.0f%% / %.0f%%", cpu, mem))
	m.netChart.SetLabel(fmt.Sprintf("%.0f KiB/s", m.net.Last()))

	running := float64(int(5 + 2*math.Sin(m.step/7) + 2))
	m.pods.Set(running, 9)
	m.pods.SetLabel(fmt.Sprintf("%.0f/9 running", running))
	m.nodes.Set(cpu*0.95, 100)
	m.nodes.SetLabel(fmt.Sprintf("%.0f%%", cpu*0.95))
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.grid.Resize(msg.Width, msg.Height)
	case tickMsg:
		m.sample()
		return m, tick()
	case tea.KeyPressMsg:
		if s := msg.String(); s == "q" || s == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *model) View() tea.View {
	v := tea.NewView(m.grid.View())
	v.AltScreen = true
	return v
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
