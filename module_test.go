package enumdiscrim

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis"
)

// TestLocalToModule covers the module-identity answers the analysistest corpus
// cannot reach: its driver loads no module metadata, so pass.Module is nil
// there and every package but the analyzed one takes the same branch. Each
// assertion names the answer the rule depends on, not the line it executes.
func TestLocalToModule(t *testing.T) {
	t.Parallel()

	self := types.NewPackage("example.test/mod/pkg", "pkg")
	sub := types.NewPackage("example.test/mod/pkg/sub", "sub")
	root := types.NewPackage("example.test/mod", "mod")
	other := types.NewPackage("example.test/other", "other")
	modular := types.NewPackage("example.test/modular", "modular")
	mod := &analysis.Module{Path: "example.test/mod"}

	assert.False(t, localToModule(&analysis.Pass{Pkg: self, Module: mod}, nil),
		"a package-less named type belongs to no module")
	assert.True(t, localToModule(&analysis.Pass{Pkg: self}, self),
		"the analyzed package is local with or without module metadata")
	assert.False(t, localToModule(&analysis.Pass{Pkg: self}, sub),
		"nil module: a sibling of the analyzed package is not known to be local")
	assert.False(t, localToModule(&analysis.Pass{Pkg: self, Module: &analysis.Module{}}, sub),
		"an empty module path behaves like nil")
	assert.True(t, localToModule(&analysis.Pass{Pkg: self, Module: mod}, sub),
		"a module subpath is local")
	assert.True(t, localToModule(&analysis.Pass{Pkg: self, Module: mod}, root),
		"the module root package is local")
	assert.False(t, localToModule(&analysis.Pass{Pkg: self, Module: mod}, other),
		"a foreign path is not local")
	assert.False(t, localToModule(&analysis.Pass{Pkg: self, Module: mod}, modular),
		"a module path matches whole path elements: example.test/mod does not swallow example.test/modular")
}

// TestRecordJudgesByWhereTheGroupIsDeclared names the split the fix rests on,
// at the one place it is decided. Two comparisons identical in every respect
// the rule judges — three inhabitants, a constant operand naming one, both
// operands of the same named type — differ only in which module declares the
// group, and only the local one opens an accumulator.
//
// The corpus pins the same boundary through the analyzer, but its driver leaves
// pass.Module nil, so there the foreign group is refused by the fallback. This
// pins the refusal that a real run makes, by module PATH, with the metadata
// loaded: a mutant that keeps the fallback and drops the path comparison is
// invisible to the corpus and dies here.
func TestRecordJudgesByWhereTheGroupIsDeclared(t *testing.T) {
	t.Parallel()

	consumer := types.NewPackage("example.test/mod/pkg", "pkg")
	local, localCmp, info := group(types.NewPackage("example.test/mod/kinds", "kinds"), "Own")
	_, foreignCmp, foreignInfo := group(types.NewPackage("example.test/other", "other"), "Their")
	for expr, tv := range foreignInfo {
		info[expr] = tv
	}
	pass := &analysis.Pass{
		Pkg:       consumer,
		Module:    &analysis.Module{Path: "example.test/mod"},
		TypesInfo: &types.Info{Types: info},
	}

	opened := record(pass, []*discrimination{}, localCmp)

	assert.Len(t, opened, 1, "a group declared inside the analyzed module is judged")
	assert.True(t, types.Identical(local, opened[0].group))
	assert.Equal(t, 3, opened[0].total, "the whole group is counted, not only the compared member")
	assert.Empty(t, record(pass, []*discrimination{}, foreignCmp),
		"an identical group declared outside the analyzed module is refused")
}

// group declares a three-member const group in pkg and returns it beside a
// synthetic `x != <first member>` comparison and the type information a pass
// needs in order to read that comparison.
func group(pkg *types.Package, name string) (*types.Named, *ast.BinaryExpr, map[ast.Expr]types.TypeAndValue) {
	named := types.NewNamed(types.NewTypeName(token.NoPos, pkg, name, nil), types.Typ[types.Int], nil)
	for i, suffix := range []string{"A", "B", "C"} {
		pkg.Scope().Insert(types.NewConst(token.NoPos, pkg, name+suffix, named, constant.MakeInt64(int64(i))))
	}
	subject, member := ast.NewIdent("x"), ast.NewIdent(name+"A")
	info := map[ast.Expr]types.TypeAndValue{
		subject: {Type: named},
		member:  {Type: named, Value: constant.MakeInt64(0)},
	}
	return named, &ast.BinaryExpr{X: subject, Op: token.NEQ, Y: member}, info
}
