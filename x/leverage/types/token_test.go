package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v2/x/leverage/types"
)

func TestUTokenFromTokenDenom(t *testing.T) {
	tokenDenom := "uumee"
	uTokenDenom := types.UTokenFromTokenDenom(tokenDenom)
	require.Equal(t, "u/"+tokenDenom, uTokenDenom)
	require.NoError(t, sdk.ValidateDenom(uTokenDenom))
}

func TestUpdateRegistryProposal_String(t *testing.T) {
	p := types.UpdateRegistryProposal{
		Title:       "test",
		Description: "test",
		Registry: []types.Token{
			{
				BaseDenom:            "uumee",
				SymbolDenom:          "umee",
				Exponent:             6,
				ReserveFactor:        sdk.NewDec(40),
				CollateralWeight:     sdk.NewDec(43),
				LiquidationThreshold: sdk.NewDec(66),
				BaseBorrowRate:       sdk.NewDec(32),
				KinkBorrowRate:       sdk.NewDec(26),
				MaxBorrowRate:        sdk.NewDec(21),
				KinkUtilization:      sdk.MustNewDecFromStr("0.25"),
				LiquidationIncentive: sdk.NewDec(88),
				EnableMsgLend:        true,
				EnableMsgBorrow:      true,
				Blacklist:            false,
			},
		},
	}
	expected := `title: test
description: test
registry:
    - base_denom: uumee
      reserve_factor: "40.000000000000000000"
      collateral_weight: "43.000000000000000000"
      liquidation_threshold: "66.000000000000000000"
      base_borrow_rate: "32.000000000000000000"
      kink_borrow_rate: "26.000000000000000000"
      max_borrow_rate: "21.000000000000000000"
      kink_utilization_rate: "0.250000000000000000"
      liquidation_incentive: "88.000000000000000000"
      symbol_denom: umee
      exponent: 6
      enable_msg_lend: true
      enable_msg_borrow: true
      blacklist: false
`
	require.Equal(t, expected, p.String())
}

func TestToken_Validate(t *testing.T) {
	testCases := map[string]struct {
		input     types.Token
		expectErr bool
	}{
		"valid token": {
			input: types.Token{
				BaseDenom:            "uumee",
				SymbolDenom:          "umee",
				Exponent:             6,
				ReserveFactor:        sdk.MustNewDecFromStr("0.25"),
				CollateralWeight:     sdk.MustNewDecFromStr("0.50"),
				LiquidationThreshold: sdk.MustNewDecFromStr("0.50"),
				BaseBorrowRate:       sdk.MustNewDecFromStr("0.01"),
				KinkBorrowRate:       sdk.MustNewDecFromStr("0.05"),
				MaxBorrowRate:        sdk.MustNewDecFromStr("1.0"),
				KinkUtilization:      sdk.MustNewDecFromStr("0.75"),
				LiquidationIncentive: sdk.MustNewDecFromStr("0.05"),
				EnableMsgLend:        true,
				EnableMsgBorrow:      true,
				Blacklist:            false,
			},
		},
		"invalid base token": {
			input: types.Token{
				BaseDenom:            "$$",
				ReserveFactor:        sdk.MustNewDecFromStr("0.25"),
				CollateralWeight:     sdk.MustNewDecFromStr("0.50"),
				LiquidationThreshold: sdk.MustNewDecFromStr("0.50"),
				BaseBorrowRate:       sdk.MustNewDecFromStr("0.01"),
				KinkBorrowRate:       sdk.MustNewDecFromStr("0.05"),
				MaxBorrowRate:        sdk.MustNewDecFromStr("1.0"),
				KinkUtilization:      sdk.MustNewDecFromStr("0.75"),
				LiquidationIncentive: sdk.MustNewDecFromStr("0.05"),
				EnableMsgLend:        true,
				EnableMsgBorrow:      true,
				Blacklist:            false,
			},
			expectErr: true,
		},
		"invalid base token (utoken)": {
			input: types.Token{
				BaseDenom:            "u/uumee",
				ReserveFactor:        sdk.MustNewDecFromStr("0.25"),
				CollateralWeight:     sdk.MustNewDecFromStr("0.50"),
				LiquidationThreshold: sdk.MustNewDecFromStr("0.50"),
				BaseBorrowRate:       sdk.MustNewDecFromStr("0.01"),
				KinkBorrowRate:       sdk.MustNewDecFromStr("0.05"),
				MaxBorrowRate:        sdk.MustNewDecFromStr("1.0"),
				KinkUtilization:      sdk.MustNewDecFromStr("0.75"),
				LiquidationIncentive: sdk.MustNewDecFromStr("0.05"),
				EnableMsgLend:        true,
				EnableMsgBorrow:      true,
				Blacklist:            false,
			},
			expectErr: true,
		},
		"invalid reserve factor": {
			input: types.Token{
				BaseDenom:            "uumee",
				ReserveFactor:        sdk.MustNewDecFromStr("-0.25"),
				CollateralWeight:     sdk.MustNewDecFromStr("0.50"),
				LiquidationThreshold: sdk.MustNewDecFromStr("0.50"),
				BaseBorrowRate:       sdk.MustNewDecFromStr("0.01"),
				KinkBorrowRate:       sdk.MustNewDecFromStr("0.05"),
				MaxBorrowRate:        sdk.MustNewDecFromStr("1.0"),
				KinkUtilization:      sdk.MustNewDecFromStr("0.75"),
				LiquidationIncentive: sdk.MustNewDecFromStr("0.05"),
				EnableMsgLend:        true,
				EnableMsgBorrow:      true,
				Blacklist:            false,
			},
			expectErr: true,
		},
		"invalid collateral weight": {
			input: types.Token{
				BaseDenom:            "uumee",
				ReserveFactor:        sdk.MustNewDecFromStr("0.25"),
				CollateralWeight:     sdk.MustNewDecFromStr("50.00"),
				LiquidationThreshold: sdk.MustNewDecFromStr("0.50"),
				BaseBorrowRate:       sdk.MustNewDecFromStr("0.01"),
				KinkBorrowRate:       sdk.MustNewDecFromStr("0.05"),
				MaxBorrowRate:        sdk.MustNewDecFromStr("1.0"),
				KinkUtilization:      sdk.MustNewDecFromStr("0.75"),
				LiquidationIncentive: sdk.MustNewDecFromStr("0.05"),
				EnableMsgLend:        true,
				EnableMsgBorrow:      true,
				Blacklist:            false,
			},
			expectErr: true,
		},
		"invalid liquidation threshold": {
			input: types.Token{
				BaseDenom:            "uumee",
				ReserveFactor:        sdk.MustNewDecFromStr("0.25"),
				CollateralWeight:     sdk.MustNewDecFromStr("0.50"),
				LiquidationThreshold: sdk.MustNewDecFromStr("0.40"),
				BaseBorrowRate:       sdk.MustNewDecFromStr("0.01"),
				KinkBorrowRate:       sdk.MustNewDecFromStr("0.05"),
				MaxBorrowRate:        sdk.MustNewDecFromStr("1.0"),
				KinkUtilization:      sdk.MustNewDecFromStr("0.75"),
				LiquidationIncentive: sdk.MustNewDecFromStr("0.05"),
				EnableMsgLend:        true,
				EnableMsgBorrow:      true,
				Blacklist:            false,
			},
			expectErr: true,
		},
		"invalid base borrow rate": {
			input: types.Token{
				BaseDenom:            "uumee",
				ReserveFactor:        sdk.MustNewDecFromStr("0.25"),
				CollateralWeight:     sdk.MustNewDecFromStr("0.50"),
				LiquidationThreshold: sdk.MustNewDecFromStr("0.50"),
				BaseBorrowRate:       sdk.MustNewDecFromStr("-0.01"),
				KinkBorrowRate:       sdk.MustNewDecFromStr("0.05"),
				MaxBorrowRate:        sdk.MustNewDecFromStr("1.0"),
				KinkUtilization:      sdk.MustNewDecFromStr("0.75"),
				LiquidationIncentive: sdk.MustNewDecFromStr("0.05"),
				EnableMsgLend:        true,
				EnableMsgBorrow:      true,
				Blacklist:            false,
			},
			expectErr: true,
		},
		"invalid kink borrow rate": {
			input: types.Token{
				BaseDenom:            "uumee",
				ReserveFactor:        sdk.MustNewDecFromStr("0.25"),
				CollateralWeight:     sdk.MustNewDecFromStr("0.50"),
				LiquidationThreshold: sdk.MustNewDecFromStr("0.50"),
				BaseBorrowRate:       sdk.MustNewDecFromStr("0.01"),
				KinkBorrowRate:       sdk.MustNewDecFromStr("-0.05"),
				MaxBorrowRate:        sdk.MustNewDecFromStr("1.0"),
				KinkUtilization:      sdk.MustNewDecFromStr("0.75"),
				LiquidationIncentive: sdk.MustNewDecFromStr("0.05"),
				EnableMsgLend:        true,
				EnableMsgBorrow:      true,
				Blacklist:            false,
			},
			expectErr: true,
		},
		"invalid max borrow rate": {
			input: types.Token{
				BaseDenom:            "uumee",
				ReserveFactor:        sdk.MustNewDecFromStr("0.25"),
				CollateralWeight:     sdk.MustNewDecFromStr("0.50"),
				LiquidationThreshold: sdk.MustNewDecFromStr("0.50"),
				BaseBorrowRate:       sdk.MustNewDecFromStr("0.01"),
				KinkBorrowRate:       sdk.MustNewDecFromStr("0.05"),
				MaxBorrowRate:        sdk.MustNewDecFromStr("-1.0"),
				KinkUtilization:      sdk.MustNewDecFromStr("0.75"),
				LiquidationIncentive: sdk.MustNewDecFromStr("0.05"),
				EnableMsgLend:        true,
				EnableMsgBorrow:      true,
				Blacklist:            false,
			},
			expectErr: true,
		},
		"invalid kink utilization rate": {
			input: types.Token{
				BaseDenom:            "uumee",
				ReserveFactor:        sdk.MustNewDecFromStr("0.25"),
				CollateralWeight:     sdk.MustNewDecFromStr("0.50"),
				LiquidationThreshold: sdk.MustNewDecFromStr("0.50"),
				BaseBorrowRate:       sdk.MustNewDecFromStr("0.01"),
				KinkBorrowRate:       sdk.MustNewDecFromStr("0.05"),
				MaxBorrowRate:        sdk.MustNewDecFromStr("1.0"),
				KinkUtilization:      sdk.ZeroDec(),
				LiquidationIncentive: sdk.MustNewDecFromStr("0.05"),
				EnableMsgLend:        true,
				EnableMsgBorrow:      true,
				Blacklist:            false,
			},
			expectErr: true,
		},
		"invalid liquidation incentive": {
			input: types.Token{
				BaseDenom:            "uumee",
				ReserveFactor:        sdk.MustNewDecFromStr("0.25"),
				CollateralWeight:     sdk.MustNewDecFromStr("0.50"),
				LiquidationThreshold: sdk.MustNewDecFromStr("0.50"),
				BaseBorrowRate:       sdk.MustNewDecFromStr("0.01"),
				KinkBorrowRate:       sdk.MustNewDecFromStr("0.05"),
				MaxBorrowRate:        sdk.MustNewDecFromStr("1.0"),
				KinkUtilization:      sdk.MustNewDecFromStr("0.75"),
				LiquidationIncentive: sdk.MustNewDecFromStr("-0.05"),
				EnableMsgLend:        true,
				EnableMsgBorrow:      true,
				Blacklist:            false,
			},
			expectErr: true,
		},
		"blacklisted but lend enabled": {
			input: types.Token{
				BaseDenom:            "uumee",
				SymbolDenom:          "umee",
				Exponent:             6,
				ReserveFactor:        sdk.MustNewDecFromStr("0.25"),
				CollateralWeight:     sdk.MustNewDecFromStr("0.50"),
				LiquidationThreshold: sdk.MustNewDecFromStr("0.50"),
				BaseBorrowRate:       sdk.MustNewDecFromStr("0.01"),
				KinkBorrowRate:       sdk.MustNewDecFromStr("0.05"),
				MaxBorrowRate:        sdk.MustNewDecFromStr("1.0"),
				KinkUtilization:      sdk.MustNewDecFromStr("0.75"),
				LiquidationIncentive: sdk.MustNewDecFromStr("0.05"),
				EnableMsgLend:        true,
				EnableMsgBorrow:      false,
				Blacklist:            true,
			},
			expectErr: true,
		},
		"blacklisted but borrow enabled": {
			input: types.Token{
				BaseDenom:            "uumee",
				SymbolDenom:          "umee",
				Exponent:             6,
				ReserveFactor:        sdk.MustNewDecFromStr("0.25"),
				CollateralWeight:     sdk.MustNewDecFromStr("0.50"),
				LiquidationThreshold: sdk.MustNewDecFromStr("0.50"),
				BaseBorrowRate:       sdk.MustNewDecFromStr("0.01"),
				KinkBorrowRate:       sdk.MustNewDecFromStr("0.05"),
				MaxBorrowRate:        sdk.MustNewDecFromStr("1.0"),
				KinkUtilization:      sdk.MustNewDecFromStr("0.75"),
				LiquidationIncentive: sdk.MustNewDecFromStr("0.05"),
				EnableMsgLend:        false,
				EnableMsgBorrow:      true,
				Blacklist:            true,
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
