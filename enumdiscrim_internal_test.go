package enumdiscrim

import (
	"go/ast"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConditionIgnoresOtherStatements pins the defensive default: a statement
// that is neither if nor for steers nothing. The inspector filter makes this
// arm unreachable in a real pass; the contract is that it stays nil rather
// than panicking.
func TestConditionIgnoresOtherStatements(t *testing.T) {
	t.Parallel()

	assert.Nil(t, condition(&ast.ReturnStmt{}))
}

// TestMemberCountOfAPackagelessType pins the guard for named types declared by
// no package (the universe's error), which can hold no const group.
func TestMemberCountOfAPackagelessType(t *testing.T) {
	t.Parallel()

	named := types.NewNamed(types.NewTypeName(0, nil, "orphan", nil), types.NewStruct(nil, nil), nil)

	assert.Zero(t, memberCount(named))
}
