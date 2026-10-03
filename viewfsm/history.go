// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package viewfsm

// DefaultHistorySize bounds a [History] whose size was not set.
const DefaultHistorySize = 50

// History is the visited-view trail behind k9s's `[` (back), `]` (forward)
// and `-` (last used view). It has browser semantics: visiting a view after
// going back drops the forward entries, a visit to the view already current
// is not recorded twice, and the oldest entries fall off past the size.
//
// It is independent of [Router] so an app that keeps its own view stack
// can still use it; [Router] records into one of its own on every
// [Router.JumpTo]. The zero value is ready to use.
type History struct {
	entries []ViewID
	pos     int
	size    int
}

// NewHistory returns a History holding at most size entries; size <= 0
// means [DefaultHistorySize].
func NewHistory(size int) *History { return &History{size: size} }

func (h *History) limit() int {
	if h.size <= 0 {
		return DefaultHistorySize
	}
	return h.size
}

// Visit records id as the current view.
func (h *History) Visit(id ViewID) {
	if len(h.entries) > 0 {
		if h.entries[h.pos] == id {
			return
		}
		h.entries = h.entries[:h.pos+1]
	}
	h.entries = append(h.entries, id)
	if over := len(h.entries) - h.limit(); over > 0 {
		h.entries = h.entries[over:]
	}
	h.pos = len(h.entries) - 1
}

// Current is the view the history is positioned on.
func (h *History) Current() (ViewID, bool) {
	if len(h.entries) == 0 {
		return 0, false
	}
	return h.entries[h.pos], true
}

// Back steps to the previous view without recording, for `[`. It reports
// false at the oldest entry.
func (h *History) Back() (ViewID, bool) {
	if h.pos == 0 || len(h.entries) == 0 {
		return 0, false
	}
	h.pos--
	return h.entries[h.pos], true
}

// Forward steps to the next view after a Back, for `]`. It reports false
// at the newest entry.
func (h *History) Forward() (ViewID, bool) {
	if h.pos >= len(h.entries)-1 {
		return 0, false
	}
	h.pos++
	return h.entries[h.pos], true
}

// Last returns to the view before the current one and records that as a
// visit, so pressing `-` twice toggles between the two most recent views —
// k9s's "last used command". It reports false when there is nothing before.
func (h *History) Last() (ViewID, bool) {
	if h.pos == 0 || len(h.entries) == 0 {
		return 0, false
	}
	id := h.entries[h.pos-1]
	h.Visit(id)
	return id, true
}
