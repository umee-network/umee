package keeper

import (
	"testing"

	"github.com/umee-network/umee/v4/util/coin"

	"github.com/cosmos/cosmos-sdk/codec"
	ctypes "github.com/cosmos/cosmos-sdk/codec/types"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/umee-network/umee/v4/x/metoken"
)

func TestKeeper_InitGenesis(t *testing.T) {
	interfaceRegistry := ctypes.NewInterfaceRegistry()
	marshaller := codec.NewProtoCodec(interfaceRegistry)

	ctx, keeper := initFullKeeper(t, marshaller)

	invalidRegistry := *metoken.DefaultGenesisState()
	invalidRegistry.Registry = []metoken.Index{
		metoken.NewIndex(sdk.NewCoin("token", sdkmath.ZeroInt()), metoken.Fee{}, nil),
	}

	invalidBalance := *metoken.DefaultGenesisState()
	invalidBalance.Balances = []metoken.IndexBalance{
		{
			MetokenSupply: sdk.Coin{
				Denom:  "test",
				Amount: sdkmath.ZeroInt(),
			},
			AssetBalances: nil,
		},
	}

	tcs := []struct {
		name   string
		g      metoken.GenesisState
		errMsg string
	}{
		{"valid genesis", *metoken.DefaultGenesisState(), ""},
		{
			"invalid registry",
			invalidRegistry,
			"meToken denom token should have the following format: me<TokenName>: invalid request",
		},
		{
			"invalid balances",
			invalidBalance,
			"meToken denom test should have the following format: me<TokenName>: invalid request",
		},
	}

	for _, tc := range tcs {
		t.Run(
			tc.name, func(t *testing.T) {
				if tc.errMsg != "" {
					assert.PanicsWithError(t, tc.errMsg, func() { keeper.InitGenesis(ctx, tc.g) })
				} else {
					assert.NotPanics(t, func() { keeper.InitGenesis(ctx, tc.g) })
				}
			},
		)
	}
}

func TestKeeper_ExportGenesis(t *testing.T) {
	interfaceRegistry := ctypes.NewInterfaceRegistry()
	marshaller := codec.NewProtoCodec(interfaceRegistry)

	ctx, keeper := initFullKeeper(t, marshaller)

	usdt := "USDT"
	usdt0 := coin.Zero(usdt)
	expectedGenesis := *metoken.DefaultGenesisState()
	expectedGenesis.Registry = []metoken.Index{
		{
			MetokenMaxSupply: sdk.NewCoin("meUSD", sdkmath.ZeroInt()),
			Fee: metoken.NewFee(
				sdk.MustNewDecFromStr("0.001"),
				sdk.MustNewDecFromStr("0.2"),
				sdk.MustNewDecFromStr("0.5"),
			),
			AcceptedAssets: []metoken.AcceptedAsset{
				metoken.NewAcceptedAsset(
					usdt,
					sdk.MustNewDecFromStr("0.2"),
					sdk.MustNewDecFromStr("1.0"),
				),
			},
		},
	}
	expectedGenesis.Balances = []metoken.IndexBalance{
		{
			MetokenSupply: coin.Zero("meUSD"),
			AssetBalances: []metoken.AssetBalance{
				metoken.NewAssetBalance(
					usdt0,
					usdt0,
					usdt0,
				),
			},
		},
	}
	expectedGenesis.NextRebalancingTime = 1683035091
	expectedGenesis.NextInterestClaimTime = 1683035600

	assert.NotPanics(t, func() { keeper.InitGenesis(ctx, expectedGenesis) })

	resultGenesis := keeper.ExportGenesis(ctx)

	assert.Equal(t, expectedGenesis, *resultGenesis)
}
