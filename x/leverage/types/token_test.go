package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v4/x/leverage/types"
)

func TestToTokenDenom(t *testing.T) {
	// Turns uToken denoms into base tokens
	assert.Equal(t, "uumee", types.ToTokenDenom("u/uumee"))
	assert.Equal(t, "ibc/abcd", types.ToTokenDenom("u/ibc/abcd"))

	// Empty return for base tokens
	assert.Equal(t, "", types.ToTokenDenom("uumee"))
	assert.Equal(t, "", types.ToTokenDenom("ibc/abcd"))

	// Empty return on repreated prefix
	assert.Equal(t, "", types.ToTokenDenom("u/u/abcd"))

	// Edge cases
	assert.Equal(t, "", types.ToTokenDenom("u/"))
	assert.Equal(t, "", types.ToTokenDenom(""))
}

func TestToUTokenDenom(t *testing.T) {
	// Turns base token denoms into base uTokens
	assert.Equal(t, "u/uumee", types.ToUTokenDenom("uumee"))
	assert.Equal(t, "u/ibc/abcd", types.ToUTokenDenom("ibc/abcd"))

	// Empty return for uTokens
	assert.Equal(t, "", types.ToUTokenDenom("u/uumee"))
	assert.Equal(t, "", types.ToUTokenDenom("u/ibc/abcd"))

	// Edge cases
	assert.Equal(t, "u/", types.ToUTokenDenom(""))
}

func validToken() types.Token {
	return types.Token{
		BaseDenom:              "uumee",
		SymbolDenom:            "umee",
		Exponent:               6,
		ReserveFactor:          sdk.MustNewDecFromStr("0.25"),
		CollateralWeight:       sdk.MustNewDecFromStr("0.5"),
		LiquidationThreshold:   sdk.MustNewDecFromStr("0.5"),
		BaseBorrowRate:         sdk.MustNewDecFromStr("0.01"),
		KinkBorrowRate:         sdk.MustNewDecFromStr("0.05"),
		MaxBorrowRate:          sdk.MustNewDecFromStr("1"),
		KinkUtilization:        sdk.MustNewDecFromStr("0.75"),
		LiquidationIncentive:   sdk.MustNewDecFromStr("0.05"),
		EnableMsgSupply:        true,
		EnableMsgBorrow:        true,
		Blacklist:              false,
		MaxCollateralShare:     sdk.MustNewDecFromStr("1"),
		MaxSupplyUtilization:   sdk.MustNewDecFromStr("1"),
		MinCollateralLiquidity: sdk.MustNewDecFromStr("1"),
		MaxSupply:              sdk.NewInt(1000),
		HistoricMedians:        24,
	}
}

func TestUpdateRegistryProposal_String(t *testing.T) {
	token := validToken()
	token.ReserveFactor = sdk.NewDec(40)
	p := types.MsgGovUpdateRegistry{
		Authority:   "authority",
		Title:       "test",
		Description: "test",
		AddTokens:   []types.Token{token},
	}
	expected := `authority: authority
title: test
description: test
addtokens:
    - base_denom: uumee
      reserve_factor: "40.000000000000000000"
      collateral_weight: "0.500000000000000000"
      liquidation_threshold: "0.500000000000000000"
      base_borrow_rate: "0.010000000000000000"
      kink_borrow_rate: "0.050000000000000000"
      max_borrow_rate: "1.000000000000000000"
      kink_utilization: "0.750000000000000000"
      liquidation_incentive: "0.050000000000000000"
      symbol_denom: umee
      exponent: 6
      enable_msg_supply: true
      enable_msg_borrow: true
      blacklist: false
      max_collateral_share: "1.000000000000000000"
      max_supply_utilization: "1.000000000000000000"
      min_collateral_liquidity: "1.000000000000000000"
      max_supply: "1000"
      historic_medians: 24
updatetokens: []
`
	assert.Equal(t, expected, p.String())
}

func TestToken_Validate(t *testing.T) {
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

	invalidMaxCollateralShare := validToken()
	invalidMaxCollateralShare.MaxCollateralShare = sdk.MustNewDecFromStr("1.05")

	invalidMaxSupplyUtilization := validToken()
	invalidMaxSupplyUtilization.MaxSupplyUtilization = sdk.MustNewDecFromStr("1.05")

	invalidMinCollateralLiquidity := validToken()
	invalidMinCollateralLiquidity.MinCollateralLiquidity = sdk.MustNewDecFromStr("-0.05")

	invalidMaxSupply1 := validToken()
	invalidMaxSupply1.MaxSupply = sdk.NewInt(-1)

	validMaxSupply1 := validToken()
	validMaxSupply1.MaxSupply = sdk.NewInt(0)
	validMaxSupply1.EnableMsgSupply = false

	validMaxSupply2 := validToken()
	validMaxSupply2.MaxSupply = sdk.NewInt(0)

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
		"invalid max supply (negative)": {
			input:     invalidMaxSupply1,
			expectErr: true,
		},
		"valid max supply (enable_msg_supply=false)": {
			input:     validMaxSupply1,
			expectErr: false,
		},
		"valid max supply (enable_msg_supply=true)": {
			input:     validMaxSupply2,
			expectErr: false,
		},
	}

	for name, tc := range testCases {
		tc := tc

		t.Run(name, func(t *testing.T) {
			err := tc.input.Validate()
			if tc.expectErr {
				assert.Error(t, err, err.Error())
			} else {
				assert.NilError(t, err)
			}
		})
	}
}
