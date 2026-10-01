package chrome

import (
	"image/color"
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

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
	// InfoLabelWidth is the most the top-left info panel may take,
	// including its 1-cell left inset. The panel shrinks to its widest
	// [Frame.InfoLines] row plus a small gap, so the shortcut block
	// starts right after it (k9s); a row wider than this wraps.
	// Defaults to 56 — wide enough for
	// "Profile: Production/AdministratorAccess" without wrapping. Apps
	// with longer labels can grow it.
	InfoLabelWidth int
	// ShortcutColumnWidth is unused.
	//
	// Deprecated: the shortcut block is left-aligned after the info
	// panel and the logo is pinned to the right edge independently, so
	// no per-row shortcut width is needed to keep the logo aligned. The
	// field is kept so existing configs compile.
	ShortcutColumnWidth int
	// MinLogoWidth hides the logo when the terminal is narrower than
	// this. Defaults to 134 (info 56 + shortcut 42 + logo ~32 + slack).
	MinLogoWidth int
	// InfoPanelRows is how many rows the info panel renders. Used by
	// [Chrome.TopSectionRows] to compute the top-section reservation.
	// 0 falls back to defaultInfoPanelRows (4).
	InfoPanelRows int
	// ShortcutRows is the height of a shortcut column: [Chrome.ShortcutGrid]
	// wraps views and actions into a new column every ShortcutRows
	// entries, capped at [Chrome.TopSectionRows]. It also feeds
	// TopSectionRows alongside len(Logo) and InfoPanelRows. 0 falls back
	// to defaultShortcutRows (6, as in k9s).
	ShortcutRows int
	// ShortcutKeyWidth is the minimum width of a shortcut key column
	// in [Chrome.Shortcut], [Chrome.ShortcutPair] and
	// [Chrome.ShortcutGrid]. A key at least this wide grows its column
	// so one space always separates it from its description. 0 falls
	// back to defaultShortcutKeyWidth (9).
	ShortcutKeyWidth int
	// ShortcutDescWidth is the minimum width of a padded shortcut
	// description column (every description except the last on a row).
	// A description at least this wide grows its column so one space
	// always separates it from the next key. 0 falls back to
	// defaultShortcutDescWidth (10).
	ShortcutDescWidth int
	// CrumbsHidden drops the breadcrumb footer entirely: [Chrome.Render]
	// skips it and [Chrome.ContentInnerSize] releases its 2-row
	// reservation, so the bordered content grows to fill the space.
	// Apps flip it with [Chrome.ToggleCrumbs], conventionally bound to
	// [KeyToggleCrumbs] (ctrl+g, as in k9s).
	CrumbsHidden bool
	// HeaderHidden drops the top section (info panel, shortcuts, logo)
	// entirely: [Chrome.Render] skips it and [Chrome.ContentInnerSize]
	// releases its TopSectionRows() reservation, so the bordered content
	// grows into the freed rows. Apps flip it with [Chrome.ToggleHeader],
	// conventionally bound to [KeyToggleHeader] (ctrl+e, as in k9s).
	HeaderHidden bool
}

// ToggleHeader shows or hides the top section. See [Chrome.HeaderHidden].
func (c *Chrome) ToggleHeader() {
	c.HeaderHidden = !c.HeaderHidden
}

// headerRows is the reservation for the top section, or nothing at all
// when the header is hidden.
func (c Chrome) headerRows() int {
	if c.HeaderHidden {
		return 0
	}
	return c.TopSectionRows()
}

// ToggleCrumbs shows or hides the breadcrumb footer. See
// [Chrome.CrumbsHidden].
func (c *Chrome) ToggleCrumbs() {
	c.CrumbsHidden = !c.CrumbsHidden
}

// footerRows is the reservation for the breadcrumb footer: the pill row
// plus a bottom gap, or nothing at all when the crumbs are hidden.
func (c Chrome) footerRows() int {
	if c.CrumbsHidden {
		return 0
	}
	return 2
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
	ShortcutKeyWidth    int
	ShortcutDescWidth   int
}

const (
	defaultInfoPanelRows = 4
	defaultShortcutRows  = 6

	defaultShortcutKeyWidth  = 9
	defaultShortcutDescWidth = 10
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
		ShortcutKeyWidth:    cfg.ShortcutKeyWidth,
		ShortcutDescWidth:   cfg.ShortcutDescWidth,
	}
}

// Frame is the per-tick assembly contract: apps populate the fields
// they want rendered, the chrome handles layout.
//
// InfoLines is the rendered top-left key/value rows.
// Shortcuts is up to [Chrome.TopSectionRows] rows of pre-rendered
// shortcut text (use [Chrome.ShortcutGrid], which wraps into extra
// columns to stay within that, or [Chrome.Shortcut] /
// [Chrome.ShortcutPair]). Rows past TopSectionRows are not drawn, so
// the frame never outgrows the reservation [Chrome.ContentInnerSize]
// made for it.
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
// Leave nil to use Theme.HelpSection: k9s draws every section heading
// in the same plain green, so an override is for an app that wants to
// set one section apart, not the k9s look.
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
// max(len(Logo), ShortcutRows, 6); the left is the info panel,
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
//	top section (info + shortcuts):  TopSectionRows(), 0 when HeaderHidden
//	bordered content frame:           2 rows
//	footer (breadcrumb + bottom gap): 2 rows, 0 when CrumbsHidden
//	filter bar:                       3 rows when active
//	command bar:                      3 rows when active
//	confirm bar:                      3 rows when active
//	status bar:                       1 row when active
func (c Chrome) ContentInnerSize(w, h int, filtering, commanding, confirming, statusBar bool) (innerW, innerH int) {
	reserved := c.headerRows() + 2 + c.footerRows()
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
	if !c.HeaderHidden {
		sb.WriteString(c.renderTopSection(f))
	}
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
	if f.StatusBar != "" {
		sb.WriteString("\n")
		sb.WriteString(c.renderStatusBar(f.StatusBar, f.StatusBarLevel, f.Width))
	}
	if !c.CrumbsHidden {
		sb.WriteString("\n")
		sb.WriteString(c.renderFooter(f.Breadcrumb, f.Width))
	}

	screen := lipgloss.NewStyle().
		Width(f.Width).
		Height(f.Height)
	if c.Theme.PaintBackground {
		screen = screen.Background(c.Theme.Bg).Foreground(c.Theme.Value)
		// The screen's Background only paints padding; cells after an
		// inner span's reset fall back to the terminal default without this.
		return theme.ReassertBackground(screen.Render(sb.String()), theme.BackgroundSeq(c.Theme.Bg))
	}
	return screen.Render(sb.String())
}

// infoGap is the blank space between the widest info-panel row and the
// first shortcut column.
const infoGap = 2

// infoBlockWidth is the width of the top-left info block: the 1-cell
// left inset, the widest info row and [infoGap], capped at
// InfoLabelWidth (past which a row wraps, as before).
func (c Chrome) infoBlockWidth(info []string) int {
	w := 0
	for _, l := range info {
		w = max(w, lipgloss.Width(l))
	}
	w = 1 + w + infoGap
	if c.InfoLabelWidth > 0 {
		w = min(w, c.InfoLabelWidth)
	}
	return w
}

func (c Chrome) renderTopSection(f Frame) string {
	// The top section must never be taller than TopSectionRows():
	// ContentInnerSize sized the content box from that number before
	// this frame existed, so every row drawn past it pushes the
	// content's bottom border and the footer off the terminal.
	maxRows := c.TopSectionRows()
	shortcuts := f.Shortcuts
	if len(shortcuts) > maxRows {
		shortcuts = shortcuts[:maxRows]
	}

	leftWidth := min(c.infoBlockWidth(f.InfoLines), f.Width)
	rightWidth := f.Width - leftWidth

	// k9s layout, the same with or without a logo:
	//
	//	info | gap | shortcut columns (left-aligned) | fill | logo | inset
	//
	// The shortcut block starts right after the info panel; the logo is
	// pinned as a block (every line padded to the widest) against the
	// 1-cell right inset, and the fill between them absorbs the slack.
	// Rows past the end of the logo get a blank logo-width segment, and
	// since shortcuts are left-aligned they line up whatever is to
	// their right.
	//
	// Every fill is rendered through a bg-painting style. Without it,
	// the inner styled content (LogoStyle / shortcut styles) emits a
	// reset escape before any raw trailing space, and the outer
	// styledBlock's bg is not re-applied — those raw cells fall back to
	// the terminal's default background, showing as a gray sliver
	// against the chrome's black.
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
	contentWidth := max(rightWidth-1, 0) // 1-cell right inset

	var logo []string
	logoWidth := 0
	if len(c.Logo) > 0 && f.Width >= c.MinLogoWidth {
		logo = c.Logo
		for _, l := range logo {
			logoWidth = max(logoWidth, lipgloss.Width(l))
		}
	}
	// The shortcut area keeps at least one blank cell before the logo; a
	// row too wide for it is truncated rather than wrapped, so it can
	// neither run into the logo nor push the rows below it down.
	shortcutWidth := contentWidth
	if logoWidth > 0 {
		shortcutWidth = max(contentWidth-logoWidth-1, 0)
	}

	rows := min(max(len(shortcuts), len(logo)), maxRows)
	rightLines := make([]string, rows)
	for i := range rows {
		s := ""
		if i < len(shortcuts) {
			s = shortcuts[i]
		}
		if lipgloss.Width(s) > shortcutWidth {
			s = ansi.Truncate(s, shortcutWidth, "")
		}
		line := s + bgSpaces(shortcutWidth-lipgloss.Width(s))
		if logoWidth > 0 {
			l := ""
			if i < len(logo) {
				l = logo[i]
			}
			line += bgSpaces(contentWidth - shortcutWidth - logoWidth)
			if l != "" {
				line += c.Theme.LogoStyle.Render(l)
			}
			line += bgSpaces(logoWidth - lipgloss.Width(l))
		}
		rightLines[i] = line + bgSpaces(1)
	}

	// 1-char inset on each side keeps InfoLines and Shortcuts off the
	// chrome's left/right edges symmetrically, matching the footer's
	// leading-space convention.
	// Both blocks are exactly TopSectionRows() tall: a wrapped info row
	// or extra shortcut rows are clipped, and a header with fewer rows of
	// content is padded, so the content box ContentInnerSize sized always
	// starts directly below the header and ends directly above the footer.
	height := maxRows
	rightInner := strings.Join(rightLines, "\n")
	leftBlock := c.styledBlock(leftWidth, height).MaxHeight(height).PaddingLeft(1).Render(strings.Join(f.InfoLines, "\n"))
	rightBlock := c.styledBlock(rightWidth, height).MaxHeight(height).Render(rightInner)

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

// Shortcut formats a single key/description pair (e.g. "<enter>",
// "Tail") for use in [Frame.Shortcuts]. Only the key is padded to a
// column of [Chrome.ShortcutKeyWidth] (grown so at least one space
// follows the key); the description is left at its natural width so
// the row has no trailing whitespace inside the styled region.
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
	return c.shortcutCell(Shortcut{Key: key, Desc: desc}, c.shortcutKeyWidth(key), 0)
}

// ShortcutPair formats two key/description pairs onto one row. The
// first description is padded to [Chrome.ShortcutDescWidth] so the
// second key lines up vertically across rows; the second description
// is rendered at natural width so the row ends on visible text. A key
// or description too wide for its column grows the column, so at
// least one space always separates adjacent columns (alignment across
// rows then holds only for rows whose cells fit — use
// [Chrome.ShortcutGrid] to size columns from all rows at once). Each
// key picks its color independently via [Chrome.keyStyle].
func (c Chrome) ShortcutPair(k1, d1, k2, d2 string) string {
	return c.shortcutCell(Shortcut{Key: k1, Desc: d1}, c.shortcutKeyWidth(k1), c.shortcutDescWidth(d1)) +
		c.shortcutCell(Shortcut{Key: k2, Desc: d2}, c.shortcutKeyWidth(k2), 0)
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
// the column(s) after it. Returns rows ready for [Frame.Shortcuts] —
// consumers register views and per-view actions once and let the
// chrome render the grid.
//
// Columns are [Chrome.ShortcutRows] tall (6 by default, as in k9s): a
// list longer than that wraps into additional columns of at most that
// many entries (views first, then actions). The column height is capped
// at [Chrome.TopSectionRows], so the header never outgrows the
// reservation [Chrome.ContentInnerSize] made for it. Otherwise the
// row count is max(len(views), len(actions)), and when one column has
// fewer entries the missing cells render as empty (width-padded)
// pairs so column alignment is preserved.
//
// Each column is as wide as its widest key and description plus a
// one-space gap, but never narrower than [Chrome.ShortcutKeyWidth] /
// [Chrome.ShortcutDescWidth], so columns line up across rows and never
// run together.
//
// Coloring is automatic: view-shaped keys (<N>) pick up
// [Theme.ShortcutView], actions pick up [Theme.ShortcutKey] — see
// [Chrome.keyStyle].
func (c Chrome) ShortcutGrid(views, actions []Shortcut) []string {
	if len(views) == 0 && len(actions) == 0 {
		return nil
	}
	maxRows := c.ShortcutRows
	if maxRows <= 0 {
		maxRows = defaultShortcutRows
	}
	maxRows = min(maxRows, c.TopSectionRows())
	cols := append(chunkShortcuts(views, maxRows), chunkShortcuts(actions, maxRows)...)

	rows := 0
	keyW := make([]int, len(cols))
	descW := make([]int, len(cols))
	for i, col := range cols {
		rows = max(rows, len(col))
		keyW[i] = c.shortcutKeyWidth("")
		descW[i] = c.shortcutDescWidth("")
		for _, s := range col {
			keyW[i] = max(keyW[i], c.shortcutKeyWidth(s.Key))
			descW[i] = max(descW[i], c.shortcutDescWidth(s.Desc))
		}
	}

	out := make([]string, rows)
	for r := range rows {
		// Drop trailing empty cells so a row ends on visible text.
		last := len(cols) - 1
		for last > 0 && r >= len(cols[last]) {
			last--
		}
		var sb strings.Builder
		for i := 0; i <= last; i++ {
			var s Shortcut
			if r < len(cols[i]) {
				s = cols[i][r]
			}
			dw := descW[i]
			if i == last {
				dw = 0
			}
			sb.WriteString(c.shortcutCell(s, keyW[i], dw))
		}
		out[r] = sb.String()
	}
	return out
}

// chunkShortcuts splits list into columns of at most n entries. An
// empty list still yields one (empty) column so a grid with no views
// keeps its action column in the same place.
func chunkShortcuts(list []Shortcut, n int) [][]Shortcut {
	if len(list) == 0 {
		return [][]Shortcut{nil}
	}
	n = max(n, 1)
	var out [][]Shortcut
	for len(list) > n {
		out = append(out, list[:n])
		list = list[n:]
	}
	return append(out, list)
}

// shortcutCell renders one key/description pair with the key padded to
// keyWidth and the description padded to descWidth (0 = natural width).
func (c Chrome) shortcutCell(s Shortcut, keyWidth, descWidth int) string {
	desc := s.Desc
	if descWidth > 0 {
		desc = padRight(desc, descWidth)
	}
	return c.keyStyle(s.Key).Render(padRight(s.Key, keyWidth)) + c.Theme.ShortcutDesc.Render(desc)
}

// shortcutKeyWidth is the key column width that fits key with at least
// one space after it, floored at the configured ShortcutKeyWidth.
func (c Chrome) shortcutKeyWidth(key string) int {
	w := c.ShortcutKeyWidth
	if w <= 0 {
		w = defaultShortcutKeyWidth
	}
	return max(w, lipgloss.Width(key)+1)
}

// shortcutDescWidth is the padded description column width that fits
// desc with at least one space after it, floored at the configured
// ShortcutDescWidth.
func (c Chrome) shortcutDescWidth(desc string) int {
	w := c.ShortcutDescWidth
	if w <= 0 {
		w = defaultShortcutDescWidth
	}
	return max(w, lipgloss.Width(desc)+1)
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
		BorderForeground(c.Theme.CommandBorder).
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
		return InjectBorderTitleColor(box, c.helpTitleStyle().Render("help"), c.Theme.HelpBorder, c.Theme)
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
	return InjectBorderTitleColor(box, c.helpTitleStyle().Render("help"), c.Theme.HelpBorder, c.Theme)
}

// helpTitleStyle returns the bold pill style for the help overlay's
// border-title text. Uses [Theme.HelpTitle] (k9s-style muted red) so
// the help title visually stands apart from action/view titles
// elsewhere in the chrome.
func (c Chrome) helpTitleStyle() lipgloss.Style {
	return c.Theme.On(c.Theme.HelpTitle).Bold(true)
}

// helpBorderedContent wraps content in a help-overlay border whose
// foreground is [Theme.HelpBorder] (cyan), not the focused
// [Theme.FocusBorder] (light sky blue) used for the main content border —
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
		headerColor = c.Theme.HelpSection
	}
	// Plain, as k9s draws its section headings: no bold, no underline.
	header := c.Theme.On(headerColor)
	desc := c.Theme.On(c.Theme.HelpDesc)

	// The key column is as wide as the section's widest key plus a space,
	// as in k9s; a fixed width left too little room for descriptions in a
	// four-section overlay on a 120-column terminal, and they wrapped
	// into the row below. A description that still does not fit ends in
	// "…", keeping one row per entry and one cell clear before the next
	// column.
	keyW := 0
	for _, e := range s.Entries {
		keyW = max(keyW, lipgloss.Width(e.Key))
	}
	keyW++
	descW := max(colWidth-keyW-1, 1)

	var sb strings.Builder
	sb.WriteString(header.Render(s.Title))
	sb.WriteString("\n")
	for _, e := range s.Entries {
		// Match the top-section convention: view-shaped keys (<N>)
		// render in ShortcutView (magenta), everything else in
		// ShortcutKey (blue). Keeps help and shortcut bar consistent.
		sb.WriteString(c.keyStyle(e.Key).Render(padRight(e.Key, keyW)))
		sb.WriteString(desc.Render(ansi.Truncate(e.Desc, descW, "…")))
		sb.WriteString("\n")
	}
	colStyle := lipgloss.NewStyle().Width(colWidth)
	if c.Theme.PaintBackground {
		colStyle = colStyle.Background(c.Theme.Bg)
	}
	return colStyle.Render(sb.String())
}

func padRight(s string, w int) string {
	n := lipgloss.Width(s)
	if n >= w {
		return s
	}
	return s + strings.Repeat(" ", w-n)
}
