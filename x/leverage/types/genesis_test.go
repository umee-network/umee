package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenesisValidation(t *testing.T) {
	genState := DefaultGenesis()
	require.NoError(t, genState.Validate())

	genState.Params.InterestEpoch = 0
	require.Error(t, genState.Validate())
}
