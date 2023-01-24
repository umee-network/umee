//go:build experimental
// +build experimental

package uibc

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisValidation(t *testing.T) {
	gs := DefaultGenesisState()
	err := gs.Validate()
	require.NoError(t, err)

	gs.TotalOutflowSum = sdk.NewDec(-123123)
	err = gs.Validate()
	require.Error(t, err)

	gs.Quotas = []Quota{
		{
			IbcDenom:   "umee",
			OutflowSum: sdk.NewDec(-11123123),
		},
	}
	err = gs.Validate()
	require.Error(t, err)
}
