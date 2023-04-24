package incentive

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gotest.tools/v3/assert"
)

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()
	assert.NilError(t, params.Validate())

	invalidUnbondingDuration := DefaultParams()
	invalidUnbondingDuration.UnbondingDuration = -1
	assert.ErrorContains(t, invalidUnbondingDuration.Validate(), "invalid unbonding duration")

	invalidEmergencyUnbondFee := DefaultParams()
	invalidEmergencyUnbondFee.EmergencyUnbondFee = sdk.OneDec()
	assert.ErrorContains(t, invalidEmergencyUnbondFee.Validate(), "invalid emergency unbonding fee")
}
