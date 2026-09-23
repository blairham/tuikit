package chrome

import (
	"image/color"
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"

	"github.com/blairham/tuikit/theme"
)

// Chrome owns layout constants and shared rendering settings. Apps
// construct one with [New], pass a [Frame] each tick, and call
// [Chrome.Render].
type Chrome struct {
	Theme theme.Theme
	// Logo is the multi-line ASCII art shown in the top-right when the
	// terminal is wide enough. Empty (or nil) disables the logo slot.
	Logo []string
	// InfoLabelWidth is how wide the top-left info panel is. Defaults
	// to 56 — wide enough for "Profile: Production/AdministratorAccess"
	// without wrapping. Apps with longer labels can grow it.
	InfoLabelWidth int
	// ShortcutColumnWidth is the per-row width of the shortcut grid.
	// Defaults to 42.
	ShortcutColumnWidth int
	// MinLogoWidth hides the logo when the terminal is narrower than
	// this. Defaults to 134 (info 56 + shortcut 42 + logo ~32 + slack).
	MinLogoWidth int
	// InfoPanelRows is how many rows the info panel renders. Used by
	// [Chrome.TopSectionRows] to compute the top-section reservation.
	// 0 falls back to defaultInfoPanelRows (4).
	InfoPanelRows int
	// ShortcutRows is how many rows the shortcut grid renders. Used
	// alongside len(Logo) to size the right column of the top section.
	// 0 falls back to defaultShortcutRows (5).
	ShortcutRows int
}

// Config carries the fields apps actually customize. Anything zero
// defaults to a sensible value.
type Config struct {
	Theme               theme.Theme
	Logo                []string
	InfoLabelWidth      int
	ShortcutColumnWidth int
	MinLogoWidth        int
	InfoPanelRows       int
	ShortcutRows        int
}

const (
	defaultInfoPanelRows = 4
	defaultShortcutRows  = 5
)

// New returns a [Chrome] with defaults filled in. Pass an empty Config
// for the canonical look (Default theme, no logo).
func New(cfg Config) Chrome {
	if cfg.InfoLabelWidth == 0 {
		cfg.InfoLabelWidth = 56
	}
	if cfg.ShortcutColumnWidth == 0 {
		cfg.ShortcutColumnWidth = 42
	}
	if cfg.MinLogoWidth == 0 {
		cfg.MinLogoWidth = 134
	}
	t := cfg.Theme
	if t.Bg == nil {
		t = theme.Default()
	}
	return Chrome{
		Theme:               t,
		Logo:                cfg.Logo,
		InfoLabelWidth:      cfg.InfoLabelWidth,
		ShortcutColumnWidth: cfg.ShortcutColumnWidth,
		MinLogoWidth:        cfg.MinLogoWidth,
		InfoPanelRows:       cfg.InfoPanelRows,
		ShortcutRows:        cfg.ShortcutRows,
	}
}

// Frame is the per-tick assembly contract: apps populate the fields
// they want rendered, the chrome handles layout.
//
// InfoLines is the rendered top-left key/value rows.
// Shortcuts is up to 5 rows of pre-rendered shortcut text (use
// [Chrome.Shortcut] / [Chrome.ShortcutPair] to format them).
// Content is the rendered body (typically a bordered table from
// [Chrome.BorderedContent] + [InjectBorderTitle]).
// Breadcrumb is the drill-stack labels, leaf last.
// Filter / Command are the active textinput models or nil. Confirm
// is the y/n prompt text or "" — pair it with [chrome.Confirm], passing
// confirm.Prompt() when confirm.Active() is true. Modal is the active
// [Modal] widget or nil; when non-nil and Modal.Active() the chrome
// paints it centered over the content area in place of Content.
// HelpVisible toggles the help overlay.
// StatusBar / ErrFlash render above the footer when non-empty.
//
//nolint:govet // field order favors readability over fieldalignment in this public API struct
type Frame struct {
	Filter      *textinput.Model
	Command     *textinput.Model
	Modal       *Modal
	Title       string
	Content     string
	StatusBar   string
	Confirm     string
	InfoLines   []string
	Shortcuts   []string
	Breadcrumb  []Crumb
	Help        HelpPanel
	Width       int
	Height      int
	HelpVisible bool
	// StatusBarLevel selects which Theme.Status color paints the bar.
	// Zero value (LevelError) keeps red+bold for callers that don't
	// set it. Set LevelWarn for amber hints (auth, retries) or
	// LevelInfo for blue informational rows.
	StatusBarLevel StatusBarLevel
}

// StatusBarLevel selects the color family for a status bar row.
type StatusBarLevel int

// Status bar levels. Zero value is LevelError so existing callers see
// no behavior change.
const (
	LevelError StatusBarLevel = iota
	LevelWarn
	LevelInfo
)

// Crumb is one segment of the breadcrumb footer.
type Crumb struct {
	Label string
	Leaf  bool
}

// HelpPanel is the content of the help overlay. Apps populate it once
// (typically static) and toggle [Frame.HelpVisible] to show/hide.
type HelpPanel struct {
	Sections []HelpSection
}

// HelpSection is one column of the help overlay.
//
// TitleColor optionally overrides the header color for this section.
// Leave nil to use Theme.Filter (the chrome's default green). k9s
// convention: color resource/hotkey sections in Theme.AccentAlt
// (magenta) and leave general/navigation sections on the default —
// matches the per-section coloring in k9s's help overlay.
type HelpSection struct {
	Title      string
	TitleColor color.Color
	Entries    []HelpEntry
}

// HelpEntry is one (key, description) row inside a [HelpSection].
type HelpEntry struct {
	Key, Desc string
}

// TopSectionRows returns the height (in rows) that [Chrome.renderTopSection]
// will occupy for this Chrome's configured Logo. The right column is
// max(len(Logo), default shortcut rows); the left is the info panel,
// which apps may grow via [Config.InfoPanelRows]. The result is the max
// of the two columns.
func (c Chrome) TopSectionRows() int {
	left := c.InfoPanelRows
	if left <= 0 {
		left = defaultInfoPanelRows
	}
	right := len(c.Logo)
	if right < c.ShortcutRows {
		right = c.ShortcutRows
	}
	if right < defaultShortcutRows {
		right = defaultShortcutRows
	}
	if left > right {
		return left
	}
	return right
}

// ContentInnerSize returns the width and height available inside the
// bordered content area, given the terminal dimensions and which
// optional bars are currently visible. Apps use this to size their
// views before rendering Content.
//
// The top-section reservation is derived from [Chrome.TopSectionRows]
// rather than hardcoded, so apps with shorter (or absent) logos don't
// leave a gap above the footer.
//
// Reservations:
//
//	top section (info + shortcuts):  TopSectionRows()
//	bordered content frame:           2 rows
//	footer (breadcrumb + bottom gap): 2 rows
//	filter bar:                       3 rows when active
//	command bar:                      3 rows when active
//	confirm bar:                      3 rows when active
//	status bar:                       1 row when active
func (c Chrome) ContentInnerSize(w, h int, filtering, commanding, confirming, statusBar bool) (innerW, innerH int) {
	reserved := c.TopSectionRows() + 2 + 2
	innerH = h - reserved
	if filtering {
		innerH -= 3
	}
	if commanding {
		innerH -= 3
	}
	if confirming {
		innerH -= 3
	}
	if statusBar {
		innerH--
	}
	if innerH < 1 {
		innerH = 1
	}
	innerW = w - 4
	if innerW < 10 {
		innerW = 10
	}
	return innerW, innerH
}

// BorderedContent wraps content in a rounded border at the given
// outer width and inner height. The caller typically passes the
// result through [InjectBorderTitle] to add the centered title.
func (c Chrome) BorderedContent(content string, outerWidth, innerHeight int) string {
	return c.Theme.TableBorder.
		Width(outerWidth).
		Height(innerHeight + 2).
		Render(content)
}

// Render assembles a full screen from a [Frame] and returns the
// string to feed to bubbletea. Applications return this from their
// tea.Model.View().
func (c Chrome) Render(f Frame) string {
	if f.Width == 0 || f.Height == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(c.renderTopSection(f))
	if f.Filter != nil {
		sb.WriteString(c.renderFilterBar(f.Filter, f.Width))
		sb.WriteString("\n")
	}
	if f.Command != nil {
		sb.WriteString(c.renderCommandBar(f.Command, f.Width))
		sb.WriteString("\n")
	}
	if f.Confirm != "" {
		sb.WriteString(c.renderConfirmBar(f.Confirm, f.Width))
		sb.WriteString("\n")
	}

	switch {
	case f.HelpVisible:
		sb.WriteString(c.renderHelpOverlay(f))
	case f.Modal != nil && f.Modal.Active():
		sb.WriteString(c.renderModalContent(f))
	default:
		sb.WriteString(f.Content)
	}
	sb.WriteString("\n")
	if f.StatusBar != "" {
		sb.WriteString(c.renderStatusBar(f.StatusBar, f.StatusBarLevel, f.Width))
		sb.WriteString("\n")
	}
	sb.WriteString(c.renderFooter(f.Breadcrumb, f.Width))

	screen := lipgloss.NewStyle().
		Width(f.Width).
		Height(f.Height)
	if c.Theme.PaintBackground {
		screen = screen.Background(c.Theme.Bg).Foreground(c.Theme.Value)
	}
	return screen.Render(sb.String())
}

func (c Chrome) renderTopSection(f Frame) string {
	rightLines := c.assembleShortcutsAndLogo(f.Shortcuts, f.Width)
	logoless := len(c.Logo) == 0 || f.Width < c.MinLogoWidth

	leftWidth := c.InfoLabelWidth
	rightWidth := f.Width - leftWidth
	if rightWidth < 0 {
		rightWidth = 0
	}

	// Layout differs by mode:
	//
	//   - logoless: left-align shortcuts to sit just right of the info
	//     panel (k9s convention). Trailing space inside the rightWidth
	//     block is filled with bg-painted spaces so the chrome's bg
	//     covers the whole top row.
	//
	//   - logo: pin each row's right edge to the chrome's right inset
	//     so the logo (concatenated to the row in assembleShortcutsAndLogo)
	//     hugs the right side; shortcuts left-align inside the row at a
	//     fixed offset from the logo via ShortcutColumnWidth.
	//
	// Trailing fills are rendered through a bg-painting style. Without
	// it, the inner styled content (LogoStyle / shortcut padder) emits a
	// reset escape before any raw trailing space, and the outer
	// styledBlock's bg is not re-applied — those raw cells fall back to
	// the terminal's default background, showing as a gray sliver
	// against the chrome's black.
	alignedRight := make([]string, len(rightLines))
	contentWidth := rightWidth - 1
	if contentWidth < 0 {
		contentWidth = 0
	}
	bgSpaces := func(n int) string {
		if n <= 0 {
			return ""
		}
		s := strings.Repeat(" ", n)
		if c.Theme.PaintBackground {
			s = lipgloss.NewStyle().Background(c.Theme.Bg).Render(s)
		}
		return s
	}
	trailingInset := bgSpaces(1)
	if logoless {
		for i, line := range rightLines {
			fill := contentWidth - lipgloss.Width(line)
			alignedRight[i] = line + bgSpaces(fill) + trailingInset
		}
	} else {
		for i, line := range rightLines {
			lead := contentWidth - lipgloss.Width(line)
			if lead < 0 {
				lead = 0
			}
			alignedRight[i] = strings.Repeat(" ", lead) + line + trailingInset
		}
	}

	// 1-char inset on each side keeps InfoLines and Shortcuts off the
	// chrome's left/right edges symmetrically, matching the footer's
	// leading-space convention.
	leftBlock := c.styledBlock(leftWidth, 0).PaddingLeft(1).Render(strings.Join(f.InfoLines, "\n"))
	rightInner := strings.Join(alignedRight, "\n")

	// Compute heights from the actual wrapped renders — if the info
	// panel wraps (long label/value), both blocks grow to match so
	// JoinHorizontal doesn't pad the shorter block with a stray bg row.
	height := lipgloss.Height(leftBlock)
	if h := lipgloss.Height(rightInner); h > height {
		height = h
	}

	leftBlock = c.styledBlock(leftWidth, height).PaddingLeft(1).Render(strings.Join(f.InfoLines, "\n"))
	rightBlock := c.styledBlock(rightWidth, height).Render(rightInner)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftBlock, rightBlock) + "\n"
}

func (c Chrome) styledBlock(width, height int) lipgloss.Style {
	s := lipgloss.NewStyle()
	if width > 0 {
		s = s.Width(width)
	}
	if height > 0 {
		s = s.Height(height)
	}
	if c.Theme.PaintBackground {
		s = s.Background(c.Theme.Bg)
	}
	return s
}

func (c Chrome) assembleShortcutsAndLogo(shortcuts []string, width int) []string {
	if len(c.Logo) == 0 || width < c.MinLogoWidth {
		// Logoless: return rows at natural width. Padding to
		// ShortcutColumnWidth here would add trailing whitespace that
		// the per-line right-align in renderTopSection treats as
		// visible width, pushing the last char off the chrome edge.
		// renderTopSection right-aligns these rows directly.
		out := make([]string, len(shortcuts))
		copy(out, shortcuts)
		return out
	}

	// Logo branch: pad each shortcut row to ShortcutColumnWidth so the
	// logo column starts at a stable offset across rows.
	padder := lipgloss.NewStyle().Width(c.ShortcutColumnWidth)
	if c.Theme.PaintBackground {
		padder = padder.Background(c.Theme.Bg)
	}

	rows := len(c.Logo)
	if len(shortcuts) > rows {
		rows = len(shortcuts)
	}
	out := make([]string, rows)
	for i := 0; i < rows; i++ {
		s := ""
		if i < len(shortcuts) {
			s = shortcuts[i]
		}
		logo := ""
		if i < len(c.Logo) {
			logo = c.Logo[i]
		}
		out[i] = padder.Render(s) + c.Theme.LogoStyle.Render(logo)
	}
	return out
}

// Shortcut formats a single key/description pair (e.g. "<enter>",
// "Tail") for use in [Frame.Shortcuts]. Only the key is padded to a
// fixed column; the description is left at its natural width so the
// row has no trailing whitespace inside the styled region.
//
// k9s distinguishes two classes of hotkey: "view" keys of the form
// "<N>" (digits, used to switch the active resource view) and "action"
// keys (everything else, used to act on the current row). Views render
// in [Theme.ShortcutView] (magenta by default); actions render in
// [Theme.ShortcutKey] (dodger blue). [Chrome.keyStyle] picks per key
// based on shape — consumers don't need to opt in.
//
// Most rows render two pairs; use [Chrome.ShortcutPair] for that.
func (c Chrome) Shortcut(key, desc string) string {
	return c.keyStyle(key).Render(lipgloss.NewStyle().Width(9).Render(key)) +
		c.Theme.ShortcutDesc.Render(desc)
}

// ShortcutPair formats two key/description pairs onto one row. The
// first description is padded to a fixed column so the second key
// lines up vertically across rows; the second description is rendered
// at natural width so the row ends on visible text. Each key picks its
// color independently via [Chrome.keyStyle].
func (c Chrome) ShortcutPair(k1, d1, k2, d2 string) string {
	kStyle := lipgloss.NewStyle().Width(9)
	dStyle := lipgloss.NewStyle().Width(10)
	return c.keyStyle(k1).Render(kStyle.Render(k1)) +
		c.Theme.ShortcutDesc.Render(dStyle.Render(d1)) +
		c.keyStyle(k2).Render(kStyle.Render(k2)) +
		c.Theme.ShortcutDesc.Render(d2)
}

// Shortcut is a key + description pair for the registered-shortcut
// layout helpers ([Chrome.ShortcutGrid]). Use [chrome.Shortcut] values
// to declare a view's hotkeys once per view and an app's view-switch
// hotkeys once globally; the chrome handles layout.
type Shortcut struct {
	Key  string
	Desc string
}

// ShortcutGrid lays out registered shortcuts k9s-style: view-switch
// hotkeys stack vertically in the leftmost column, actions stack in
// the second column. Returns rows ready for [Frame.Shortcuts] —
// consumers register views and per-view actions once and let the
// chrome render the grid.
//
// Row count = max(len(views), len(actions)). When one column has
// fewer entries the missing cells render as empty (width-padded)
// pairs so column alignment is preserved.
//
// Coloring is automatic: view-shaped keys (<N>) pick up
// [Theme.ShortcutView], actions pick up [Theme.ShortcutKey] — see
// [Chrome.keyStyle].
func (c Chrome) ShortcutGrid(views, actions []Shortcut) []string {
	rows := max(len(views), len(actions))
	out := make([]string, rows)
	for i := range rows {
		var v, a Shortcut
		if i < len(views) {
			v = views[i]
		}
		if i < len(actions) {
			a = actions[i]
		}
		out[i] = c.ShortcutPair(v.Key, v.Desc, a.Key, a.Desc)
	}
	return out
}

// keyStyle picks the foreground style for a shortcut key based on
// whether it's a view-switch or an action. View keys ("<N>" where N is
// one or more digits) render in [Theme.ShortcutView]; everything else
// renders in [Theme.ShortcutKey].
func (c Chrome) keyStyle(key string) lipgloss.Style {
	if isViewKey(key) {
		return c.Theme.ShortcutView
	}
	return c.Theme.ShortcutKey
}

// isViewKey reports whether key is a k9s-style view-switch hotkey of
// the form "<N>" where N is one or more decimal digits.
func isViewKey(key string) bool {
	if len(key) < 3 || key[0] != '<' || key[len(key)-1] != '>' {
		return false
	}
	for i := 1; i < len(key)-1; i++ {
		if key[i] < '0' || key[i] > '9' {
			return false
		}
	}
	return true
}

// infoLabelColumnWidth is the conventional column width for info-panel
// labels (e.g. "Context:  ", "Auth:     "). Used by [Chrome.VersionLine]
// so its rendered row aligns with hand-formatted [Frame.InfoLines].
const infoLabelColumnWidth = 10

// VersionLine renders a k9s-style version row for the info panel. The
// label is padded to [infoLabelColumnWidth] columns and rendered with
// Theme.InfoLabel; current is rendered with Theme.InfoValue. When
// latest is non-empty and differs from current, " ⚡ <latest>" is
// appended in Theme.Status.Warn (yellow) — k9s's "newer version
// available" cue. Returns "" if current is "".
//
// Apps include the result in [Frame.InfoLines]; bump
// [Config.InfoPanelRows] so the chrome reserves room for the extra row
// when sizing the content area.
func (c Chrome) VersionLine(label, current, latest string) string {
	if current == "" {
		return ""
	}
	line := c.Theme.InfoLabel.Render(padRight(label, infoLabelColumnWidth)) +
		c.Theme.InfoValue.Render(current)
	if latest != "" && latest != current {
		line += " " + c.Theme.On(c.Theme.Status.Warn).Bold(true).Render("⚡ "+latest)
	}
	return line
}

func (c Chrome) renderFilterBar(input *textinput.Model, width int) string {
	// lipgloss/v2 Width includes borders in the rendered cell count, so
	// Width(width) — not Width(width-2) — produces a bar that spans the
	// full frame and lines up with the bordered content beneath it.
	box := lipgloss.NewStyle().
		Width(width).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(c.Theme.Filter).
		Padding(0, 1)
	if c.Theme.PaintBackground {
		box = box.Background(c.Theme.Bg)
	}
	return box.Render(input.View())
}

func (c Chrome) renderConfirmBar(prompt string, width int) string {
	inner := width - 4
	if inner < 10 {
		inner = 10
	}
	body := c.Theme.Error.Render(WrapText(prompt+" (y/n)", inner))
	bodyStyle := lipgloss.NewStyle().Width(inner)
	if c.Theme.PaintBackground {
		bodyStyle = bodyStyle.Background(c.Theme.Bg)
	}
	box := lipgloss.NewStyle().
		Width(width).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(c.Theme.Status.Error).
		Padding(0, 1)
	if c.Theme.PaintBackground {
		box = box.Background(c.Theme.Bg)
	}
	return box.Render(bodyStyle.Render(body))
}

func (c Chrome) renderCommandBar(input *textinput.Model, width int) string {
	inner := width - 4
	if inner < 10 {
		inner = 10
	}
	body := input.View()
	bodyStyle := lipgloss.NewStyle().Width(inner)
	if c.Theme.PaintBackground {
		bodyStyle = bodyStyle.Background(c.Theme.Bg)
	}
	box := lipgloss.NewStyle().
		Width(width).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(c.Theme.Accent).
		Padding(0, 1)
	if c.Theme.PaintBackground {
		box = box.Background(c.Theme.Bg)
	}
	return box.Render(bodyStyle.Render(body))
}

func (c Chrome) renderStatusBar(msg string, level StatusBarLevel, width int) string {
	innerWidth := width - 4
	if innerWidth < 20 {
		innerWidth = 20
	}
	var fg color.Color
	switch level {
	case LevelWarn:
		fg = c.Theme.Status.Warn
	case LevelInfo:
		fg = c.Theme.Status.Info
	default:
		fg = c.Theme.Status.Error
	}
	// Height(1) matches the 1-row reservation in ContentInnerSize so
	// the rendered output fills the slot the chrome reserved for it
	// (and no more — anything taller would push the footer down off
	// the terminal). A single line sits one row above the breadcrumb
	// pill, mirroring the bordered-content → footer spacing.
	s := lipgloss.NewStyle().
		Width(width).
		Height(1).
		Foreground(fg).
		Bold(true)
	if c.Theme.PaintBackground {
		s = s.Background(c.Theme.Bg)
	}
	return s.Render(" " + WrapText(msg, innerWidth))
}

func (c Chrome) renderFooter(crumbs []Crumb, width int) string {
	sep := c.Theme.MutedStyle.Render(" › ")
	leaf := c.Theme.PromptStyle
	pill := lipgloss.NewStyle().
		Background(c.Theme.BreadcrumbBg).
		Foreground(c.Theme.Value).
		Padding(0, 1)

	parts := make([]string, 0, len(crumbs))
	for _, cr := range crumbs {
		if cr.Leaf {
			parts = append(parts, leaf.Render(cr.Label))
		} else {
			parts = append(parts, pill.Render(cr.Label))
		}
	}

	footer := lipgloss.NewStyle().Width(width).Height(2)
	if c.Theme.PaintBackground {
		footer = footer.Background(c.Theme.Bg)
	}
	return footer.Render(" " + strings.Join(parts, sep))
}

func (c Chrome) renderHelpOverlay(f Frame) string {
	width := f.Width
	innerW, innerH := c.ContentInnerSize(
		width, f.Height,
		f.Filter != nil, f.Command != nil, f.Confirm != "", f.StatusBar != "",
	)

	if len(f.Help.Sections) == 0 {
		body := CenterInBox("No help configured.", innerW, innerH, c.Theme)
		box := c.helpBorderedContent(body, width-2, innerH)
		return InjectBorderTitle(box, c.helpTitleStyle().Render("help"), c.Theme)
	}

	colWidth := (width - 8) / len(f.Help.Sections)
	if colWidth < 25 {
		colWidth = 25
	}
	cols := make([]string, 0, len(f.Help.Sections))
	for _, section := range f.Help.Sections {
		cols = append(cols, c.renderHelpSection(section, colWidth))
	}
	body := lipgloss.JoinHorizontal(lipgloss.Top, cols...)
	box := c.helpBorderedContent(body, width-2, innerH)
	return InjectBorderTitle(box, c.helpTitleStyle().Render("help"), c.Theme)
}

// helpTitleStyle returns the bold pill style for the help overlay's
// border-title text. Uses [Theme.HelpTitle] (k9s-style muted red) so
// the help title visually stands apart from action/view titles
// elsewhere in the chrome.
func (c Chrome) helpTitleStyle() lipgloss.Style {
	return c.Theme.On(c.Theme.HelpTitle).Bold(true)
}

// helpBorderedContent wraps content in a help-overlay border whose
// foreground is [Theme.HelpBorder] (cyan), not the default
// [Theme.Border] (dodger blue) used for the main content border —
// matches k9s's distinct help-overlay border.
func (c Chrome) helpBorderedContent(content string, outerWidth, innerHeight int) string {
	s := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(c.Theme.HelpBorder).
		Width(outerWidth).
		Height(innerHeight + 2)
	if c.Theme.PaintBackground {
		s = s.BorderBackground(c.Theme.Bg).Background(c.Theme.Bg)
	}
	return s.Render(content)
}

func (c Chrome) renderHelpSection(s HelpSection, colWidth int) string {
	headerColor := s.TitleColor
	if headerColor == nil {
		headerColor = c.Theme.Filter
	}
	header := c.Theme.On(headerColor).Bold(true).Underline(true)
	desc := c.Theme.On(c.Theme.HelpDesc)

	var sb strings.Builder
	sb.WriteString(header.Render(s.Title))
	sb.WriteString("\n")
	for _, e := range s.Entries {
		// Match the top-section convention: view-shaped keys (<N>)
		// render in ShortcutView (magenta), everything else in
		// ShortcutKey (blue). Keeps help and shortcut bar consistent.
		sb.WriteString(c.keyStyle(e.Key).Render(padRight(e.Key, 14)))
		sb.WriteString(desc.Render(e.Desc))
		sb.WriteString("\n")
	}
	colStyle := lipgloss.NewStyle().Width(colWidth)
	if c.Theme.PaintBackground {
		colStyle = colStyle.Background(c.Theme.Bg)
	}
	return colStyle.Render(sb.String())
}

func padRight(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}
