package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/app"
	"github.com/umee-network/umee/v5/util/coin"
	"github.com/umee-network/umee/v5/x/metoken"
	"github.com/umee-network/umee/v5/x/metoken/keeper"
	"github.com/umee-network/umee/v5/x/metoken/mocks"
)

func TestRebalanceReserves(t *testing.T) {
	index := mocks.StableIndex(mocks.MeUSDDenom)

	s := initTestSuite(t, nil, nil)
	msgServer, ctx, app := s.msgServer, s.ctx, s.app

	_, err := msgServer.GovUpdateRegistry(
		ctx, &metoken.MsgGovUpdateRegistry{
			Authority:   app.GovKeeper.GetGovernanceAccount(s.ctx).GetAddress().String(),
			AddIndex:    []metoken.Index{index},
			UpdateIndex: nil,
		},
	)
	require.NoError(t, err)

	// create and fund a user with 1000 USDT, 1000 USDC and 1000 IST
	user := s.newAccount(
		t,
		coin.New(mocks.USDTBaseDenom, 1000_000000),
		coin.New(mocks.USDCBaseDenom, 1000_000000),
		coin.New(mocks.ISTBaseDenom, 1000_000000),
	)

	// swap 547 USDT, 200 USDC and 740 IST to have an initial meUSD balance
	swaps := []*metoken.MsgSwap{
		{
			User:         user.String(),
			Asset:        sdk.NewCoin(mocks.USDTBaseDenom, sdkmath.NewInt(547_000000)),
			MetokenDenom: mocks.MeUSDDenom,
		},
		{
			User:         user.String(),
			Asset:        sdk.NewCoin(mocks.USDCBaseDenom, sdkmath.NewInt(200_000000)),
			MetokenDenom: mocks.MeUSDDenom,
		},
		{
			User:         user.String(),
			Asset:        sdk.NewCoin(mocks.ISTBaseDenom, sdkmath.NewInt(740_000000)),
			MetokenDenom: mocks.MeUSDDenom,
		},
	}

	for _, swap := range swaps {
		_, err := msgServer.Swap(ctx, swap)
		require.NoError(t, err)
	}

	k := app.MetokenKeeperB.Keeper(&ctx)
	// check the initial balances are balanced
	checkBalances(t, ctx, app, k, index.Denom, true)

	// change index setting modifying the reserve_portion
	// usdt_reserve_portion from 0.2 to 0.25
	usdtReservePortion := sdk.MustNewDecFromStr("0.25")
	i, usdtSettings := index.AcceptedAsset(mocks.USDTBaseDenom)
	require.True(t, i >= 0)
	usdtSettings.ReservePortion = usdtReservePortion
	index.SetAcceptedAsset(usdtSettings)

	// usdc_reserve_portion from 0.2 to 0.5
	usdcReservePortion := sdk.MustNewDecFromStr("0.5")
	i, usdcSettings := index.AcceptedAsset(mocks.USDCBaseDenom)
	require.True(t, i >= 0)
	usdcSettings.ReservePortion = usdcReservePortion
	index.SetAcceptedAsset(usdcSettings)

	// ist_reserve_portion from 0.2 to 0.035
	istReservePortion := sdk.MustNewDecFromStr("0.035")
	i, istSettings := index.AcceptedAsset(mocks.ISTBaseDenom)
	require.True(t, i >= 0)
	istSettings.ReservePortion = istReservePortion
	index.SetAcceptedAsset(istSettings)

	// update index
	_, err = msgServer.GovUpdateRegistry(
		ctx, &metoken.MsgGovUpdateRegistry{
			Authority:   app.GovKeeper.GetGovernanceAccount(ctx).GetAddress().String(),
			AddIndex:    nil,
			UpdateIndex: []metoken.Index{index},
		},
	)
	require.NoError(t, err)

	// confirm now the balances are unbalanced
	checkBalances(t, ctx, app, k, index.Denom, false)

	err = k.RebalanceReserves()
	require.NoError(t, err)

	// confirm the balances are good now
	checkBalances(t, ctx, app, k, index.Denom, true)
}

func checkBalances(
	t *testing.T,
	ctx sdk.Context,
	app *app.UmeeApp,
	k keeper.Keeper,
	meTokenDenom string,
	shouldBeBalanced bool,
) {
	meTokenAddr := keeper.ModuleAddr()
	// get index
	index, err := k.RegisteredIndex(meTokenDenom)
	require.NoError(t, err)

	// get index balances
	balances, err := k.IndexBalances(index.Denom)
	require.NoError(t, err)

	// get x/bank balance
	bankBalance := app.BankKeeper.GetAllBalances(ctx, meTokenAddr)

	for _, balance := range balances.AssetBalances {
		// confirm the index is balanced as required in the configuration
		i, assetSettings := index.AcceptedAsset(balance.Denom)
		require.True(t, i >= 0)

		desiredReserves := assetSettings.ReservePortion.MulInt(balance.AvailableSupply()).TruncateInt()
		require.Equal(t, shouldBeBalanced, desiredReserves.Equal(balance.Reserved))

		// check the supply in x/leverage and confirm is the same
		allSupplied, err := app.LeverageKeeper.GetAllSupplied(ctx, meTokenAddr)
		require.NoError(t, err)

		found, assetSupplied := allSupplied.Find(balance.Denom)
		require.True(t, found)
		require.True(t, assetSupplied.Amount.Equal(balance.Leveraged))

		// confirm balance in x/bank and x/module state is the same
		found, bBalance := bankBalance.Find(balance.Denom)
		require.True(t, found)
		require.True(t, bBalance.Amount.Equal(balance.Interest.Add(balance.Reserved).Add(balance.Fees)))
	}
}
