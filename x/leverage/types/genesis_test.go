package types

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestGenesisValidation(t *testing.T) {
	genState := DefaultGenesis()
	assert.NilError(t, genState.Validate())

	// TODO #484: expand this test to cover failure cases.
}
