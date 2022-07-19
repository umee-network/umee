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
				BaseDenom:              "uumee",
				SymbolDenom:            "umee",
				Exponent:               6,
				ReserveFactor:          sdk.NewDec(40),
				CollateralWeight:       sdk.NewDec(43),
				LiquidationThreshold:   sdk.NewDec(66),
				BaseBorrowRate:         sdk.NewDec(32),
				KinkBorrowRate:         sdk.NewDec(26),
				MaxBorrowRate:          sdk.NewDec(21),
				KinkUtilization:        sdk.MustNewDecFromStr("0.25"),
				LiquidationIncentive:   sdk.NewDec(88),
				EnableMsgSupply:        true,
				EnableMsgBorrow:        true,
				Blacklist:              false,
				MaxCollateralShare:     sdk.MustNewDecFromStr("0.1"),
				MaxSupplyUtilization:   sdk.MustNewDecFromStr("0.5"),
				MinCollateralLiquidity: sdk.MustNewDecFromStr("0.75"),
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
      kink_utilization: "0.250000000000000000"
      liquidation_incentive: "88.000000000000000000"
      symbol_denom: umee
      exponent: 6
      enable_msg_supply: true
      enable_msg_borrow: true
      blacklist: false
      max_collateral_share: "0.100000000000000000"
      max_supply_utilization: "0.500000000000000000"
      min_collateral_liquidity: "0.750000000000000000"
`
	require.Equal(t, expected, p.String())
}

func TestToken_Validate(t *testing.T) {
	validToken := func() types.Token {
		return types.Token{
			BaseDenom:              "uumee",
			SymbolDenom:            "umee",
			Exponent:               6,
			ReserveFactor:          sdk.MustNewDecFromStr("0.25"),
			CollateralWeight:       sdk.MustNewDecFromStr("0.50"),
			LiquidationThreshold:   sdk.MustNewDecFromStr("0.50"),
			BaseBorrowRate:         sdk.MustNewDecFromStr("0.01"),
			KinkBorrowRate:         sdk.MustNewDecFromStr("0.05"),
			MaxBorrowRate:          sdk.MustNewDecFromStr("1.0"),
			KinkUtilization:        sdk.MustNewDecFromStr("0.75"),
			LiquidationIncentive:   sdk.MustNewDecFromStr("0.05"),
			EnableMsgSupply:        true,
			EnableMsgBorrow:        true,
			Blacklist:              false,
			MaxCollateralShare:     sdk.MustNewDecFromStr("1.0"),
			MaxSupplyUtilization:   sdk.MustNewDecFromStr("0.9"),
			MinCollateralLiquidity: sdk.MustNewDecFromStr("1.0"),
		}
	}
	invalidBaseToken := validToken()
	invalidBaseToken.BaseDenom = "$$"
	invalidBaseToken.SymbolDenom = ""

	invalidUToken := validToken()
	invalidUToken.BaseDenom = "u/uumee"
	invalidUToken.SymbolDenom = ""

	invalidReserveFactor := validToken()
	invalidReserveFactor.ReserveFactor = sdk.MustNewDecFromStr("-0.25")

	invalidCollateralWeight := validToken()
	invalidCollateralWeight.CollateralWeight = sdk.MustNewDecFromStr("50.00")

	invalidLiquidationThreshold := validToken()
	invalidLiquidationThreshold.LiquidationThreshold = sdk.MustNewDecFromStr("0.40")

	invalidBaseBorrowRate := validToken()
	invalidBaseBorrowRate.BaseBorrowRate = sdk.MustNewDecFromStr("-0.01")

	invalidKinkBorrowRate := validToken()
	invalidKinkBorrowRate.KinkBorrowRate = sdk.MustNewDecFromStr("-0.05")

	invalidMaxBorrowRate := validToken()
	invalidMaxBorrowRate.MaxBorrowRate = sdk.MustNewDecFromStr("-1.0")

	invalidKinkUtilization := validToken()
	invalidKinkUtilization.KinkUtilization = sdk.ZeroDec()

	invalidLiquidationIncentive := validToken()
	invalidLiquidationIncentive.LiquidationIncentive = sdk.MustNewDecFromStr("-0.05")

	invalidBlacklistedBorrow := validToken()
	invalidBlacklistedBorrow.EnableMsgBorrow = false
	invalidBlacklistedBorrow.Blacklist = true

	invalidBlacklistedSupply := validToken()
	invalidBlacklistedSupply.EnableMsgSupply = false
	invalidBlacklistedSupply.Blacklist = true

	invalidMaxCollateralShare := validToken()
	invalidMaxCollateralShare.MaxCollateralShare = sdk.MustNewDecFromStr("1.05")

	invalidMaxSupplyUtilization := validToken()
	invalidMaxSupplyUtilization.MaxSupplyUtilization = sdk.MustNewDecFromStr("1.05")

	invalidMinCollateralLiquidity := validToken()
	invalidMinCollateralLiquidity.MinCollateralLiquidity = sdk.MustNewDecFromStr("-0.05")

	testCases := map[string]struct {
		input     types.Token
		expectErr bool
	}{
		"valid token": {
			input: validToken(),
		},
		"invalid base token": {
			input:     invalidBaseToken,
			expectErr: true,
		},
		"invalid base token (utoken)": {
			input:     invalidUToken,
			expectErr: true,
		},
		"invalid reserve factor": {
			input:     invalidReserveFactor,
			expectErr: true,
		},
		"invalid collateral weight": {
			input:     invalidCollateralWeight,
			expectErr: true,
		},
		"invalid liquidation threshold": {
			input:     invalidLiquidationThreshold,
			expectErr: true,
		},
		"invalid base borrow rate": {
			input:     invalidBaseBorrowRate,
			expectErr: true,
		},
		"invalid kink borrow rate": {
			input:     invalidKinkBorrowRate,
			expectErr: true,
		},
		"invalid max borrow rate": {
			input:     invalidMaxBorrowRate,
			expectErr: true,
		},
		"invalid kink utilization rate": {
			input:     invalidKinkUtilization,
			expectErr: true,
		},
		"invalid liquidation incentive": {
			input:     invalidLiquidationIncentive,
			expectErr: true,
		},
		"blacklisted but supply enabled": {
			input:     invalidBlacklistedSupply,
			expectErr: true,
		},
		"blacklisted but borrow enabled": {
			input:     invalidBlacklistedBorrow,
			expectErr: true,
		},
		"invalid max collateral share": {
			input:     invalidMaxCollateralShare,
			expectErr: true,
		},
		"invalid max supply utilization": {
			input:     invalidMaxSupplyUtilization,
			expectErr: true,
		},
		"invalid min collateral liquidity": {
			input:     invalidMinCollateralLiquidity,
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
