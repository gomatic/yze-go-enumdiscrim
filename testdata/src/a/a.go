// Package a is the go-hx segment-kind shape: a const group whose if-shaped
// discrimination silently drops any member added later.
package a

// Kind discriminates segment kinds.
type Kind int

// The segment kinds.
const (
	KindDef Kind = iota
	KindRef
	KindRaw
)

// Mode is a TWO-member domain, where a two-way comparison is the whole story.
type Mode int

// The modes.
const (
	ModeRead Mode = iota
	ModeWrite
)

// Segment carries a kind.
type Segment struct {
	Kind Kind
}

// RenderIf discriminates the three-member group with an if — the reported
// shape: KindRaw added later falls through to the generic branch silently.
func RenderIf(segment Segment) string {
	if segment.Kind != KindDef { // want `discriminates Kind, a const group of 3 members`
		return "generic"
	}
	return "definition"
}

// RenderLoop steers a loop with the same shape.
func RenderLoop(segments []Segment) int {
	count := 0
	for i := 0; i < len(segments) && segments[i].Kind == KindRef; i++ { // want `discriminates Kind, a const group of 3 members`
		count++
	}
	return count
}

// RenderSwitch discriminates exhaustively — exhaustive's domain, not this
// probe's.
func RenderSwitch(segment Segment) string {
	switch segment.Kind {
	case KindDef:
		return "definition"
	case KindRef:
		return "reference"
	case KindRaw:
		return "raw"
	}
	return ""
}

// TwoWay compares within a two-member domain: the comparison IS exhaustive.
func TwoWay(mode Mode) string {
	if mode == ModeRead {
		return "read"
	}
	return "write"
}

// Bind stores the comparison without steering control flow; out of the
// probe's scope.
func Bind(segment Segment) bool {
	isDef := segment.Kind == KindDef
	return isDef
}

// Compare compares two VALUES of the enum type — no constant involved, no
// discrimination of the group.
func Compare(a, b Segment) bool {
	return a.Kind == b.Kind
}

// limit is a declared constant of a BASIC type — no named enum, no group.
const limit = 3

// Bounded compares against a basic-typed constant; nothing named, nothing
// reported.
func Bounded(n int) bool {
	if n == limit {
		return true
	}
	return false
}

// Fault is an error-implementing const group: comparing it against an error
// INTERFACE is not enum discrimination (the operands' types differ), in
// either operand order.
type Fault string

// The faults.
const (
	FaultA Fault = "a"
	FaultB Fault = "b"
	FaultC Fault = "c"
)

// Error renders the fault.
func (f Fault) Error() string { return string(f) }

// Faulted compares an interface against fault constants, both orders; a
// different rule's business, silent here.
func Faulted(err error) bool {
	if err == FaultA {
		return true
	}
	if FaultB == err {
		return true
	}
	return false
}

// SameKind steers on a VALUE-to-value comparison: no constant, no group
// discriminated.
func SameKind(a, b Segment) string {
	if a.Kind == b.Kind {
		return "same"
	}
	return "different"
}
