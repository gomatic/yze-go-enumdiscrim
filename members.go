package enumdiscrim

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// comparedEnum returns the named type an enum-discrimination shape compares and
// the exact value of the inhabitant it names: one operand a constant of the
// type, both operands of the SAME named type.
func comparedEnum(pass *analysis.Pass, cmp *ast.BinaryExpr) (*types.Named, string, bool) {
	value, ok := constantValue(pass, cmp.X, cmp.Y)
	if !ok {
		return nil, "", false
	}
	left, isNamed := types.Unalias(pass.TypesInfo.TypeOf(cmp.X)).(*types.Named)
	if !isNamed || !types.Identical(left, pass.TypesInfo.TypeOf(cmp.Y)) {
		return nil, "", false
	}
	return left, value, true
}

// constantValue returns the exact value of whichever operand is a constant,
// preferring the right — the ordinary spelling — and accepting the left. The
// operand is judged by being CONSTANT, not by being a declared name, so that
// `k != 0` and `k != Kind(0)` discriminate the same inhabitant as `k != KindA`
// and rewriting one spelling as another buys no silence.
func constantValue(pass *analysis.Pass, left, right ast.Expr) (string, bool) {
	if value, ok := exactValue(pass, right); ok {
		return value, true
	}
	return exactValue(pass, left)
}

// exactValue returns expr's constant value, or false when expr is not constant.
// It does not unparenthesise, because go/types already records a parenthesised
// expression with the value of what it parenthesises; a second Unalias on the
// right operand of a comparison is dead for the same reason, since
// types.Identical resolves aliases itself. Each was driven as a mutant over 93
// modules and produced output no run could tell from the sound binary's.
func exactValue(pass *analysis.Pass, expr ast.Expr) (string, bool) {
	typed, recorded := pass.TypesInfo.Types[expr]
	if !recorded || typed.Value == nil {
		return "", false
	}
	return typed.Value.ExactString(), true
}

// memberValues returns the INHABITANTS of the named type's const group — the
// distinct values a comparison can discriminate, not the declarations that
// name them. A second constant naming an existing value adds a name to the
// package and no member to the domain, and counting it would both cross the
// threshold and state a falsehood in the message.
//
// A package-less named type is refused before this point: localToModule
// answers false for a type no package declares, since such a type is in no
// module. That refusal is the single answer for the case. The older reasoning
// — that (*types.Package).Scope() returns the universe scope for a nil
// receiver, so the group comes out empty and the threshold declines it — is
// redundant beside it rather than a second justification.
func memberValues(named *types.Named) map[string]struct{} {
	scope := named.Obj().Pkg().Scope()
	inhabitants := make(map[string]struct{})
	for _, name := range scope.Names() {
		if member, ok := memberValue(scope.Lookup(name), named); ok {
			inhabitants[member] = struct{}{}
		}
	}
	return inhabitants
}

// memberValue returns the exact value of obj when obj is a constant of the
// named type. ExactString is the identity of a constant value: two names for
// one inhabitant share it, and no two inhabitants do.
func memberValue(obj types.Object, named *types.Named) (string, bool) {
	constant, ok := obj.(*types.Const)
	if !ok || !types.Identical(constant.Type(), named) {
		return "", false
	}
	return constant.Val().ExactString(), true
}
