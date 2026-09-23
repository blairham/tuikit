package viewfsm

import "testing"

const (
	vRoot ViewID = iota
	vChild
	vGrandchild
	vSibling
)

func newTestRouter() *Router {
	return NewRouter(map[ViewID]Spec{
		vRoot:       {Name: "Root", Hotkey: "1"},
		vChild:      {Name: "Child", Hotkey: "2"},
		vGrandchild: {Name: "Grandchild"},
		vSibling:    {Name: "Sibling", Hotkey: "3"},
	}, vRoot)
}

func TestRouter_InitialState(t *testing.T) {
	t.Parallel()
	r := newTestRouter()
	if r.Active() != vRoot {
		t.Errorf("Active() = %d; want vRoot", r.Active())
	}
	if !r.IsAtRoot() {
		t.Error("expected IsAtRoot to be true on fresh router")
	}
	if len(r.Stack()) != 1 {
		t.Errorf("Stack length = %d; want 1", len(r.Stack()))
	}
}

func TestRouter_PushPop(t *testing.T) {
	t.Parallel()
	r := newTestRouter()
	r.Push(vChild)
	if r.Active() != vChild {
		t.Errorf("after Push, Active() = %d; want vChild", r.Active())
	}
	if r.IsAtRoot() {
		t.Error("after Push, expected IsAtRoot=false")
	}

	r.Push(vGrandchild)
	if len(r.Stack()) != 3 {
		t.Errorf("Stack length after two pushes = %d; want 3", len(r.Stack()))
	}

	if !r.Pop() {
		t.Error("Pop should return true when not at root")
	}
	if r.Active() != vChild {
		t.Errorf("after Pop, Active() = %d; want vChild", r.Active())
	}

	r.Pop()
	if !r.IsAtRoot() {
		t.Error("after two pops, expected IsAtRoot=true")
	}
	if r.Pop() {
		t.Error("Pop at root should return false")
	}
}

func TestRouter_PushUnknownIsNoop(t *testing.T) {
	t.Parallel()
	r := newTestRouter()
	r.Push(ViewID(99)) // unknown
	if r.Active() != vRoot {
		t.Error("Push of unknown view should be no-op")
	}
}

func TestRouter_ResolveHotkey(t *testing.T) {
	t.Parallel()
	r := newTestRouter()
	id, ok := r.ResolveHotkey("2")
	if !ok || id != vChild {
		t.Errorf("ResolveHotkey(\"2\") = (%d, %v); want (vChild, true)", id, ok)
	}
	if _, ok := r.ResolveHotkey("9"); ok {
		t.Error("ResolveHotkey on unbound key should return false")
	}
}

func TestRouter_JumpToResetsStack(t *testing.T) {
	t.Parallel()
	r := newTestRouter()
	r.Push(vChild)
	r.Push(vGrandchild)
	r.JumpTo(vSibling)
	if r.Active() != vSibling {
		t.Errorf("after JumpTo, Active() = %d; want vSibling", r.Active())
	}
	if len(r.Stack()) != 1 {
		t.Errorf("after JumpTo, Stack length = %d; want 1", len(r.Stack()))
	}
}

func TestRouter_Replace(t *testing.T) {
	t.Parallel()
	r := newTestRouter()
	r.Push(vChild)
	r.Replace(vSibling)
	if r.Active() != vSibling {
		t.Errorf("after Replace, Active() = %d; want vSibling", r.Active())
	}
	if len(r.Stack()) != 2 {
		t.Errorf("Replace should preserve depth, got %d", len(r.Stack()))
	}
}

func TestRouter_Breadcrumb(t *testing.T) {
	t.Parallel()
	r := newTestRouter()
	r.Push(vChild)
	r.Push(vGrandchild)
	crumbs := r.Breadcrumb()
	if len(crumbs) != 3 {
		t.Fatalf("Breadcrumb len = %d; want 3", len(crumbs))
	}
	wantLabels := []string{"Root", "Child", "Grandchild"}
	for i, c := range crumbs {
		if c.Label != wantLabels[i] {
			t.Errorf("crumb[%d].Label = %q; want %q", i, c.Label, wantLabels[i])
		}
		wantLeaf := i == 2
		if c.Leaf != wantLeaf {
			t.Errorf("crumb[%d].Leaf = %v; want %v", i, c.Leaf, wantLeaf)
		}
	}
}
