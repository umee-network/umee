package types_test

import (
	"testing"

	"github.com/umee-network/umee/v4/x/leverage/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	tassert "github.com/stretchr/testify/assert"

	"gotest.tools/v3/assert"
)

func TestMsgGovUpdateRegistryValidateBasic(t *testing.T) {
	tcs := []struct {
		name string
		q    types.MsgGovUpdateRegistry
		err  string
	}{
		{"no authority", types.MsgGovUpdateRegistry{}, "invalid authority"},
		{
			"duplicated add token", types.MsgGovUpdateRegistry{
				Authority:   authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Title:       "Title",
				Description: "Description",
				AddTokens: []types.Token{
					{BaseDenom: "uumee"},
					{BaseDenom: "uumee"},
				},
			}, "duplicate token",
		},
		{
			"invalid add token", types.MsgGovUpdateRegistry{
				Authority:   authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Title:       "Title",
				Description: "Description",
				AddTokens: []types.Token{
					{BaseDenom: "uumee"},
				},
			}, "invalid denom",
		},
		{
			"duplicated update token", types.MsgGovUpdateRegistry{
				Authority:   authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Title:       "Title",
				Description: "Description",
				UpdateTokens: []types.Token{
					{BaseDenom: "uumee"},
					{BaseDenom: "uumee"},
				},
			}, "duplicate token",
		},
		{
			"invalid update token", types.MsgGovUpdateRegistry{
				Authority:   authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Title:       "Title",
				Description: "Description",
				UpdateTokens: []types.Token{
					{BaseDenom: "uumee"},
				},
			}, "invalid denom",
		},
		{
			"empty add and update tokens", types.MsgGovUpdateRegistry{
				Authority:   authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Title:       "Title",
				Description: "Description",
			}, "empty add and update tokens",
		},
		{
			"valid", types.MsgGovUpdateRegistry{
				Authority:   authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Title:       "Title",
				Description: "Description",
				AddTokens: []types.Token{
					{
						BaseDenom:              "uumee",
						SymbolDenom:            "UMEE",
						Exponent:               6,
						ReserveFactor:          sdk.MustNewDecFromStr("0.2"),
						CollateralWeight:       sdk.MustNewDecFromStr("0.25"),
						LiquidationThreshold:   sdk.MustNewDecFromStr("0.25"),
						BaseBorrowRate:         sdk.MustNewDecFromStr("0.02"),
						KinkBorrowRate:         sdk.MustNewDecFromStr("0.22"),
						MaxBorrowRate:          sdk.MustNewDecFromStr("1.52"),
						KinkUtilization:        sdk.MustNewDecFromStr("0.8"),
						LiquidationIncentive:   sdk.MustNewDecFromStr("0.1"),
						EnableMsgSupply:        true,
						EnableMsgBorrow:        true,
						Blacklist:              false,
						MaxCollateralShare:     sdk.MustNewDecFromStr("1"),
						MaxSupplyUtilization:   sdk.MustNewDecFromStr("0.9"),
						MinCollateralLiquidity: sdk.MustNewDecFromStr("0"),
						MaxSupply:              sdk.NewInt(100_000_000000),
						HistoricMedians:        24,
					},
				},
			}, "",
		},
	}

	for _, tc := range tcs {
		t.Run(
			tc.name, func(t *testing.T) {
				err := tc.q.ValidateBasic()
				if tc.err == "" {
					assert.NilError(t, err)
				} else {
					assert.ErrorContains(t, err, tc.err)
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
		ReserveFactor:          sdk.MustNewDecFromStr("0.2"),
		CollateralWeight:       sdk.MustNewDecFromStr("0.25"),
		LiquidationThreshold:   sdk.MustNewDecFromStr("0.25"),
		BaseBorrowRate:         sdk.MustNewDecFromStr("0.02"),
		KinkBorrowRate:         sdk.MustNewDecFromStr("0.22"),
		MaxBorrowRate:          sdk.MustNewDecFromStr("1.52"),
		KinkUtilization:        sdk.MustNewDecFromStr("0.8"),
		LiquidationIncentive:   sdk.MustNewDecFromStr("0.1"),
		EnableMsgSupply:        true,
		EnableMsgBorrow:        true,
		Blacklist:              false,
		MaxCollateralShare:     sdk.MustNewDecFromStr("1"),
		MaxSupplyUtilization:   sdk.MustNewDecFromStr("0.9"),
		MinCollateralLiquidity: sdk.MustNewDecFromStr("0"),
		MaxSupply:              sdk.NewInt(100_000_000000),
		HistoricMedians:        24,
	}
	msg := types.NewMsgUpdateRegistry(
		authtypes.NewModuleAddress(govtypes.ModuleName).String(), "title", "description",
		[]types.Token{umee}, []types.Token{},
	)

	tassert.NotEmpty(t, msg.String(), "result shouldn't be empty")
	tassert.NotNil(t, msg.GetSignBytes(), "sign byte shouldn't be nil")
	tassert.NotEmpty(t, msg.GetSigners(), "signers shouldn't be empty")
}
