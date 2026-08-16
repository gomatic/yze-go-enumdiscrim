// Package a is the go-hx segment-kind shape: a const group whose if-shaped
// discrimination silently drops any member added later.
package a

import "b"

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

// modeDefault NAMES an existing mode rather than adding one: Mode still has
// two inhabitants, so TwoWay above stays a two-way domain and stays silent.
const modeDefault Mode = ModeRead

// Defaulted steers on the named default; the domain is still two-way.
func Defaulted(mode Mode) string {
	if mode == modeDefault {
		return "read"
	}
	return "write"
}

// KindAlias spells Kind by another name; the alias resolves to the same group.
type KindAlias = Kind

// Aliased discriminates the three-member group through the alias spelling.
func Aliased(kind KindAlias) string {
	if kind != KindDef { // want `discriminates Kind, a const group of 3 members`
		return "generic"
	}
	return "definition"
}

// Parenthesised discriminates with both operands parenthesised.
func Parenthesised(kind Kind) string {
	if (kind) != (KindDef) { // want `discriminates Kind, a const group of 3 members`
		return "generic"
	}
	return "definition"
}

// LeftHanded puts the constant on the LEFT of the comparison.
func LeftHanded(kind Kind) string {
	if KindDef != kind { // want `discriminates Kind, a const group of 3 members`
		return "generic"
	}
	return "definition"
}

// anyKind reports whether any segment satisfies the predicate; the predicate's
// body is a value the caller's condition is computed from.
func anyKind(segments []Segment, matches func(Kind) bool) bool {
	for _, segment := range segments {
		if matches(segment.Kind) {
			return true
		}
	}
	return false
}

// takes reports its argument; a comparison passed to it is an argument, not a
// test the condition performs.
func takes(flag bool) bool { return flag }

// NotSteering holds three comparisons that appear inside the condition without
// steering it — a closure body, a call argument and a map index — beside a
// sibling that does steer and IS reported.
func NotSteering(segments []Segment, kind Kind, byKind map[bool]bool) string {
	if anyKind(segments, func(k Kind) bool { return k == KindDef }) {
		return "closure"
	}
	if takes(kind == KindDef) {
		return "argument"
	}
	if byKind[kind == KindDef] {
		return "index"
	}
	if kind == KindRef { // want `discriminates Kind, a const group of 3 members`
		return "steering"
	}
	return "generic"
}

// Nested puts the same comparison inside an inner if whose enclosing condition
// is a call: the inner if is a statement of its own and is reported exactly
// once, never a second time for lexically sitting inside the outer condition.
func Nested(kind Kind) string {
	if takes(func() bool {
		if kind == KindDef { // want `discriminates Kind, a const group of 3 members`
			return true
		}
		return false
	}()) {
		return "nested"
	}
	return "generic"
}

// ChainExhaustive names every member of the group across its arms and ends in a
// terminal fallback: nothing falls through, silently or otherwise.
func ChainExhaustive(kind Kind) string {
	if kind == KindDef {
		return "definition"
	} else if kind == KindRef {
		return "reference"
	} else if kind == KindRaw {
		return "raw"
	}
	panic("unknown kind")
}

// ChainPartial leaves KindRaw unnamed, so KindRaw falls through — reported ONCE
// at the head of the chain rather than once per arm.
func ChainPartial(kind Kind) string {
	if kind == KindDef { // want `discriminates Kind, a const group of 3 members`
		return "definition"
	} else if kind == KindRef {
		return "reference"
	}
	return "generic"
}

// Disjunction names two members in ONE condition: one construct, one decision,
// one finding.
func Disjunction(kind Kind) string {
	if kind == KindDef || kind == KindRef { // want `discriminates Kind, a const group of 3 members`
		return "named"
	}
	return "generic"
}

// Negated tests through a ! and is still a test.
func Negated(kind Kind) string {
	if !(kind == KindDef) { // want `discriminates Kind, a const group of 3 members`
		return "generic"
	}
	return "definition"
}

// ValueSpelled writes the member's VALUE instead of its name — the same
// discrimination spelled worse, and reported alike.
func ValueSpelled(kind Kind) string {
	if kind != 0 { // want `discriminates Kind, a const group of 3 members`
		return "generic"
	}
	return "definition"
}

// ConvertedSpelled converts the value to the type — same inhabitant again.
func ConvertedSpelled(kind Kind) string {
	if kind != Kind(0) { // want `discriminates Kind, a const group of 3 members`
		return "generic"
	}
	return "definition"
}

// OffGroup compares against a value the group does not declare: no inhabitant
// is named, so no member of the group is discriminated and nothing falls
// through the comparison that the group could have been asked about.
func OffGroup(kind Kind) string {
	if kind != Kind(7) {
		return "generic"
	}
	return "impossible"
}

// InitBound binds the comparison in the if's OWN init statement — the stated
// scope limitation, silent, beside the identical condition-borne sibling in
// RenderIf above which is reported.
func InitBound(kind Kind) string {
	if isDef := kind == KindDef; isDef {
		return "definition"
	}
	return "generic"
}

// Tagless discriminates in a tagless switch — out of scope here, and out of
// exhaustive's scope too.
func Tagless(kind Kind) string {
	switch {
	case kind == KindDef:
		return "definition"
	}
	return "generic"
}

// Foreign discriminates a const group declared in ANOTHER package: b.Tag has
// three inhabitants and the comparison steers the if, so every other clause of
// the rule is satisfied and only the module-locality refusal is holding it
// silent. Delete the localToModule call in record and this line reports; it is
// the same three-member shape as RenderIf above, which does report, so the
// pair is the boundary in both directions.
func Foreign(tag b.Tag) string {
	if tag != b.TagOne {
		return "generic"
	}
	return "one"
}

// ForeignChain leaves the same foreign group partly named across a chain — the
// reported shape at ChainPartial, silent here for the one reason that differs.
func ForeignChain(tag b.Tag) string {
	if tag == b.TagOne {
		return "one"
	} else if tag == b.TagTwo {
		return "two"
	}
	return "generic"
}

// Own is declared here and discriminated here: the positive half of the same
// boundary, so a mutant that refuses every group — not merely the foreign ones
// — takes this with it.
type Own int

// The own-module inhabitants.
const (
	OwnA Own = iota
	OwnB
	OwnC
)

// Owned discriminates the module's OWN group and is reported.
func Owned(own Own) string {
	if own != OwnA { // want `discriminates Own, a const group of 3 members`
		return "generic"
	}
	return "a"
}
