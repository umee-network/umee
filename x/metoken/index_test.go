package metoken

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"
)

func TestIndex_Validate(t *testing.T) {
	invalidMaxSupply := validIndex()
	invalidMaxSupply.Denom = "me/USD"
	invalidMaxSupply.MaxSupply = sdkmath.NewInt(-1)

	invalidFee := validIndex()
	invalidFee.Fee = NewFee(sdk.MustNewDecFromStr("-1.0"), sdk.Dec{}, sdk.Dec{})

	invalidDenomAcceptedAsset := validIndex()
	invalidDenomAcceptedAsset.AcceptedAssets = []AcceptedAsset{
		NewAcceptedAsset("????", sdk.MustNewDecFromStr("-0.2"), sdk.MustNewDecFromStr("1.0")),
	}

	invalidAcceptedAsset := validIndex()
	invalidAcceptedAsset.AcceptedAssets = []AcceptedAsset{
		NewAcceptedAsset("USDT", sdk.MustNewDecFromStr("-0.2"), sdk.MustNewDecFromStr("1.0")),
	}

	invalidTargetAllocation := validIndex()
	invalidTargetAllocation.AcceptedAssets = []AcceptedAsset{
		validAcceptedAsset("USDT"),
		validAcceptedAsset("USDC"),
	}

	duplicatedAcceptedAsset := validIndex()
	duplicate := validAcceptedAsset("USDT")
	duplicate.TargetAllocation = sdk.MustNewDecFromStr("0.5")
	duplicatedAcceptedAsset.AcceptedAssets = []AcceptedAsset{
		duplicate,
		duplicate,
	}

	tcs := []struct {
		name   string
		i      Index
		errMsg string
	}{
		{"valid index", validIndex(), ""},
		{
			"invalid max supply",
			invalidMaxSupply,
			"maxSupply cannot be negative",
		},
		{
			"invalid fee",
			invalidFee,
			"should be between 0.0 and 1.0",
		},
		{
			"invalid denom accepted asset",
			invalidDenomAcceptedAsset,
			"invalid denom",
		},
		{
			"invalid accepted asset",
			invalidAcceptedAsset,
			"should be between 0.0 and 1.0",
		},
		{
			"invalid total allocation",
			invalidTargetAllocation,
			"of all the accepted assets should be 1.0",
		},
		{
			"duplicated accepted asset",
			duplicatedAcceptedAsset,
			"duplicated accepted asset",
		},
	}

	for _, tc := range tcs {
		t.Run(
			tc.name, func(t *testing.T) {
				err := tc.i.Validate()
				if tc.errMsg != "" {
					assert.ErrorContains(t, err, tc.errMsg)
				} else {
					assert.NilError(t, err)
				}
			},
		)
	}
}

func TestFee_Validate(t *testing.T) {
	invalidMinFee := validFee()
	invalidMinFee.MinFee = sdk.MustNewDecFromStr("1.01")

	negativeBalancedFee := validFee()
	negativeBalancedFee.BalancedFee = sdk.MustNewDecFromStr("-1.01")

	greaterOneBalancedFee := validFee()
	greaterOneBalancedFee.BalancedFee = sdk.MustNewDecFromStr("1.01")

	balancedFeeLowerMinFee := validFee()
	balancedFeeLowerMinFee.BalancedFee = sdk.MustNewDecFromStr("0.0001")

	negativeMaxFee := validFee()
	negativeMaxFee.MaxFee = sdk.MustNewDecFromStr("-1.01")

	greaterOneMaxFee := validFee()
	greaterOneMaxFee.MaxFee = sdk.MustNewDecFromStr("1.01")

	maxFeeEqualBalancedFee := validFee()
	maxFeeEqualBalancedFee.MaxFee = sdk.MustNewDecFromStr("0.2")

	tcs := []struct {
		name   string
		f      Fee
		errMsg string
	}{
		{"valid fee", validFee(), ""},
		{
			"min_fee > 1.0",
			invalidMinFee,
			"should be between 0.0 and 1.0",
		},
		{
			"negative balanced_fee",
			negativeBalancedFee,
			"should be between 0.0 and 1.0",
		},
		{
			"balanced_fee > 1.0",
			greaterOneBalancedFee,
			"should be between 0.0 and 1.0",
		},
		{
			"balanced_fee lower min_fee",
			balancedFeeLowerMinFee,
			"should be greater than min_fee",
		},
		{
			"negative max_fee",
			negativeMaxFee,
			"should be between 0.0 and 1.0",
		},
		{
			"max_fee > 1.0",
			greaterOneMaxFee,
			"should be between 0.0 and 1.0",
		},
		{
			"max_fee = balanced_fee",
			maxFeeEqualBalancedFee,
			"should be greater than balanced_fee",
		},
	}

	for _, tc := range tcs {
		t.Run(
			tc.name, func(t *testing.T) {
				err := tc.f.Validate()
				if tc.errMsg != "" {
					assert.ErrorContains(t, err, tc.errMsg)
				} else {
					assert.NilError(t, err)
				}
			},
		)
	}
}

func TestAcceptedAsset_Validate(t *testing.T) {
	invalidTargetAllocation := validAcceptedAsset("USDT")
	invalidTargetAllocation.TargetAllocation = sdk.MustNewDecFromStr("1.1")

	tcs := []struct {
		name   string
		aa     AcceptedAsset
		errMsg string
	}{
		{"valid accepted asset", validAcceptedAsset("USDT"), ""},
		{
			"invalid target allocation",
			invalidTargetAllocation,
			"should be between 0.0 and 1.0",
		},
	}

	for _, tc := range tcs {
		t.Run(
			tc.name, func(t *testing.T) {
				err := tc.aa.Validate()
				if tc.errMsg != "" {
					assert.ErrorContains(t, err, tc.errMsg)
				} else {
					assert.NilError(t, err)
				}
			},
		)
	}
}

func validIndex() Index {
	return Index{
		Denom:     "me/USD",
		MaxSupply: sdkmath.ZeroInt(),
		Exponent:  6,
		Fee:       validFee(),
		AcceptedAssets: []AcceptedAsset{
			validAcceptedAsset("USDT"),
		},
	}
}

func validFee() Fee {
	return NewFee(
		sdk.MustNewDecFromStr("0.001"),
		sdk.MustNewDecFromStr("0.2"),
		sdk.MustNewDecFromStr("0.5"),
	)
}

func validAcceptedAsset(denom string) AcceptedAsset {
	return NewAcceptedAsset(
		denom,
		sdk.MustNewDecFromStr("0.2"),
		sdk.MustNewDecFromStr("1.0"),
	)
}
