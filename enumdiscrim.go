// Package enumdiscrim provides a go/analysis probe for the discrimination
// shape the exhaustive linter is blind to: `if x != EnumConst` (or ==) where
// EnumConst belongs to a const group with more than two members. A switch
// with a missed member is exhaustive's finding; the same defect wearing an
// if-statement — the shape that silently rendered an unknown segment kind as
// raw text in go-hx — is what this reports.
//
// This is a PROBE, not a gate. A fast-path check (`if kind != Special {
// return generic }`) is a legitimate two-way split of a many-way domain, and
// only the author knows which reading is intended. To bound the noise:
//
//   - the const group is DECLARED INSIDE THE ANALYZED MODULE. A foreign
//     package's constants are not a domain this author governs: the group can
//     gain a member from a Go release without the analyzed code changing, so
//     the unhandled count is a number about somebody else's library and the
//     prescribed switch is one the author cannot write. Every if-shaped
//     comparison against a stdlib tag type in this suite is that shape — `tok
//     == token.GOTO` asks "is this a goto?", a two-way question, over 90 token
//     values — and answering it exhaustively would be worse code, not better;
//   - BOTH operands share the constant's named type (comparing an error
//     interface against a sentinel constant is a different rule's business);
//   - the comparison STEERS an if or for condition — it is the condition, or
//     an operand of the &&, || and ! that combine it. A comparison appearing
//     inside a call argument, a closure body or a map index is a value the
//     condition is computed FROM and is not reported;
//   - the const group has more than two INHABITANTS, counted as distinct
//     constant values: a second name for an existing value (a default, a
//     bound, a sentinel alias) adds no member to the domain;
//   - an if/else-if chain is judged ONCE, at its head, over every member any
//     arm names. A chain naming every member of the group discriminates the
//     whole domain and is silent — a member added later reappears as a
//     finding, which is the moment the defect actually exists.
//
// The constant operand is judged by its VALUE, not by its spelling: `k != 0`
// and `k != Kind(0)` name the same inhabitant as `k != KindA` and are reported
// alike, so rewriting a member as its value buys silence and worse code. It
// must BE an inhabitant — a comparison against a value the group does not
// declare (a bitmask test, a duration, a zero value no member carries)
// discriminates nothing in the group and is not reported.
//
// MODULE LOCALITY IS A DISABLEMENT CHANNEL and is recorded as one rather than
// described as a virtue. The group's home is the package that DECLARES the
// type, so an author can silence a real finding by moving the const group out
// of the analyzed module — a nested go.mod and a `replace` directive do that
// without leaving the repository. What it costs is a module boundary: a go.mod,
// a require line and a dependency-manifest entry, each of which a reviewer and
// an inventory can read. The honest remedy costs an exhaustive switch, which is
// less work than the escape, so the cheapest silence here is the fix. What does
// NOT buy silence: an alias (`type K = pkg.Kind` resolves to the declaring
// package through types.Unalias), a defined type whose members the author
// declares (they are declared in the author's package, which is local), and
// moving the group to any other package of the same module (a module path is
// matched as a whole path element, so example.test/mod does not swallow
// example.test/modular).
//
// SCOPE LIMITATIONS, stated so silence is never mistaken for the analyzer
// being broken. The comparison must be the condition itself. A comparison
// bound in the if's own init statement (`if isA := k == KindA; isA`), a
// comparison hoisted to a preceding assignment, and a tagless switch
// (`switch { case k == KindA: }`) are all out of scope: each takes a dataflow
// or a statement form this rule does not follow. The tagless switch is out of
// github.com/nishanths/exhaustive's scope too, so that shape is seen by
// nothing in the suite.
package enumdiscrim

import (
	"go/ast"
	"go/token"
	"go/types"

	goyze "github.com/gomatic/go-yze"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// message is the diagnostic for an if-shaped enum discrimination. It names how
// many members go unhandled, because that number is what the author has to
// answer and it is what changes when a member is added later.
const message = "==/!= discriminates %s, a const group of %d members; %d of them are never named here and fall through silently — prefer an exhaustive switch"

// minimumMembers is the const-group size above which a two-way comparison is
// worth flagging: two members ARE a two-way domain.
const minimumMembers = 2

// Analyzer reports if-shaped discrimination of multi-member const groups.
var Analyzer = &analysis.Analyzer{
	Name:     "enumdiscrim",
	Doc:      "reports ==/!= control-flow discrimination of an analyzed-module const group with more than two members, the shape exhaustive cannot see",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// Registration declares this analyzer to the yze framework.
var Registration = goyze.Registration{
	Name:       "enumdiscrim",
	Categories: []goyze.Category{"correctness"},
	URL:        "https://docs.gomatic.dev/yze/enumdiscrim",
	Analyzer:   Analyzer,
}

// run judges every if/else-if chain and every for condition once.
func run(pass *analysis.Pass) (any, error) {
	ins, _ := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	arms := make(map[ast.Node]bool)
	ins.Preorder([]ast.Node{(*ast.IfStmt)(nil), (*ast.ForStmt)(nil)}, func(n ast.Node) {
		if !arms[n] {
			report(pass, discriminations(pass, steered(n, arms)))
		}
	})
	return nil, nil
}

// steered returns the conditions one statement steers on: a for's condition, or
// every condition in an if/else-if chain. Each else-if is recorded in arms so
// the chain is judged at its head instead of once per arm — one construct an
// author reviews as a whole is one finding, and a chain's arity is not a count
// of the decisions its author made.
func steered(n ast.Node, arms map[ast.Node]bool) []ast.Expr {
	if loop, ok := n.(*ast.ForStmt); ok {
		return nonNil(loop.Cond)
	}
	conditions := []ast.Expr{}
	for stmt, ok := n.(*ast.IfStmt); ok; stmt, ok = n.(*ast.IfStmt) {
		conditions = append(conditions, stmt.Cond)
		if stmt.Else == nil {
			break
		}
		arms[stmt.Else] = true
		n = stmt.Else
	}
	return conditions
}

// nonNil wraps an optional condition as the zero or one conditions it is.
func nonNil(cond ast.Expr) []ast.Expr {
	if cond == nil {
		return nil
	}
	return []ast.Expr{cond}
}

// discrimination accumulates what one chain says about one const group: the
// group, where to report it, how large it is, and which inhabitants any arm
// names.
type discrimination struct {
	group   *types.Named
	members map[string]struct{}
	at      token.Pos
	total   int
}

// discriminations groups every steering comparison in the chain by the const
// group it discriminates, in source order so the report is stable.
func discriminations(pass *analysis.Pass, conditions []ast.Expr) []*discrimination {
	found := []*discrimination{}
	for _, cond := range conditions {
		steering(pass, cond, func(cmp *ast.BinaryExpr) {
			found = record(pass, found, cmp)
		})
	}
	return found
}

// record adds the inhabitant cmp names to its group, opening the group on first
// sight at cmp's position. A comparison against a value the group does not
// declare names no inhabitant and opens nothing, and a group declared outside
// the analyzed module is refused before it is even counted — its size is a fact
// about a library, not about a domain this author can close.
func record(pass *analysis.Pass, found []*discrimination, cmp *ast.BinaryExpr) []*discrimination {
	named, value, ok := comparedEnum(pass, cmp)
	if !ok || !localToModule(pass, named.Obj().Pkg()) {
		return found
	}
	if opened := lookup(found, named); opened != nil {
		opened.members[value] = struct{}{}
		return found
	}
	members := memberValues(named)
	if _, inhabits := members[value]; !inhabits {
		return found
	}
	opened := &discrimination{group: named, members: map[string]struct{}{value: {}}, at: cmp.Pos(), total: len(members)}
	return append(found, opened)
}

// lookup returns the accumulator already open for the named group, or nil.
func lookup(found []*discrimination, named *types.Named) *discrimination {
	for _, opened := range found {
		if types.Identical(opened.group, named) {
			return opened
		}
	}
	return nil
}

// report emits one diagnostic per group the chain leaves incompletely handled.
// A group every arm between them names is the whole domain discriminated, so
// nothing falls through and nothing is reported.
func report(pass *analysis.Pass, found []*discrimination) {
	for _, opened := range found {
		unhandled := opened.total - len(opened.members)
		if opened.total > minimumMembers && unhandled > 0 {
			pass.Reportf(opened.at, message, opened.group.Obj().Name(), opened.total, unhandled)
		}
	}
}

// steering visits every comparison that steers the condition. It descends
// through parentheses and through the operators that compose tests into a test
// — and through nothing else, which is what bounds the rule to what the doc
// comment claims: a comparison inside a call argument, a closure body or a map
// index is a value the condition is computed from, not a test it performs.
func steering(pass *analysis.Pass, expr ast.Expr, visit func(*ast.BinaryExpr)) {
	comparison, operands := parts(pass, ast.Unparen(expr))
	if comparison != nil {
		visit(comparison)
		return
	}
	for _, operand := range operands {
		steering(pass, operand, visit)
	}
}

// parts splits a boolean expression into the equality comparison it IS, or the
// operands it composes one from.
//
// Composition is read off the OPERAND TYPES rather than enumerated from the
// operator table: an operator composes tests exactly when its operands are
// themselves boolean, which in Go is !, && and || and nothing else, since no
// other operator accepts a bool. Arithmetic, a call, a closure and an index all
// compute a value, and a comparison inside a value is not a test.
func parts(pass *analysis.Pass, expr ast.Expr) (*ast.BinaryExpr, []ast.Expr) {
	switch e := expr.(type) {
	case *ast.UnaryExpr:
		return nil, booleanOperands(pass, e.X)
	case *ast.BinaryExpr:
		if e.Op == token.EQL || e.Op == token.NEQ {
			return e, nil
		}
		return nil, booleanOperands(pass, e.X, e.Y)
	}
	return nil, nil
}

// booleanOperands returns the operands when every one of them is boolean — the
// mark of an operator that composes tests — and none otherwise.
func booleanOperands(pass *analysis.Pass, operands ...ast.Expr) []ast.Expr {
	for _, operand := range operands {
		basic, ok := pass.TypesInfo.TypeOf(operand).Underlying().(*types.Basic)
		if !ok || basic.Info()&types.IsBoolean == 0 {
			return nil
		}
	}
	return operands
}
