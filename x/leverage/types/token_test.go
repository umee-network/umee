package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v6/x/leverage/types"
)

func validToken() types.Token {
	return types.Token{
		BaseDenom:              "uumee",
		SymbolDenom:            "umee",
		Exponent:               6,
		ReserveFactor:          sdkmath.LegacyMustNewDecFromStr("0.25"),
		CollateralWeight:       sdkmath.LegacyMustNewDecFromStr("0.5"),
		LiquidationThreshold:   sdkmath.LegacyMustNewDecFromStr("0.51"),
		BaseBorrowRate:         sdkmath.LegacyMustNewDecFromStr("0.01"),
		KinkBorrowRate:         sdkmath.LegacyMustNewDecFromStr("0.05"),
		MaxBorrowRate:          sdkmath.LegacyMustNewDecFromStr("1"),
		KinkUtilization:        sdkmath.LegacyMustNewDecFromStr("0.75"),
		LiquidationIncentive:   sdkmath.LegacyMustNewDecFromStr("0.05"),
		EnableMsgSupply:        true,
		EnableMsgBorrow:        true,
		Blacklist:              false,
		MaxCollateralShare:     sdkmath.LegacyMustNewDecFromStr("1"),
		MaxSupplyUtilization:   sdkmath.LegacyMustNewDecFromStr("1"),
		MinCollateralLiquidity: sdkmath.LegacyMustNewDecFromStr("1"),
		MaxSupply:              sdkmath.NewInt(1000),
		HistoricMedians:        24,
	}
}

func TestUpdateRegistryProposalString(t *testing.T) {
	token := validToken()
	token.ReserveFactor = sdkmath.LegacyNewDec(40)
	p := types.MsgGovUpdateRegistry{
		Authority:   "authority",
		Description: "test",
		AddTokens:   []types.Token{token},
	}
	expected := `authority: authority
description: test
addtokens:
    - base_denom: uumee
      reserve_factor: "40.000000000000000000"
      collateral_weight: "0.500000000000000000"
      liquidation_threshold: "0.510000000000000000"
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

func TestTokenValidate(t *testing.T) {
	invalidBaseToken := validToken()
	invalidBaseToken.BaseDenom = "$$"
	invalidBaseToken.SymbolDenom = ""

	invalidUToken := validToken()
	invalidUToken.BaseDenom = "u/uumee"
	invalidUToken.SymbolDenom = ""

	invalidReserveFactor := validToken()
	invalidReserveFactor.ReserveFactor = sdkmath.LegacyMustNewDecFromStr("-0.25")

	invalidCollateralWeight := validToken()
	invalidCollateralWeight.CollateralWeight = sdkmath.LegacyMustNewDecFromStr("1.1")

	invalidCollateralWeight2 := validToken()
	invalidCollateralWeight2.CollateralWeight = sdkmath.LegacyOneDec()

	invalidLiquidationThreshold := validToken()
	invalidLiquidationThreshold.LiquidationThreshold = sdkmath.LegacyMustNewDecFromStr("-0.25")

	invalidLiquidationThreshold2 := validToken()
	invalidLiquidationThreshold2.LiquidationThreshold = sdkmath.LegacyOneDec()

	invalidBaseBorrowRate := validToken()
	invalidBaseBorrowRate.BaseBorrowRate = sdkmath.LegacyMustNewDecFromStr("-0.01")

	invalidKinkBorrowRate := validToken()
	invalidKinkBorrowRate.KinkBorrowRate = sdkmath.LegacyMustNewDecFromStr("-0.05")

	invalidMaxBorrowRate := validToken()
	invalidMaxBorrowRate.MaxBorrowRate = sdkmath.LegacyMustNewDecFromStr("-1.0")

	invalidKinkUtilization := validToken()
	invalidKinkUtilization.KinkUtilization = sdkmath.LegacyZeroDec()

	invalidLiquidationIncentive := validToken()
	invalidLiquidationIncentive.LiquidationIncentive = sdkmath.LegacyMustNewDecFromStr("-0.05")

	invalidBlacklistedBorrow := validToken()
	invalidBlacklistedBorrow.EnableMsgBorrow = false
	invalidBlacklistedBorrow.Blacklist = true

	invalidMaxCollateralShare := validToken()
	invalidMaxCollateralShare.MaxCollateralShare = sdkmath.LegacyMustNewDecFromStr("1.05")

	invalidMaxSupplyUtilization := validToken()
	invalidMaxSupplyUtilization.MaxSupplyUtilization = sdkmath.LegacyMustNewDecFromStr("1.05")

	invalidMinCollateralLiquidity := validToken()
	invalidMinCollateralLiquidity.MinCollateralLiquidity = sdkmath.LegacyMustNewDecFromStr("-0.05")

	invalidMaxSupply1 := validToken()
	invalidMaxSupply1.MaxSupply = sdkmath.NewInt(-1)

	validMaxSupply1 := validToken()
	validMaxSupply1.MaxSupply = sdkmath.NewInt(0)
	validMaxSupply1.EnableMsgSupply = false

	validMaxSupply2 := validToken()
	validMaxSupply2.MaxSupply = sdkmath.NewInt(0)

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
		"invalid collateral weight2": {
			input:     invalidCollateralWeight2,
			expectErr: true,
		},
		"invalid liquidation threshold": {
			input:     invalidLiquidationThreshold,
			expectErr: true,
		},
		"invalid liquidation threshold2": {
			input:     invalidLiquidationThreshold2,
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
