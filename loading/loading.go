// Package loading provides a k9s-style "loading…" indicator — an animated
// spinner beside a rotating one-line tip — for use as an initial loading screen
// or a between-views transition while data is fetched.
//
// The app owns the lifecycle: call [Model.Tick] to start the animation (in
// Init or when a load begins), forward each [TickMsg] to [Model.Update] while
// still loading, and render [Model.View] (or [Model.Centered]) until the data
// arrives. Stop forwarding ticks — or just stop rendering it — once loaded.
package loading

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/blairham/tuikit/theme"
)

// TickMsg is the spinner animation tick. Match it in the app's Update and
// forward it to [Model.Update] while loading; the returned cmd keeps the
// animation running.
type TickMsg = spinner.TickMsg

// defaultRotateEvery is how many ticks pass before the tip advances. The dot
// spinner ticks ~10/s, so ~30 ticks ≈ 3s per tip.
const defaultRotateEvery = 30

// Model is a spinner + rotating-tip loading indicator.
type Model struct {
	theme       theme.Theme
	tipStyle    lipgloss.Style
	tips        []string
	spinner     spinner.Model
	tipIndex    int
	ticker      int
	rotateEvery int
}

// New builds a loading model themed from t, cycling the given tips. With no
// tips it shows just the spinner.
func New(t theme.Theme, tips []string) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	// The dot is styled with the theme's logo/brand color (Theme.Logo) rather
	// than Theme.Accent, so it reads as part of the app's identity (orange in
	// Default / NoPaintBackground) instead of the cyan highlight used for
	// titles and selection. The tip keeps Theme.MutedStyle.
	sp.Style = t.On(t.Logo)
	return Model{
		spinner:     sp,
		tips:        tips,
		rotateEvery: defaultRotateEvery,
		tipStyle:    t.MutedStyle,
		theme:       t,
	}
}

// Tick returns the command that starts (or restarts) the animation. Include it
// in Init or batch it when a load begins.
func (m Model) Tick() tea.Cmd { return m.spinner.Tick }

// Update advances the spinner and rotates the tip on a TickMsg. Call it for
// each tick while loading; the returned cmd schedules the next tick. Stop
// calling it (or drop the cmd) when the load completes.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	m.ticker++
	if m.ticker >= m.rotateEvery {
		m.ticker = 0
		m.tipIndex++
	}
	return cmd
}

// Reset returns the rotation to the first tip — call when a new load starts so
// each load opens on the same tip.
func (m *Model) Reset() {
	m.ticker = 0
	m.tipIndex = 0
}

// Tip returns the current tip (empty when no tips were configured).
func (m Model) Tip() string {
	if len(m.tips) == 0 {
		return ""
	}
	return m.tips[m.tipIndex%len(m.tips)]
}

// View renders "spinner tip" on a single line. The separating space is folded
// into the styled tip span so it carries the tip's background — a bare space
// between the two styled spans would render with the terminal default and show
// as a stray gap when the theme paints backgrounds.
func (m Model) View() string {
	tip := m.Tip()
	if tip == "" {
		return m.spinner.View()
	}
	return m.spinner.View() + m.tipStyle.Render(" "+tip)
}

// Centered renders the indicator centered within a width×height box — the usual
// loading-screen / transition layout. It paints a uniform background (when the
// theme calls for it) so there are no unpainted gaps around the centered line.
func (m Model) Centered(width, height int) string {
	s := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center)
	if m.theme.PaintBackground {
		s = s.Background(m.theme.Bg)
	}
	return s.Render(m.View())
}
