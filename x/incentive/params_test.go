package incentive

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gotest.tools/v3/assert"
)

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()
	assert.NilError(t, params.Validate())

	invalidMaxUnbondings := DefaultParams()
	invalidMaxUnbondings.MaxUnbondings = 0
	assert.ErrorContains(t, invalidMaxUnbondings.Validate(), "max unbondings cannot be zero")

	invalidLongUnbond := DefaultParams()
	invalidLongUnbond.UnbondingDurationLong = 0
	assert.ErrorContains(t, invalidLongUnbond.Validate(), "unbonding duration cannot be zero")

	invalidMidUnbond := DefaultParams()
	invalidMidUnbond.UnbondingDurationMiddle = 0
	assert.ErrorContains(t, invalidMidUnbond.Validate(), "unbonding duration cannot be zero")

	invalidShortUnbond := DefaultParams()
	invalidShortUnbond.UnbondingDurationShort = 0
	assert.ErrorContains(t, invalidShortUnbond.Validate(), "unbonding duration cannot be zero")

	invalidMidTier := DefaultParams()
	invalidMidTier.TierWeightMiddle = sdk.MustNewDecFromStr("1.5")
	assert.ErrorContains(t, invalidMidTier.Validate(), "tier weight cannot exceed 1")

	negativeMidTier := DefaultParams()
	negativeMidTier.TierWeightMiddle = sdk.MustNewDecFromStr("-0.5")
	assert.ErrorContains(t, negativeMidTier.Validate(), "tier weight cannot be negative")

	invalidShortTier := DefaultParams()
	invalidShortTier.TierWeightShort = sdk.MustNewDecFromStr("1.5")
	assert.ErrorContains(t, invalidShortTier.Validate(), "tier weight cannot exceed 1")

	negativeShortTier := DefaultParams()
	negativeShortTier.TierWeightShort = sdk.MustNewDecFromStr("-0.5")
	assert.ErrorContains(t, negativeShortTier.Validate(), "tier weight cannot be negative")

	invalidCommunityFund := DefaultParams()
	invalidCommunityFund.CommunityFundAddress = "abcdefgh"
	assert.ErrorContains(t, invalidCommunityFund.Validate(), "decoding bech32 failed")

	invalidTierOrder := DefaultParams()
	invalidTierOrder.UnbondingDurationLong = 1
	assert.ErrorIs(t, invalidTierOrder.Validate(), ErrUnbondingTierOrder)

	invalidWeightOrder := DefaultParams()
	invalidWeightOrder.TierWeightShort = sdk.MustNewDecFromStr("0.9")
	assert.ErrorIs(t, invalidWeightOrder.Validate(), ErrUnbondingWeightOrder)
}
