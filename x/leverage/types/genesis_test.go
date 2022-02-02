package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenesisValidation(t *testing.T) {
	genState := DefaultGenesis()
	require.NoError(t, genState.Validate())

	// TODO #484: expand this test to cover failure cases.
}
