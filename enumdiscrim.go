// Package enumdiscrim provides a go/analysis probe for the discrimination
// shape the exhaustive linter is blind to: `if x != EnumConst` (or ==) where
// EnumConst belongs to a const group with more than two members. A switch
// with a missed member is exhaustive's finding; the same defect wearing an
// if-statement — the shape that silently rendered an unknown segment kind as
// raw text in go-hx — is only visible here.
//
// This is a PROBE, not a gate. A fast-path check (`if kind != Special {
// return generic }`) is a legitimate two-way split of a many-way domain, and
// only the author knows which reading is intended. To bound the noise, a
// comparison is reported only when BOTH operands share the constant's named
// type (comparing an error interface against a sentinel constant is a
// different rule's business) and only when it steers an if or for condition.
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

// message is the diagnostic for an if-shaped enum discrimination.
const message = "==/!= discriminates %s, a const group of %d members; a missed member falls through silently — prefer an exhaustive switch"

// minimumMembers is the const-group size above which a two-way comparison is
// worth flagging: two members ARE a two-way domain.
const minimumMembers = 2

// Analyzer reports if-shaped discrimination of multi-member const groups.
var Analyzer = &analysis.Analyzer{
	Name:     "enumdiscrim",
	Doc:      "reports ==/!= control-flow discrimination of a const group with more than two members, the shape exhaustive cannot see",
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

// run inspects every if and for condition for enum comparisons.
func run(pass *analysis.Pass) (any, error) {
	ins, _ := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	ins.Preorder([]ast.Node{(*ast.IfStmt)(nil), (*ast.ForStmt)(nil)}, func(n ast.Node) {
		if cond := condition(n); cond != nil {
			checkCondition(pass, cond)
		}
	})
	return nil, nil
}

// condition extracts the steering expression of an if or for statement.
func condition(n ast.Node) ast.Expr {
	switch stmt := n.(type) {
	case *ast.IfStmt:
		return stmt.Cond
	case *ast.ForStmt:
		return stmt.Cond
	}
	return nil
}

// checkCondition reports each enum comparison within the condition, including
// operands of && and ||.
func checkCondition(pass *analysis.Pass, cond ast.Expr) {
	ast.Inspect(cond, func(n ast.Node) bool {
		cmp, ok := n.(*ast.BinaryExpr)
		if ok && (cmp.Op == token.EQL || cmp.Op == token.NEQ) {
			checkComparison(pass, cmp)
		}
		return true
	})
}

// checkComparison reports a comparison whose operands share a named type with
// a const group larger than two.
func checkComparison(pass *analysis.Pass, cmp *ast.BinaryExpr) {
	named := comparedEnum(pass, cmp)
	if named == nil {
		return
	}
	if count := memberCount(named); count > minimumMembers {
		pass.Reportf(cmp.Pos(), message, named.Obj().Name(), count)
	}
}

// comparedEnum returns the named type of an enum-discrimination shape: one
// operand a declared constant, both operands of the SAME named type.
func comparedEnum(pass *analysis.Pass, cmp *ast.BinaryExpr) *types.Named {
	if !isDeclaredConst(pass, cmp.X) && !isDeclaredConst(pass, cmp.Y) {
		return nil
	}
	left, ok := types.Unalias(pass.TypesInfo.TypeOf(cmp.X)).(*types.Named)
	if !ok || !types.Identical(left, pass.TypesInfo.TypeOf(cmp.Y)) {
		return nil
	}
	return left
}

// isDeclaredConst reports whether expr directly references a declared constant.
func isDeclaredConst(pass *analysis.Pass, expr ast.Expr) bool {
	obj := referencedObject(pass, expr)
	_, ok := obj.(*types.Const)
	return ok
}

// referencedObject resolves a plain identifier or selector to its object.
func referencedObject(pass *analysis.Pass, expr ast.Expr) types.Object {
	switch e := ast.Unparen(expr).(type) {
	case *ast.Ident:
		return pass.TypesInfo.ObjectOf(e)
	case *ast.SelectorExpr:
		return pass.TypesInfo.ObjectOf(e.Sel)
	}
	return nil
}

// memberCount counts the constants of the named type declared in its own
// package — the const group the comparison discriminates.
func memberCount(named *types.Named) int {
	pkg := named.Obj().Pkg()
	if pkg == nil {
		return 0
	}
	count := 0
	for _, name := range pkg.Scope().Names() {
		if isMember(pkg.Scope().Lookup(name), named) {
			count++
		}
	}
	return count
}

// isMember reports whether obj is a constant of the named type.
func isMember(obj types.Object, named *types.Named) bool {
	constant, ok := obj.(*types.Const)
	return ok && types.Identical(constant.Type(), named)
}
