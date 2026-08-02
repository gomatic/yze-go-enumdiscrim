package enumdiscrim_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis/analysistest"

	enumdiscrim "github.com/gomatic/yze-go-enumdiscrim"
)

// TestIfShapedDiscrimination pins both directions: the if and for conditions
// discriminating the three-member group report; the exhaustive switch, the
// two-member domain, the stored comparison, and the value-to-value comparison
// stay silent.
func TestIfShapedDiscrimination(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), enumdiscrim.Analyzer, "a")
}

// TestRegistrationIsWellFormed pins the yze wiring.
func TestRegistrationIsWellFormed(t *testing.T) {
	assert.NoError(t, enumdiscrim.Registration.Validate())
	assert.Equal(t, "yze/enumdiscrim", enumdiscrim.Registration.RuleID())
	assert.Same(t, enumdiscrim.Analyzer, enumdiscrim.Registration.Analyzer)
}
