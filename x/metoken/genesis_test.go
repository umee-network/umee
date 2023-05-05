package metoken

import (
	"testing"

	"github.com/umee-network/umee/v4/util/coin"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"
)

func TestGenesisState_Validate(t *testing.T) {
	invalidRegistry := *DefaultGenesisState()
	invalidRegistry.Registry = []Index{
		NewIndex(sdk.NewCoin("token", sdkmath.ZeroInt()), Fee{}, nil),
	}

	invalidBalance := *DefaultGenesisState()
	invalidBalance.Balances = []IndexBalance{
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
		g      GenesisState
		errMsg string
	}{
		{"default genesis", *DefaultGenesisState(), ""},
		{
			"invalid registry",
			invalidRegistry,
			"meToken denom token should have the following format: me<TokenName>",
		},
		{
			"invalid balances",
			invalidBalance,
			"meToken denom test should have the following format: me<TokenName>",
		},
	}

	for _, tc := range tcs {
		t.Run(
			tc.name, func(t *testing.T) {
				err := tc.g.Validate()
				if tc.errMsg != "" {
					assert.ErrorContains(t, err, tc.errMsg)
				} else {
					assert.NilError(t, err)
				}
			},
		)
	}
}

func TestIndexBalance_Validate(t *testing.T) {
	usdt0 := coin.Zero("USDT")

	invalidMetokenDenom := validIndexBalance()
	invalidMetokenDenom.MetokenSupply = sdk.NewCoin("test", sdkmath.ZeroInt())

	invalidMetokenAmount := validIndexBalance()
	invalidMetokenAmount.MetokenSupply = coin.Negative1("meToken")

	invalidAssetInLeverage := validIndexBalance()
	invalidAssetInLeverage.AssetBalances[0] = NewAssetBalance(
		coin.Negative1("USDT"),
		sdk.Coin{},
		sdk.Coin{},
	)

	invalidAssetInReserves := validIndexBalance()
	invalidAssetInReserves.AssetBalances[0] = NewAssetBalance(
		usdt0,
		coin.Negative1("USDT"),
		sdk.Coin{},
	)

	invalidAssetInFees := validIndexBalance()
	invalidAssetInFees.AssetBalances[0] = NewAssetBalance(
		usdt0,
		usdt0,
		coin.Negative1("USDT"),
	)

	differentAssetsInBalance := validIndexBalance()
	differentAssetsInBalance.AssetBalances[0].Fees = coin.Zero("USDC")

	tcs := []struct {
		name   string
		ib     IndexBalance
		errMsg string
	}{
		{"valid index balance", validIndexBalance(), ""},
		{
			"invalid meToken denom",
			invalidMetokenDenom,
			"meToken denom test should have the following format: me<TokenName>",
		},
		{
			"invalid meToken amount",
			invalidMetokenAmount,
			"negative coin amount",
		},
		{
			"invalid assetInLeverage",
			invalidAssetInLeverage,
			"negative coin amount",
		},
		{
			"invalid assetInReserves",
			invalidAssetInReserves,
			"negative coin amount",
		},
		{
			"invalid assetInFee",
			invalidAssetInFees,
			"negative coin amount",
		},
		{
			"invalid assets denom",
			differentAssetsInBalance,
			"different assets in the Index balance",
		},
		{
			"valid index balance",
			NewIndexBalance(
				sdk.NewCoin("meUSD", sdkmath.ZeroInt()), []AssetBalance{
					NewAssetBalance(
						usdt0,
						usdt0,
						usdt0,
					),
				},
			),
			"",
		},
	}

	for _, tc := range tcs {
		t.Run(
			tc.name, func(t *testing.T) {
				err := tc.ib.Validate()
				if tc.errMsg != "" {
					assert.ErrorContains(t, err, tc.errMsg)
				} else {
					assert.NilError(t, err)
				}
			},
		)
	}
}

func validIndexBalance() IndexBalance {
	usdt0 := coin.Zero("USDT")
	return IndexBalance{
		MetokenSupply: coin.Zero("meUSD"),
		AssetBalances: []AssetBalance{
			NewAssetBalance(
				usdt0,
				usdt0,
				usdt0,
			),
		},
	}
}
