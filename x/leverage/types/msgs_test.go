package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	tassert "github.com/stretchr/testify/assert"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v6/tests/accs"
	"github.com/umee-network/umee/v6/util/checkers"
	"github.com/umee-network/umee/v6/x/leverage/types"
)

func TestMsgGovSetParams(t *testing.T) {
	tcs := []struct {
		msg  types.MsgGovSetParams
		err  string
		name string
	}{
		{
			types.MsgGovSetParams{
				Params: types.DefaultParams(),
			},
			"empty address",
			"no authority",
		},
		{
			types.MsgGovSetParams{
				checkers.GovModuleAddr,
				types.Params{
					CompleteLiquidationThreshold: sdkmath.LegacyZeroDec(),
					SmallLiquidationSize:         sdkmath.LegacyZeroDec(),
					OracleRewardFactor:           sdkmath.LegacyZeroDec(),
					MinimumCloseFactor:           sdkmath.LegacyZeroDec(),
					DirectLiquidationFee:         sdkmath.LegacyZeroDec(),
				},
			},
			"complete liquidation threshold must be positive",
			"invalid params",
		},
		{
			types.MsgGovSetParams{
				checkers.GovModuleAddr,
				types.DefaultParams(),
			},
			"",
			"valid",
		},
	}

	for _, tc := range tcs {
		err := tc.msg.ValidateBasic()
		if tc.err == "" {
			assert.NilError(t, err, tc.name)
		} else {
			assert.ErrorContains(t, err, tc.err, tc.name)
		}
	}
}

func TestMsgGovUpdateRegistryValidateBasic(t *testing.T) {
	validToken := types.Token{
		BaseDenom:              "uumee",
		SymbolDenom:            "UMEE",
		Exponent:               6,
		ReserveFactor:          sdkmath.LegacyMustNewDecFromStr("0.2"),
		CollateralWeight:       sdkmath.LegacyMustNewDecFromStr("0.25"),
		LiquidationThreshold:   sdkmath.LegacyMustNewDecFromStr("0.5"),
		BaseBorrowRate:         sdkmath.LegacyMustNewDecFromStr("0.02"),
		KinkBorrowRate:         sdkmath.LegacyMustNewDecFromStr("0.22"),
		MaxBorrowRate:          sdkmath.LegacyMustNewDecFromStr("1.52"),
		KinkUtilization:        sdkmath.LegacyMustNewDecFromStr("0.8"),
		LiquidationIncentive:   sdkmath.LegacyMustNewDecFromStr("0.1"),
		EnableMsgSupply:        true,
		EnableMsgBorrow:        true,
		Blacklist:              false,
		MaxCollateralShare:     sdkmath.LegacyMustNewDecFromStr("1"),
		MaxSupplyUtilization:   sdkmath.LegacyMustNewDecFromStr("0.9"),
		MinCollateralLiquidity: sdkmath.LegacyMustNewDecFromStr("0"),
		MaxSupply:              sdkmath.NewInt(100_000_000000),
		HistoricMedians:        24,
	}
	duplicateBaseDenom := validToken
	duplicateBaseDenom.SymbolDenom = "umee2"

	invalidSymbol := validToken
	invalidSymbol.SymbolDenom = ""

	newMsg := func(addTokens, updateTokens []types.Token) types.MsgGovUpdateRegistry {
		return types.MsgGovUpdateRegistry{
			Authority:    checkers.GovModuleAddr,
			Description:  "",
			AddTokens:    addTokens,
			UpdateTokens: updateTokens,
		}
	}

	validMsg := newMsg([]types.Token{validToken}, nil)
	validMsg2 := newMsg([]types.Token{validToken}, nil)
	validMsg2.Authority = accs.Alice.String()
	validMsg2.Description = "some description"

	tcs := []struct {
		name string
		q    types.MsgGovUpdateRegistry
		err  string
	}{
		{"no authority", types.MsgGovUpdateRegistry{}, "empty address"},
		{
			"duplicated base_denom add",
			newMsg([]types.Token{validToken, duplicateBaseDenom}, nil),
			"duplicate token",
		},
		{
			"duplicated update token",
			newMsg(nil, []types.Token{validToken, duplicateBaseDenom}),
			"duplicate token",
		},
		{
			"invalid add token",
			newMsg([]types.Token{invalidSymbol}, nil),
			"symbol_denom: invalid denom",
		},
		{
			"invalid update token",
			newMsg(nil, []types.Token{invalidSymbol}),
			"symbol_denom: invalid denom",
		},
		{
			"empty add and update tokens", newMsg(nil, nil),
			"empty add and update tokens",
		},
		{
			"valid", validMsg, "",
		},
		{
			"valid: non gov module address", validMsg2, "",
		},
	}

	for i, tc := range tcs {
		t.Run(
			tc.name, func(t *testing.T) {
				err := tc.q.ValidateBasic()
				if tc.err == "" {
					assert.NilError(t, err, "test: %v", i)
				} else {
					assert.ErrorContains(t, err, tc.err, "test: %v", i)
				}
			},
		)
	}
}

func TestMsgGovUpdateRegistryOtherFunctionality(t *testing.T) {
	umee := types.Token{
		BaseDenom:              "uumee",
		SymbolDenom:            "UMEE",
		Exponent:               6,
		ReserveFactor:          sdkmath.LegacyMustNewDecFromStr("0.2"),
		CollateralWeight:       sdkmath.LegacyMustNewDecFromStr("0.25"),
		LiquidationThreshold:   sdkmath.LegacyMustNewDecFromStr("0.25"),
		BaseBorrowRate:         sdkmath.LegacyMustNewDecFromStr("0.02"),
		KinkBorrowRate:         sdkmath.LegacyMustNewDecFromStr("0.22"),
		MaxBorrowRate:          sdkmath.LegacyMustNewDecFromStr("1.52"),
		KinkUtilization:        sdkmath.LegacyMustNewDecFromStr("0.8"),
		LiquidationIncentive:   sdkmath.LegacyMustNewDecFromStr("0.1"),
		EnableMsgSupply:        true,
		EnableMsgBorrow:        true,
		Blacklist:              false,
		MaxCollateralShare:     sdkmath.LegacyMustNewDecFromStr("1"),
		MaxSupplyUtilization:   sdkmath.LegacyMustNewDecFromStr("0.9"),
		MinCollateralLiquidity: sdkmath.LegacyMustNewDecFromStr("0"),
		MaxSupply:              sdkmath.NewInt(100_000_000000),
		HistoricMedians:        24,
	}
	msg := types.NewMsgGovUpdateRegistry(
		checkers.GovModuleAddr,
		"",
		[]types.Token{umee}, []types.Token{},
	)

	expResult := `authority: umee10d07y265gmmuvt4z0w9aw880jnsr700jg5w6jp
description: ""
addtokens: []
updatetokens:
    - base_denom: uumee
      reserve_factor: "0.200000000000000000"
      collateral_weight: "0.250000000000000000"
      liquidation_threshold: "0.250000000000000000"
      base_borrow_rate: "0.020000000000000000"
      kink_borrow_rate: "0.220000000000000000"
      max_borrow_rate: "1.520000000000000000"
      kink_utilization: "0.800000000000000000"
      liquidation_incentive: "0.100000000000000000"
      symbol_denom: UMEE
      exponent: 6
      enable_msg_supply: true
      enable_msg_borrow: true
      blacklist: false
      max_collateral_share: "1.000000000000000000"
      max_supply_utilization: "0.900000000000000000"
      min_collateral_liquidity: "0.000000000000000000"
      max_supply: "100000000000"
      historic_medians: 24
`
	assert.Equal(t, expResult, msg.String())
	tassert.NotNil(t, msg.GetSignBytes(), "sign byte shouldn't be nil")
	tassert.NotEmpty(t, msg.GetSigners(), "signers shouldn't be empty")
}

// TODO : tests for MsgGovUpdateSpecialAssets
