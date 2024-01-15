package uibc

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"
)

func TestGenesisValidation(t *testing.T) {
	gs := DefaultGenesisState()
	err := gs.Validate()
	assert.NilError(t, err)

	gs.OutflowSum = sdkmath.LegacyNewDec(-123123)
	err = gs.Validate()
	assert.ErrorContains(t, err, "outflow sum cannot be negative")

	gs.Outflows = []sdk.DecCoin{{Denom: "umee", Amount: sdkmath.LegacyNewDec(-11123123)}}
	err = gs.Validate()
	assert.ErrorContains(t, err, "amount cannot be negative")
}
