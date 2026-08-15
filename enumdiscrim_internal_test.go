package enumdiscrim

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis"
)

// source is a Go file whose expressions a test needs typed.
type source string

// expression is the source of one expression inside that file.
type expression string

// typed type-checks src and returns a pass carrying its type information plus a
// lookup from expression source to the AST node holding it.
func typed(t *testing.T, src source) (*analysis.Pass, map[expression]ast.Expr) {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "probe.go", string(src), 0)
	require.NoError(t, err)

	info := &types.Info{
		Types: map[ast.Expr]types.TypeAndValue{},
		Defs:  map[*ast.Ident]types.Object{},
		Uses:  map[*ast.Ident]types.Object{},
	}
	config := types.Config{Importer: importer.Default()}
	_, err = config.Check("probe", fset, []*ast.File{file}, info)
	require.NoError(t, err)

	found := map[expression]ast.Expr{}
	ast.Inspect(file, func(n ast.Node) bool {
		if expr, ok := n.(ast.Expr); ok {
			found[expression(types.ExprString(expr))] = expr
		}
		return true
	})
	return &analysis.Pass{TypesInfo: info}, found
}

// TestPartsComposesTestsExactlyWhenOperandsAreBoolean names the invariant parts
// documents: an operator composes tests exactly when its operands are boolean,
// which is why the composition is read off the operand types and never off an
// enumeration of operator tokens. Both directions are asserted — every operator
// Go gives a bool operand composes, and every operator that does not, does not.
func TestPartsComposesTestsExactlyWhenOperandsAreBoolean(t *testing.T) {
	t.Parallel()

	pass, found := typed(t, `package probe

func probe(a, b bool, m, n int) bool {
	_ = !a
	_ = a && b
	_ = a || b
	_ = -n
	_ = m + n
	_ = m < n
	_ = m == n
	return a
}
`)

	composes := []expression{"!a", "a && b", "a || b"}
	for _, spelling := range composes {
		comparison, operands := parts(pass, found[spelling])
		assert.Nil(t, comparison, spelling)
		assert.NotEmpty(t, operands, spelling)
	}

	computes := []expression{"-n", "m + n", "m < n"}
	for _, spelling := range computes {
		comparison, operands := parts(pass, found[spelling])
		assert.Nil(t, comparison, spelling)
		assert.Empty(t, operands, spelling)
	}

	comparison, operands := parts(pass, found["m == n"])
	assert.NotNil(t, comparison)
	assert.Empty(t, operands)
}

// TestPartsIsNeitherForANonOperator pins the third answer: an expression that is
// no operator at all is neither a test nor a composition of tests.
func TestPartsIsNeitherForANonOperator(t *testing.T) {
	t.Parallel()

	pass, found := typed(t, `package probe

func probe(flags map[bool]bool, a bool) bool {
	return flags[a]
}
`)

	comparison, operands := parts(pass, found["flags[a]"])

	assert.Nil(t, comparison)
	assert.Empty(t, operands)
}

// TestSteeredIgnoresAForWithNoCondition pins the one optional condition in the
// grammar: `for {}` steers on nothing, so the chain it contributes is empty and
// no group is accumulated from it.
func TestSteeredIgnoresAForWithNoCondition(t *testing.T) {
	t.Parallel()

	assert.Empty(t, steered(&ast.ForStmt{}, map[ast.Node]bool{}))
}
