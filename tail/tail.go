package tail

import (
	"image/color"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"github.com/blairham/tuikit/table"
	"github.com/blairham/tuikit/theme"
)

// Model wraps a bubbles viewport with a line buffer, a follow toggle,
// and a regex filter (via [table.RowFilter]).
//
// Apps feed lines in via [Model.AppendLine] as they arrive (e.g. from a
// background goroutine reading from a log shipper). The viewport
// auto-scrolls when follow mode is on; the filter hides non-matching
// lines without dropping them from the buffer, so toggling the filter
// off restores everything. Markers ([Model.AppendMarker]) pass every
// filter.
type Model struct {
	filter   table.RowFilter
	bgSeq    string
	lines    []line
	visible  []string
	viewport viewport.Model
	ready    bool
	follow   bool
	wrap     bool
	maxLines int
}

// line is one buffered line. A marker shows whatever the filter.
type line struct {
	text   string
	marker bool
}

// New constructs an empty tail model with follow=true.
func New() *Model {
	return &Model{follow: true}
}

// SetBackground makes [Model.View] paint the given background continuously
// behind the content. Formatted log lines contain resets (\x1b[m, and combined
// forms like \x1b[0;32m) that clear the background mid-line, so a surrounding width+background style only
// fills the leading/trailing padding and leaves the text on the terminal's own
// background. View re-asserts the background after each reset to close those
// gaps. Pass the theme's Bg.
func (m *Model) SetBackground(c color.Color) {
	m.bgSeq = theme.BackgroundSeq(c)
}

// Resize updates the underlying viewport dimensions.
func (m *Model) Resize(width, height int) {
	if !m.ready {
		m.viewport = viewport.New(
			viewport.WithWidth(width),
			viewport.WithHeight(height),
		)
		m.viewport.SoftWrap = m.wrap
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

// Wrap reports whether long lines soft-wrap.
func (m *Model) Wrap() bool { return m.wrap }

// SetWrap soft-wraps long lines onto the following rows instead of cutting
// them at the viewport's edge. It holds across the lazy viewport creation in
// [Model.Resize], so it can be set before the first resize.
func (m *Model) SetWrap(on bool) {
	m.wrap = on
	if m.ready {
		m.viewport.SoftWrap = on
		m.viewport.SetContent(strings.Join(m.visible, "\n"))
		if m.follow {
			m.viewport.GotoBottom()
		}
	}
}

// Clear empties the buffer. Lines appended afterwards fill it again, so a
// live stream keeps going from a blank view — k9s's ctrl-k.
func (m *Model) Clear() {
	m.lines = m.lines[:0]
	m.visible = m.visible[:0]
	if m.ready {
		m.viewport.SetContent("")
		m.viewport.GotoTop()
	}
}

// VisibleLines returns a copy of the filter-passing lines, as stored
// (styling included): what the view shows, for saving or copying it.
func (m *Model) VisibleLines() []string {
	return append([]string(nil), m.visible...)
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
func (m *Model) AppendLine(text string) {
	m.appendLines([]line{{text: text}})
}

// AppendMarker appends a line that every filter lets through — a separator
// such as k9s's log mark, a stamped rule at the current end of the stream.
// Filtering for the lines that matter after a mark is when the mark is
// needed most, so a marker stays visible, in its place in the buffer, and
// counts in [Model.VisibleLines] while shown. [Model.Clear] drops markers
// with everything else.
func (m *Model) AppendMarker(text string) {
	m.appendLines([]line{{text: text, marker: true}})
}

// shows reports whether l passes the current filter.
func (m *Model) shows(l line) bool {
	return l.marker || m.filter.Empty() || m.filter.MatchesAny(l.text)
}

// SetMaxLines caps the buffer at n lines: past it the oldest are dropped,
// markers included — a long-running stream otherwise grows without bound,
// and k9s caps a log view the same way (logger.buffer). A smaller cap trims
// the buffer at once. 0 keeps every line.
//
// While following, the view stays on the newest line. Scrolled back, it
// stays on the lines it shows: the offset moves up by the visible lines
// dropped above it. With wrap on that is by line, not by wrapped row, so a
// dropped line that wrapped can shift the view by a row.
func (m *Model) SetMaxLines(n int) {
	m.maxLines = max(n, 0)
	m.refresh(m.trim())
}

// MaxLines is the buffer cap; 0 is none.
func (m *Model) MaxLines() int { return m.maxLines }

// trim drops the oldest lines past the cap and returns how many of them
// were visible.
func (m *Model) trim() int {
	drop := len(m.lines) - m.maxLines
	if m.maxLines == 0 || drop <= 0 {
		return 0
	}
	dropped := 0
	for _, l := range m.lines[:drop] {
		if m.shows(l) {
			dropped++
		}
	}
	// Copied rather than re-sliced, so the dropped lines' memory goes too.
	m.lines = append([]line(nil), m.lines[drop:]...)
	m.visible = append([]string(nil), m.visible[dropped:]...)
	return dropped
}

// refresh redraws after a change that removed dropped visible lines from
// the top: pinned to the bottom when following, else moved up by them.
func (m *Model) refresh(dropped int) {
	if !m.ready {
		return
	}
	offset := m.viewport.YOffset()
	m.viewport.SetContent(strings.Join(m.visible, "\n"))
	if m.follow {
		m.viewport.GotoBottom()
	} else {
		m.viewport.SetYOffset(max(offset-dropped, 0))
	}
}

// appendLines buffers ls, trims to the cap, and refreshes the viewport once.
func (m *Model) appendLines(ls []line) {
	for _, l := range ls {
		m.lines = append(m.lines, l)
		if m.shows(l) {
			m.visible = append(m.visible, l.text)
		}
	}
	m.refresh(m.trim())
}

// ReplaceLines makes lines the whole buffer, for a view that refetches a
// document — an inspect pane — rather than streaming one. The filter, the
// follow state and the scroll offset carry over, so a refresh, manual or on
// a timer, keeps the reader's place: a following view stays at the bottom,
// any other at its offset, clamped when the new content is shorter.
// Markers are dropped with the old lines; the cap applies.
func (m *Model) ReplaceLines(lines []string) {
	m.lines = m.lines[:0]
	for _, text := range lines {
		m.lines = append(m.lines, line{text: text})
	}
	m.rebuildVisible()
	m.trim()
	if !m.ready {
		return
	}
	offset := m.viewport.YOffset()
	m.viewport.SetContent(strings.Join(m.visible, "\n"))
	if m.follow {
		m.viewport.GotoBottom()
	} else {
		m.viewport.SetYOffset(offset) // the viewport clamps it to the content
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
	ls := make([]line, len(lines))
	for i, text := range lines {
		ls[i] = line{text: text}
	}
	m.appendLines(ls)
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
	older := make([]line, 0, len(lines)+len(m.lines))
	for _, text := range lines {
		older = append(older, line{text: text})
	}
	m.lines = append(older, m.lines...)

	// Count how many prepended lines pass the filter, so the offset shifts by
	// exactly the number of rows inserted above the current view.
	added := len(lines)
	if !m.filter.Empty() {
		added = 0
		for _, text := range lines {
			if m.filter.MatchesAny(text) {
				added++
			}
		}
	}

	m.rebuildVisible()
	// Into a full buffer the newest lines are kept, so what does not fit is
	// the history just prepended.
	added -= m.trim()
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
// background re-asserted after every SGR that resets it so it stays continuous behind
// styled log lines.
func (m *Model) View() string {
	if !m.ready {
		return ""
	}
	out := m.viewport.View()
	if m.bgSeq != "" {
		out = theme.ReassertBackground(out, m.bgSeq)
	}
	return out
}

func (m *Model) rebuildVisible() {
	m.visible = m.visible[:0]
	for _, l := range m.lines {
		if m.shows(l) {
			m.visible = append(m.visible, l.text)
		}
	}
}
