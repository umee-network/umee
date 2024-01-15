package incentive

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"gotest.tools/v3/assert"
)

func TestDefaultParams(t *testing.T) {
	t.Parallel()

	params := DefaultParams()
	assert.NilError(t, params.Validate())

	invalidUnbondingDuration := DefaultParams()
	invalidUnbondingDuration.UnbondingDuration = -1
	assert.ErrorContains(t, invalidUnbondingDuration.Validate(), "invalid unbonding duration")

	invalidEmergencyUnbondFee := DefaultParams()
	invalidEmergencyUnbondFee.EmergencyUnbondFee = sdkmath.LegacyOneDec()
	assert.ErrorContains(t, invalidEmergencyUnbondFee.Validate(), "invalid emergency unbonding fee")
}
