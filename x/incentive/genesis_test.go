package incentive

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestValidateGenesis(t *testing.T) {
	genesis := DefaultGenesis()
	assert.NilError(t, genesis.Validate())
}
