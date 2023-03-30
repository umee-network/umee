package incentive

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()
	assert.NilError(t, params.Validate())

	invalidMaxUnbondings := DefaultParams()
	invalidMaxUnbondings.MaxUnbondings = 0
	assert.ErrorContains(t, invalidMaxUnbondings.Validate(), "max unbondings cannot be zero")

	invalidCommunityFund := DefaultParams()
	invalidCommunityFund.CommunityFundAddress = "abcdefgh"
	assert.ErrorContains(t, invalidCommunityFund.Validate(), "decoding bech32 failed")
}
