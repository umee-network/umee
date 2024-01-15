package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"
)

func TestGenesisValidation(t *testing.T) {
	testAddr := "umee1s84d29zk3k20xk9f0hvczkax90l9t94g72n6wm"
	validDenom := "umee"

	tcs := []struct {
		name      string
		q         GenesisState
		expectErr bool
		errMsg    string
	}{
		{"default genesis", *DefaultGenesis(), false, ""},
		{
			"invalid params",
			*NewGenesisState(
				Params{
					CompleteLiquidationThreshold: sdkmath.LegacyMustNewDecFromStr("-0.4"),
				}, nil, nil, nil, nil, 0, nil, nil, nil, nil,
			),
			true,
			"complete liquidation threshold must be positive",
		},
		{
			"invalid token registry",
			GenesisState{
				Params: DefaultParams(),
				Registry: []Token{
					{},
				},
			},
			true,
			"invalid denom",
		},
		{
			"invalid adjusted borrows address",
			GenesisState{
				Params: DefaultParams(),
				AdjustedBorrows: []AdjustedBorrow{
					NewAdjustedBorrow("", sdk.DecCoin{}),
				},
			},
			true,
			"empty address string is not allowed",
		},
		{
			"invalid adjusted borrows amount",
			GenesisState{
				Params: DefaultParams(),
				AdjustedBorrows: []AdjustedBorrow{
					NewAdjustedBorrow(testAddr, sdk.DecCoin{}),
				},
			},
			true,
			"invalid denom",
		},
		{
			"invalid collateral address",
			GenesisState{
				Params: DefaultParams(),
				Collateral: []Collateral{
					NewCollateral("", sdk.Coin{}),
				},
			},
			true,
			"empty address string is not allowed",
		},
		{
			"invalid collateral amount",
			GenesisState{
				Params: DefaultParams(),
				Collateral: []Collateral{
					NewCollateral(testAddr, sdk.Coin{}),
				},
			},
			true,
			"invalid denom",
		},
		{
			"invalid reserves",
			GenesisState{
				Params: DefaultParams(),
				Reserves: sdk.Coins{
					sdk.Coin{
						Denom: "",
					},
				},
			},
			true,
			"invalid denom",
		},
		{
			"invalid badDebt address",
			GenesisState{
				Params: DefaultParams(),
				BadDebts: []BadDebt{
					NewBadDebt("", ""),
				},
			},
			true,
			"empty address string is not allowed",
		},
		{
			"invalid badDebt denom",
			GenesisState{
				Params: DefaultParams(),
				BadDebts: []BadDebt{
					NewBadDebt(testAddr, ""),
				},
			},
			true,
			"invalid denom",
		},
		{
			"invalid interestScalar denom",
			GenesisState{
				Params: DefaultParams(),
				InterestScalars: []InterestScalar{
					NewInterestScalar("", sdkmath.LegacyZeroDec()),
				},
			},
			true,
			"invalid denom",
		},
		{
			"invalid interestScalar address",
			GenesisState{
				Params: DefaultParams(),
				InterestScalars: []InterestScalar{
					NewInterestScalar(validDenom, sdkmath.LegacyZeroDec()),
				},
			},
			true,
			"exchange rate less than one",
		},
	}

	for _, tc := range tcs {
		t.Run(
			tc.name, func(t *testing.T) {
				err := tc.q.Validate()
				if tc.expectErr {
					assert.ErrorContains(t, err, tc.errMsg)
				} else {
					assert.NilError(t, err)
				}
			},
		)
	}
}
