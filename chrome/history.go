// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package chrome

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
)

// HistoryLimit is the most entries a [CommandBar] or [FilterBar]
// remembers. Recording past it drops the oldest entry.
const HistoryLimit = 50

const (
	keyStrUp   = "up"
	keyStrDown = "down"
)

// inputHistory is the submitted-value memory shared by the command and
// filter bars: newest last, no consecutive duplicates, no empty values,
// at most [HistoryLimit] entries.
//
// pos is the entry currently recalled into the input, or -1 when the
// user is not recalling. draft is what the input held when recall
// began, restored when the user steps down past the newest entry.
type inputHistory struct {
	draft   string
	entries []string
	pos     int
}

func newInputHistory() inputHistory { return inputHistory{pos: -1} }

// add records v unless it is empty or repeats the newest entry, then
// trims to HistoryLimit. It does not touch the recall position.
func (h *inputHistory) add(v string) {
	if strings.TrimSpace(v) == "" {
		return
	}
	if n := len(h.entries); n > 0 && h.entries[n-1] == v {
		return
	}
	h.entries = append(h.entries, v)
	if over := len(h.entries) - HistoryLimit; over > 0 {
		h.entries = append([]string(nil), h.entries[over:]...)
	}
}

// set replaces the entries, applying the same rules as add to each
// value in order, and ends any recall in progress.
func (h *inputHistory) set(values []string) {
	h.entries = nil
	for _, v := range values {
		h.add(v)
	}
	h.reset()
}

// snapshot returns a copy of the entries, oldest first.
func (h *inputHistory) snapshot() []string {
	return append([]string{}, h.entries...)
}

// reset ends any recall in progress.
func (h *inputHistory) reset() {
	h.pos = -1
	h.draft = ""
}

// older steps one entry back. current is the input's value, saved as
// the draft when recall begins. ok is false when there is nothing older
// to show, and the input should be left alone.
func (h *inputHistory) older(current string) (value string, ok bool) {
	switch {
	case len(h.entries) == 0:
		return "", false
	case h.pos < 0:
		h.draft = current
		h.pos = len(h.entries) - 1
	case h.pos > 0:
		h.pos--
	default:
		return "", false
	}
	return h.entries[h.pos], true
}

// newer steps one entry forward; past the newest it ends the recall and
// hands back the draft. ok is false when the user is not recalling.
func (h *inputHistory) newer() (value string, ok bool) {
	if h.pos < 0 {
		return "", false
	}
	if h.pos < len(h.entries)-1 {
		h.pos++
		return h.entries[h.pos], true
	}
	draft := h.draft
	h.reset()
	return draft, true
}

// recall handles an up or down key press against in, and reports
// whether the input's value changed. Any other key is ignored.
func (h *inputHistory) recall(in *textinput.Model, keyStr string) (changed bool) {
	var value string
	switch keyStr {
	case keyStrUp:
		value, changed = h.older(in.Value())
	case keyStrDown:
		value, changed = h.newer()
	}
	if changed {
		in.SetValue(value)
		in.CursorEnd()
	}
	return changed
}
