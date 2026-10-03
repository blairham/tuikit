// SPDX-FileCopyrightText: 2026 Blair Hamilton
// SPDX-License-Identifier: Apache-2.0

package viewfsm

import "testing"

const (
	vA ViewID = iota
	vB
	vC
	vD
)

func expect(t *testing.T, what string, id ViewID, ok bool, want ViewID, wantOK bool) {
	t.Helper()
	if ok != wantOK || (ok && id != want) {
		t.Errorf("%s = (%v, %v), want (%v, %v)", what, id, ok, want, wantOK)
	}
}

// TestHistoryBrowserSemantics: back and forward walk the trail without
// recording, a visit after going back drops the forward entries, and a
// repeat visit to the current view is not recorded twice.
func TestHistoryBrowserSemantics(t *testing.T) {
	var h History // zero value is usable
	_, ok := h.Back()
	expect(t, "Back on empty", 0, ok, 0, false)

	for _, v := range []ViewID{vA, vB, vB, vC} {
		h.Visit(v)
	}
	id, ok := h.Back()
	expect(t, "Back from C", id, ok, vB, true)
	id, ok = h.Back()
	expect(t, "Back from B (the repeat was collapsed)", id, ok, vA, true)
	id, ok = h.Back()
	expect(t, "Back at oldest", id, ok, 0, false)
	id, ok = h.Forward()
	expect(t, "Forward", id, ok, vB, true)

	h.Visit(vD) // from B: drops C
	id, ok = h.Forward()
	expect(t, "Forward after a new visit", id, ok, 0, false)
	id, ok = h.Back()
	expect(t, "Back from D", id, ok, vB, true)
}

// TestHistoryLastToggles: `-` returns to the previous view, and pressed
// again comes back — the two most recent views alternate.
func TestHistoryLastToggles(t *testing.T) {
	h := NewHistory(0)
	_, ok := h.Last()
	expect(t, "Last on empty", 0, ok, 0, false)
	h.Visit(vA)
	h.Visit(vB)
	for i, want := range []ViewID{vA, vB, vA} {
		id, ok := h.Last()
		expect(t, "Last #"+string(rune('1'+i)), id, ok, want, true)
	}
}

func TestHistoryIsBounded(t *testing.T) {
	h := NewHistory(3)
	for _, v := range []ViewID{vA, vB, vC, vD} {
		h.Visit(v)
	}
	var back []ViewID
	for {
		id, ok := h.Back()
		if !ok {
			break
		}
		back = append(back, id)
	}
	if len(back) != 2 || back[0] != vC || back[1] != vB {
		t.Errorf("bounded history walked back through %v, want [C B] (A dropped)", back)
	}
}

// TestRouterRecordsJumps: JumpTo records, history moves change the active
// view and the stack without recording, and an unknown ID records nothing.
func TestRouterRecordsJumps(t *testing.T) {
	r := NewRouter(map[ViewID]Spec{vA: {Name: "a"}, vB: {Name: "b"}, vC: {Name: "c"}}, vA)
	r.JumpTo(vB)
	r.Push(vC) // drill-ins are not history
	r.JumpTo(ViewID(99))
	id, ok := r.HistoryBack()
	expect(t, "HistoryBack", id, ok, vA, true)
	if r.Active() != vA || !r.IsAtRoot() {
		t.Errorf("after HistoryBack active %v root %v", r.Active(), r.IsAtRoot())
	}
	id, ok = r.HistoryForward()
	expect(t, "HistoryForward", id, ok, vB, true)
	id, ok = r.LastView()
	expect(t, "LastView", id, ok, vA, true)
	id, ok = r.LastView()
	expect(t, "LastView again", id, ok, vB, true)
	if r.Active() != vB {
		t.Errorf("active = %v, want B", r.Active())
	}
}
