package a

import "testing"

// TestRender covers the fixtures.
func TestRender(t *testing.T) {
	if RenderIf(Segment{Kind: KindDef}) != "definition" {
		t.Fatal("definition")
	}
	if RenderSwitch(Segment{Kind: KindRaw}) != "raw" {
		t.Fatal("raw")
	}
	if TwoWay(ModeRead) != "read" {
		t.Fatal("read")
	}
	if !Bind(Segment{Kind: KindDef}) || Compare(Segment{}, Segment{Kind: KindRef}) {
		t.Fatal("bind/compare")
	}
	if RenderLoop([]Segment{{Kind: KindRef}}) != 1 {
		t.Fatal("loop")
	}
	if !Bounded(3) || Bounded(4) {
		t.Fatal("bounded")
	}
	if Faulted(nil) || SameKind(Segment{}, Segment{}) != "same" {
		t.Fatal("faulted/samekind")
	}
}
