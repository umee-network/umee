package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/x/leverage/types"
)

func TestUpdateAssetsProposal_String(t *testing.T) {
	p := types.UpdateAssetsProposal{
		Title:       "test",
		Description: "test",
		Assets: []types.Asset{
			{
				BaseTokenDenom:   "uumee",
				ExchangeRate:     sdk.NewDec(40),
				CollateralWeight: sdk.NewDec(43),
				BaseBorrowRate:   sdk.NewDec(32),
			},
		},
	}
	expected := `title: test
description: test
assets:
    - base_token_denom: uumee
      exchange_rate: "40.000000000000000000"
      collateral_weight: "43.000000000000000000"
      base_borrow_rate: "32.000000000000000000"
`
	require.Equal(t, expected, p.String())
}

func TestAsset_Validate(t *testing.T) {
	testCases := map[string]struct {
		input     types.Asset
		expectErr bool
	}{
		"valid asset": {
			input: types.Asset{
				BaseTokenDenom:   "uumee",
				ExchangeRate:     sdk.MustNewDecFromStr("0.40"),
				CollateralWeight: sdk.MustNewDecFromStr("0.50"),
				BaseBorrowRate:   sdk.MustNewDecFromStr("0.01"),
			},
		},
		"invalid base token": {
			input: types.Asset{
				BaseTokenDenom:   "$$",
				ExchangeRate:     sdk.MustNewDecFromStr("0.40"),
				CollateralWeight: sdk.MustNewDecFromStr("0.50"),
				BaseBorrowRate:   sdk.MustNewDecFromStr("0.01"),
			},
			expectErr: true,
		},
		"invalid exchange rate": {
			input: types.Asset{
				BaseTokenDenom:   "uumee",
				ExchangeRate:     sdk.MustNewDecFromStr("40.00"),
				CollateralWeight: sdk.MustNewDecFromStr("0.50"),
				BaseBorrowRate:   sdk.MustNewDecFromStr("0.01"),
			},
			expectErr: true,
		},
		"invalid collateral weight": {
			input: types.Asset{
				BaseTokenDenom:   "uumee",
				ExchangeRate:     sdk.MustNewDecFromStr("0.40"),
				CollateralWeight: sdk.MustNewDecFromStr("50.00"),
				BaseBorrowRate:   sdk.MustNewDecFromStr("0.01"),
			},
			expectErr: true,
		},
		"invalid base borrow rate": {
			input: types.Asset{
				BaseTokenDenom:   "uumee",
				ExchangeRate:     sdk.MustNewDecFromStr("0.40"),
				CollateralWeight: sdk.MustNewDecFromStr("0.50"),
				BaseBorrowRate:   sdk.MustNewDecFromStr("10.00"),
			},
			expectErr: true,
		},
	}

	for name, tc := range testCases {
		tc := tc

		t.Run(name, func(t *testing.T) {
			err := tc.input.Validate()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
