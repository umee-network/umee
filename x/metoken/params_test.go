package metoken

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestParams_Validate(t *testing.T) {
	p := DefaultParams()
	assert.Check(t, p.ClaimingFrequency > 0)
	assert.Check(t, p.RebalancingFrequency > 0)
}
