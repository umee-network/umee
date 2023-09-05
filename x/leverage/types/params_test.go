package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gotest.tools/v3/assert"
)

func TestParams_Validate(t *testing.T) {
	negativeDec := sdk.MustNewDecFromStr("-0.4")
	exceededDec := sdk.MustNewDecFromStr("1.4")

	tcs := []struct {
		name string
		p    Params
		err  string
	}{
		{"default params", DefaultParams(), ""},
		{
			"negative complete liquidation threshold",
			Params{
				CompleteLiquidationThreshold: negativeDec,
			},
			"complete liquidation threshold must be positive",
		},
		{
			"exceeded complete liquidation threshold",
			Params{
				CompleteLiquidationThreshold: exceededDec,
			},
			"complete liquidation threshold cannot exceed 1",
		},
		{
			"negative minimum close factor",
			Params{
				CompleteLiquidationThreshold: sdk.MustNewDecFromStr("0.4"),
				MinimumCloseFactor:           negativeDec,
			},
			"minimum close factor cannot be negative",
		},
		{
			"exceeded minimum close factor",
			Params{
				CompleteLiquidationThreshold: sdk.MustNewDecFromStr("0.4"),
				MinimumCloseFactor:           exceededDec,
			},
			"minimum close factor cannot exceed 1",
		},
		{
			"negative oracle reward factor",
			Params{
				CompleteLiquidationThreshold: sdk.MustNewDecFromStr("0.4"),
				MinimumCloseFactor:           sdk.MustNewDecFromStr("0.05"),
				OracleRewardFactor:           negativeDec,
			},
			"oracle reward factor cannot be negative",
		},
		{
			"exceeded oracle reward factor",
			Params{
				CompleteLiquidationThreshold: sdk.MustNewDecFromStr("0.4"),
				MinimumCloseFactor:           sdk.MustNewDecFromStr("0.05"),
				OracleRewardFactor:           exceededDec,
			},
			"oracle reward factor cannot exceed 1",
		},
		{
			"negative small liquidation size",
			Params{
				CompleteLiquidationThreshold: sdk.MustNewDecFromStr("0.4"),
				MinimumCloseFactor:           sdk.MustNewDecFromStr("0.05"),
				OracleRewardFactor:           sdk.MustNewDecFromStr("0.01"),
				SmallLiquidationSize:         negativeDec,
			},
			"small liquidation size cannot be negative",
		},
		{
			"negative direct liquidation fee",
			Params{
				CompleteLiquidationThreshold: sdk.MustNewDecFromStr("0.4"),
				MinimumCloseFactor:           sdk.MustNewDecFromStr("0.05"),
				OracleRewardFactor:           sdk.MustNewDecFromStr("0.01"),
				SmallLiquidationSize:         sdk.MustNewDecFromStr("500.00"),
				DirectLiquidationFee:         negativeDec,
			},
			"direct liquidation fee cannot be negative",
		},
		{
			"exceeded direct liquidation fee",
			Params{
				CompleteLiquidationThreshold: sdk.MustNewDecFromStr("0.4"),
				MinimumCloseFactor:           sdk.MustNewDecFromStr("0.05"),
				OracleRewardFactor:           sdk.MustNewDecFromStr("0.01"),
				SmallLiquidationSize:         sdk.MustNewDecFromStr("500.00"),
				DirectLiquidationFee:         exceededDec,
			},
			"direct liquidation fee must be less than 1",
		},
	}

	for _, tc := range tcs {
		t.Run(
			tc.name, func(t *testing.T) {
				err := tc.p.Validate()
				if tc.err == "" {
					assert.NilError(t, err)
				} else {
					assert.ErrorContains(t, err, tc.err)
				}
			},
		)
	}
}
