package chrome

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/blairham/tuikit/theme"
)

// Modal is a centered popup confirmation widget. Where [Confirm] reserves
// a bordered strip at the bottom of the screen, Modal renders as a small
// bordered box overlaid on the content area — the shape k9s uses for
// delete/restart prompts. Apps reach for Modal when the prompt warrants
// a visual interrupt; Confirm stays the right choice for inline yes/no.
//
// Keys match [Confirm]: y/Y/Enter = yes, n/N/Esc = no, others swallowed.
// The visual difference is layout only — same single-keystroke semantics.
type Modal struct {
	theme       theme.Theme
	title       string
	prompt      string
	okLabel     string
	cancelLabel string
	active      bool
}

// ModalDispatch is the callback invoked with the y/n result. Identical
// semantics to [ConfirmDispatch]: non-empty errMsg keeps the modal open
// so the user can re-answer; "" closes it.
type ModalDispatch func(yes bool) (errMsg string, cmd tea.Cmd)

// ModalOpts customizes the per-Open title and button labels. Zero values
// fall back to "Confirm" / "OK" / "Cancel".
type ModalOpts struct {
	Title       string
	OkLabel     string
	CancelLabel string
}

// NewModal returns a Modal ready for use. The rendered border + buttons
// use the theme's chrome palette (Theme.Border for the frame, Theme.Title
// for the centered title pill) so the popup looks like a sibling of the
// other bordered content surfaces. Apps that want a different palette
// can recolor by constructing their own Theme.
func NewModal(t theme.Theme) *Modal {
	return &Modal{theme: t}
}

// Active reports whether the modal is currently open.
func (m *Modal) Active() bool { return m.active }

// Open activates the modal with the given prompt + per-open options.
// Pass just the question text — the chrome adds the surrounding border,
// title, and button row automatically.
func (m *Modal) Open(prompt string, opts ModalOpts) {
	m.active = true
	m.prompt = prompt
	m.title = opts.Title
	if m.title == "" {
		m.title = "Confirm"
	}
	m.okLabel = opts.OkLabel
	if m.okLabel == "" {
		m.okLabel = "OK"
	}
	m.cancelLabel = opts.CancelLabel
	if m.cancelLabel == "" {
		m.cancelLabel = "Cancel"
	}
}

// Close deactivates the modal and clears the stored prompt.
func (m *Modal) Close() {
	m.active = false
	m.prompt = ""
	m.title = ""
	m.okLabel = ""
	m.cancelLabel = ""
}

// Prompt returns the active question text ("" when inactive).
func (m *Modal) Prompt() string { return m.prompt }

// Title returns the active modal title ("" when inactive).
func (m *Modal) Title() string { return m.title }

// Update consumes a single tea.KeyMsg.
//
//   - y / Y / enter: dispatch(true).
//   - n / N / esc:   dispatch(false).
//   - any other key: swallowed, returns handled=true, no dispatch.
//
// Non-KeyMsg messages are not consumed (handled=false). When inactive,
// Update is a no-op and returns handled=false.
func (m *Modal) Update(msg tea.Msg, dispatch ModalDispatch) (handled bool, cmd tea.Cmd) {
	if !m.active {
		return false, nil
	}
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return false, nil
	}
	switch key.String() {
	case "y", "Y", keyStrEnter:
		return true, m.fire(dispatch, true)
	case "n", "N", keyStrEsc:
		return true, m.fire(dispatch, false)
	}
	return true, nil
}

func (m *Modal) fire(dispatch ModalDispatch, yes bool) tea.Cmd {
	if dispatch == nil {
		m.Close()
		return nil
	}
	errMsg, cmd := dispatch(yes)
	if errMsg != "" {
		return cmd
	}
	m.Close()
	return cmd
}

// renderModalContent fills the content area with a centered modal box.
// Called from [Chrome.Render] when [Frame.Modal] is active; replaces
// the normal Content rendering rather than overlaying it (lipgloss has
// no cell-level compositor). The modal canvas is wrapped in the same
// [Chrome.BorderedContent] frame the app would have rendered behind it,
// so the exterior chrome border stays visible while the modal is open.
//
// The bordered wrapper spans the full f.Width — matching the convention
// apps use for normal Content — so the right edge lines
// up with the top section and bars rather than leaving a 2-cell gap.
func (c Chrome) renderModalContent(f Frame) string {
	_, innerH := c.ContentInnerSize(
		f.Width, f.Height,
		f.Filter != nil, f.Command != nil, f.Confirm != "", f.StatusBar != "",
	)
	canvasW := f.Width - 2
	if canvasW < 10 {
		canvasW = 10
	}
	canvas := c.renderModalOverlay(f.Modal, canvasW, innerH)
	return c.BorderedContent(canvas, f.Width, innerH)
}

// renderModalOverlay returns a rendering of the modal box placed at the
// center of an areaW × areaH cell. The caller stitches this into the
// frame in place of the normal Content rendering when Modal.Active().
func (c Chrome) renderModalOverlay(m *Modal, areaW, areaH int) string {
	// Pick a modal width: wide enough for the prompt but capped so the
	// shape stays compact. ~60% of the area, with sane floor/ceiling.
	boxW := areaW * 6 / 10
	if boxW < 36 {
		boxW = 36
	}
	if boxW > 80 {
		boxW = 80
	}
	if boxW > areaW-4 {
		boxW = areaW - 4
	}
	if boxW < 20 {
		boxW = 20
	}

	innerW := boxW - 4 // -2 border, -2 horizontal padding
	if innerW < 10 {
		innerW = 10
	}

	body := WrapText(m.prompt, innerW)
	bodyStyle := lipgloss.NewStyle().Width(innerW).Foreground(c.Theme.Value)
	if c.Theme.PaintBackground {
		bodyStyle = bodyStyle.Background(c.Theme.Bg)
	}

	buttons := c.renderModalButtons(m, innerW)

	gapStyle := lipgloss.NewStyle().Width(innerW)
	if c.Theme.PaintBackground {
		gapStyle = gapStyle.Background(c.Theme.Bg)
	}
	gap := gapStyle.Render("")

	content := strings.Join([]string{
		bodyStyle.Render(body),
		gap,
		buttons,
	}, "\n")

	box := lipgloss.NewStyle().
		Width(boxW).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(c.Theme.Border).
		Padding(0, 1)
	if c.Theme.PaintBackground {
		box = box.Background(c.Theme.Bg).BorderBackground(c.Theme.Bg)
	}
	rendered := box.Render(content)

	rendered = InjectBorderTitle(rendered, c.Theme.Title.Render(m.title), c.Theme)

	canvas := lipgloss.NewStyle().Width(areaW).Height(areaH).Align(lipgloss.Center, lipgloss.Center)
	if c.Theme.PaintBackground {
		canvas = canvas.Background(c.Theme.Bg)
	}
	return canvas.Render(rendered)
}

func (c Chrome) renderModalButtons(m *Modal, innerW int) string {
	// Two pill-style buttons: Cancel (muted) and OK (filled). The OK
	// pill is highlighted because Enter triggers it — matches the k9s
	// convention where the default action is visually emphasized.
	cancel := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(c.Theme.Muted).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(c.Theme.Muted)
	ok := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(c.Theme.PromptText).
		Background(c.Theme.Border).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(c.Theme.Border).
		Bold(true)
	// Buttons are 3 rows tall (top border / content / bottom border).
	// The gap between them must also be 3 rows of bg-painted cells —
	// a Height(1) gap lets JoinHorizontal pad the missing rows with the
	// terminal default, which shows through as a gray block between the
	// two pills under PaintBackground.
	gapStyle := lipgloss.NewStyle().Width(2).Height(3)
	if c.Theme.PaintBackground {
		cancel = cancel.Background(c.Theme.Bg).BorderBackground(c.Theme.Bg)
		ok = ok.BorderBackground(c.Theme.Bg)
		gapStyle = gapStyle.Background(c.Theme.Bg)
	}

	row := lipgloss.JoinHorizontal(
		lipgloss.Top,
		cancel.Render(m.cancelLabel),
		gapStyle.Render(""),
		ok.Render(m.okLabel),
	)

	rowStyle := lipgloss.NewStyle().Width(innerW).Align(lipgloss.Center)
	if c.Theme.PaintBackground {
		rowStyle = rowStyle.Background(c.Theme.Bg)
	}
	return rowStyle.Render(row)
}
