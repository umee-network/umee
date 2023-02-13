//xgo:build experimental
// +xbuild experimental

package uibc

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"
)

func TestGenesisValidation(t *testing.T) {
	gs := DefaultGenesisState()
	err := gs.Validate()
	assert.NilError(t, err)

	gs.TotalOutflowSum = sdk.NewDec(-123123)
	err = gs.Validate()
	assert.ErrorContains(t, err, "total outflow sum cannot be negative")

	gs.Quotas = sdk.DecCoins{sdk.NewInt64DecCoin("umee", -11123123)}
	err = gs.Validate()
	assert.ErrorContains(t, err, "amount cannot be negative")
}
