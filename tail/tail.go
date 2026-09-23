package tail

import (
	"image/color"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/blairham/tuikit/table"
)

// Model wraps a bubbles viewport with a line buffer, a follow toggle,
// and a regex filter (via [table.RowFilter]).
//
// Apps feed lines in via [Model.AppendLine] as they arrive (e.g. from a
// background goroutine reading from a log shipper). The viewport
// auto-scrolls when follow mode is on; the filter hides non-matching
// lines without dropping them from the buffer, so toggling the filter
// off restores everything.
type Model struct {
	filter   table.RowFilter
	bgSeq    string
	lines    []string
	visible  []string
	viewport viewport.Model
	ready    bool
	follow   bool
}

// New constructs an empty tail model with follow=true.
func New() *Model {
	return &Model{follow: true}
}

// SetBackground makes [Model.View] paint the given background continuously
// behind the content. Formatted log lines contain reset codes (\x1b[m) that
// clear the background mid-line, so a surrounding width+background style only
// fills the leading/trailing padding and leaves the text on the terminal's own
// background. View re-asserts the background after each reset to close those
// gaps. Pass the theme's Bg.
func (m *Model) SetBackground(c color.Color) {
	// Derive the raw background SGR from lipgloss so it honors the active color
	// profile (truecolor / 256 / none) instead of hardcoding an escape.
	sample := lipgloss.NewStyle().Background(c).Render(" ")
	if i := strings.IndexByte(sample, 'm'); strings.HasPrefix(sample, "\x1b[") && i > 0 {
		m.bgSeq = sample[:i+1]
	}
}

// Resize updates the underlying viewport dimensions.
func (m *Model) Resize(width, height int) {
	if !m.ready {
		m.viewport = viewport.New(
			viewport.WithWidth(width),
			viewport.WithHeight(height),
		)
		m.ready = true
	} else {
		m.viewport.SetWidth(width)
		m.viewport.SetHeight(height)
	}
	m.viewport.SetContent(strings.Join(m.visible, "\n"))
	if m.follow {
		m.viewport.GotoBottom()
	}
}

// Ready reports whether the viewport has been initialized at least once.
func (m *Model) Ready() bool { return m.ready }

// Follow returns whether auto-scroll is active.
func (m *Model) Follow() bool { return m.follow }

// SetFollow toggles auto-scroll.
func (m *Model) SetFollow(on bool) {
	m.follow = on
	if on && m.ready {
		m.viewport.GotoBottom()
	}
}

// LineCount returns the number of buffered lines (pre-filter).
func (m *Model) LineCount() int { return len(m.lines) }

// VisibleCount returns the number of filter-passing lines.
func (m *Model) VisibleCount() int { return len(m.visible) }

// SetFilter applies a new filter and rebuilds the visible buffer.
func (m *Model) SetFilter(expr string) {
	m.filter = table.ParseFilter(expr)
	m.rebuildVisible()
	if m.ready {
		m.viewport.SetContent(strings.Join(m.visible, "\n"))
		if m.follow {
			m.viewport.GotoBottom()
		}
	}
}

// AppendLine adds a new raw line to the buffer. If the filter is active
// and the line doesn't match, it's still kept in [Model.lines] but
// excluded from the visible view.
func (m *Model) AppendLine(line string) {
	m.lines = append(m.lines, line)
	if m.filter.Empty() || m.filter.MatchesAny(line) {
		m.visible = append(m.visible, line)
	}
	if m.ready {
		m.viewport.SetContent(strings.Join(m.visible, "\n"))
		if m.follow {
			m.viewport.GotoBottom()
		}
	}
}

// AppendLines appends a batch of lines in a single update — one content
// rebuild and at most one auto-scroll, instead of repeating that work per line
// as [Model.AppendLine] does. Filter handling matches AppendLine. Apps feeding
// a backfill or a multi-line page should prefer this so the view renders in one
// frame instead of scrolling line-by-line.
func (m *Model) AppendLines(lines []string) {
	if len(lines) == 0 {
		return
	}
	for _, line := range lines {
		m.lines = append(m.lines, line)
		if m.filter.Empty() || m.filter.MatchesAny(line) {
			m.visible = append(m.visible, line)
		}
	}
	if m.ready {
		m.viewport.SetContent(strings.Join(m.visible, "\n"))
		if m.follow {
			m.viewport.GotoBottom()
		}
	}
}

// PrependLines inserts older lines at the top of the buffer while preserving
// the user's current view: the line they were looking at stays in place by
// shifting the viewport offset down by the number of newly-visible lines. When
// following, the view stays pinned to the bottom. This is the building block
// for lazy "scroll up to load older history" paging — pair it with
// [Model.AtTop] to decide when to fetch.
func (m *Model) PrependLines(lines []string) {
	if len(lines) == 0 {
		return
	}
	m.lines = append(append([]string(nil), lines...), m.lines...)

	// Count how many prepended lines pass the filter, so the offset shifts by
	// exactly the number of rows inserted above the current view.
	added := len(lines)
	if !m.filter.Empty() {
		added = 0
		for _, line := range lines {
			if m.filter.MatchesAny(line) {
				added++
			}
		}
	}

	m.rebuildVisible()
	if !m.ready {
		return
	}
	offset := m.viewport.YOffset()
	m.viewport.SetContent(strings.Join(m.visible, "\n"))
	if m.follow {
		m.viewport.GotoBottom()
	} else {
		m.viewport.SetYOffset(offset + added)
	}
}

// AtTop reports whether the viewport is scrolled to the very top. Apps use this
// as the cue to fetch older history and feed it to [Model.PrependLines].
func (m *Model) AtTop() bool {
	return m.ready && m.viewport.AtTop()
}

// Update forwards a tea.Msg into the underlying viewport.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return cmd
}

// HandleScrollKey reacts to navigation keys and reports whether the key
// was handled. Apps call this before falling through to their default
// table-update dispatch.
//
// Bindings:
//
//	ctrl+f / pgdown  half-page down (pauses follow)
//	ctrl+b / pgup    half-page up   (pauses follow)
//	ctrl+d           full-page down (pauses follow)
//	ctrl+u           full-page up   (pauses follow)
//	g / home         goto top       (pauses follow)
//	G / end          goto bottom    (enables follow)
//	j / down         one line down  (pauses follow)
//	k / up           one line up    (pauses follow)
//	f                toggle follow
func (m *Model) HandleScrollKey(key string) bool {
	if !m.ready {
		return false
	}
	switch key {
	case "ctrl+f", "pgdown":
		m.follow = false
		m.viewport.HalfPageDown()
	case "ctrl+b", "pgup":
		m.follow = false
		m.viewport.HalfPageUp()
	case "ctrl+d":
		m.follow = false
		m.viewport.PageDown()
	case "ctrl+u":
		m.follow = false
		m.viewport.PageUp()
	case "g", "home":
		m.follow = false
		m.viewport.GotoTop()
	case "G", "end":
		m.follow = true
		m.viewport.GotoBottom()
	case "j", "down":
		m.follow = false
		m.viewport.ScrollDown(1)
	case "k", "up":
		m.follow = false
		m.viewport.ScrollUp(1)
	case "f":
		m.follow = !m.follow
		if m.follow {
			m.viewport.GotoBottom()
		}
	default:
		return false
	}
	return true
}

// View renders the viewport. Returns an empty string if Resize() has
// not been called yet — callers should fall back to a placeholder. When a
// background is set (see [Model.SetBackground]) the rendered content has the
// background re-asserted after every reset code so it stays continuous behind
// styled log lines.
func (m *Model) View() string {
	if !m.ready {
		return ""
	}
	out := m.viewport.View()
	if m.bgSeq != "" {
		out = fillBackground(out, m.bgSeq)
	}
	return out
}

// fillBackground re-asserts the background SGR immediately after every reset in
// s. lipgloss does not restore a surrounding background after the resets that
// terminate inner styled spans, so without this the text between/after styled
// tokens falls back to the terminal's own background. The leading and trailing
// padding are handled by the caller's width+background wrapper (e.g. the
// bordered content box).
func fillBackground(s, bgSeq string) string {
	s = strings.ReplaceAll(s, "\x1b[0m", "\x1b[0m"+bgSeq)
	s = strings.ReplaceAll(s, "\x1b[m", "\x1b[m"+bgSeq)
	return s
}

func (m *Model) rebuildVisible() {
	m.visible = m.visible[:0]
	if m.filter.Empty() {
		m.visible = append(m.visible, m.lines...)
		return
	}
	for _, l := range m.lines {
		if m.filter.MatchesAny(l) {
			m.visible = append(m.visible, l)
		}
	}
}
