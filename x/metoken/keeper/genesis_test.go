package keeper

import (
	"testing"
	"time"

	"github.com/umee-network/umee/v6/x/metoken/mocks"

	"github.com/umee-network/umee/v6/util/coin"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/umee-network/umee/v6/x/metoken"
)

func TestKeeper_InitGenesis(t *testing.T) {
	keeper := initSimpleKeeper(t)

	invalidRegistry := *metoken.DefaultGenesisState()
	invalidRegistry.Registry = []metoken.Index{
		metoken.NewIndex("token", sdkmath.ZeroInt(), 6, metoken.Fee{}, nil),
	}

	invalidBalance := *metoken.DefaultGenesisState()
	invalidBalance.Balances = []metoken.IndexBalances{
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
			"meToken denom token should have the following format: me/<TokenName>: invalid request",
		},
		{
			"invalid balances",
			invalidBalance,
			"meToken denom test should have the following format: me/<TokenName>: invalid request",
		},
	}

	for _, tc := range tcs {
		t.Run(
			tc.name, func(t *testing.T) {
				if tc.errMsg != "" {
					assert.PanicsWithError(t, tc.errMsg, func() { keeper.InitGenesis(tc.g) })
				} else {
					assert.NotPanics(t, func() { keeper.InitGenesis(tc.g) })
				}
			},
		)
	}
}

func TestKeeper_ExportGenesis(t *testing.T) {
	keeper := initSimpleKeeper(t)

	usdt := "USDT"
	int0 := sdkmath.ZeroInt()
	expectedGenesis := *metoken.DefaultGenesisState()
	expectedGenesis.Registry = []metoken.Index{
		{
			Denom:     mocks.MeUSDDenom,
			MaxSupply: sdkmath.ZeroInt(),
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
	expectedGenesis.Balances = []metoken.IndexBalances{
		{
			MetokenSupply: coin.Zero(mocks.MeUSDDenom),
			AssetBalances: []metoken.AssetBalance{
				metoken.NewAssetBalance(
					usdt,
					int0,
					int0,
					int0,
					int0,
				),
			},
		},
	}
	expectedGenesis.NextRebalancingTime = time.UnixMilli(time.Now().UnixMilli())
	expectedGenesis.NextInterestClaimTime = time.UnixMilli(time.Now().UnixMilli())

	assert.NotPanics(t, func() { keeper.InitGenesis(expectedGenesis) })

	resultGenesis := keeper.ExportGenesis()

	assert.Equal(t, expectedGenesis, *resultGenesis)
}
