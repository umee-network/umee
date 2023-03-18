package incentive

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestParams(t *testing.T) {
	params := DefaultParams()
	err := params.Validate()
	assert.NilError(t, err)
}
