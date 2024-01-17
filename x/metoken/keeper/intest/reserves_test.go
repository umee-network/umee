package intest

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v6/app"
	"github.com/umee-network/umee/v6/util/checkers"
	"github.com/umee-network/umee/v6/util/coin"
	"github.com/umee-network/umee/v6/x/metoken"
	"github.com/umee-network/umee/v6/x/metoken/keeper"
	"github.com/umee-network/umee/v6/x/metoken/mocks"
)

func TestRebalanceReserves(t *testing.T) {
	index := mocks.StableIndex(mocks.MeUSDDenom)

	s := initTestSuite(t, nil, nil)
	msgServer, ctx, app := s.msgServer, s.ctx, s.app

	_, err := msgServer.GovUpdateRegistry(
		ctx, &metoken.MsgGovUpdateRegistry{
			Authority:   checkers.GovModuleAddr,
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
	checkBalances(t, ctx, app, k, index.Denom, true, true)

	// change index setting modifying the reserve_portion
	// usdt_reserve_portion from 0.2 to 0.25
	usdtReservePortion := sdkmath.LegacyMustNewDecFromStr("0.25")
	usdtSettings, i := index.AcceptedAsset(mocks.USDTBaseDenom)
	require.True(t, i >= 0)
	usdtSettings.ReservePortion = usdtReservePortion
	index.SetAcceptedAsset(usdtSettings)

	// usdc_reserve_portion from 0.2 to 0.5
	usdcReservePortion := sdkmath.LegacyMustNewDecFromStr("0.5")
	usdcSettings, i := index.AcceptedAsset(mocks.USDCBaseDenom)
	require.True(t, i >= 0)
	usdcSettings.ReservePortion = usdcReservePortion
	index.SetAcceptedAsset(usdcSettings)

	// ist_reserve_portion from 0.2 to 0.035
	istReservePortion := sdkmath.LegacyMustNewDecFromStr("0.035")
	istSettings, i := index.AcceptedAsset(mocks.ISTBaseDenom)
	require.True(t, i >= 0)
	istSettings.ReservePortion = istReservePortion
	index.SetAcceptedAsset(istSettings)

	// update index
	_, err = msgServer.GovUpdateRegistry(
		ctx, &metoken.MsgGovUpdateRegistry{
			Authority:   checkers.GovModuleAddr,
			AddIndex:    nil,
			UpdateIndex: []metoken.Index{index},
		},
	)
	require.NoError(t, err)

	// confirm now the balances are unbalanced
	checkBalances(t, ctx, app, k, index.Denom, false, true)

	err = k.RebalanceReserves()
	require.NoError(t, err)

	// confirm the balances are good now
	checkBalances(t, ctx, app, k, index.Denom, true, true)

	// change index setting modifying the reserve_portion
	// usdt_reserve_portion from 0.25 to 1.0
	usdtReservePortion = sdk.MustNewDecFromStr("1.0")
	usdtSettings, i = index.AcceptedAsset(mocks.USDTBaseDenom)
	require.True(t, i >= 0)
	usdtSettings.ReservePortion = usdtReservePortion
	index.SetAcceptedAsset(usdtSettings)

	// usdc_reserve_portion from 0.5 to 1.0
	usdcReservePortion = sdk.MustNewDecFromStr("1.0")
	usdcSettings, i = index.AcceptedAsset(mocks.USDCBaseDenom)
	require.True(t, i >= 0)
	usdcSettings.ReservePortion = usdcReservePortion
	index.SetAcceptedAsset(usdcSettings)

	// ist_reserve_portion from 0.035 to 1.0
	istReservePortion = sdk.MustNewDecFromStr("1.0")
	istSettings, i = index.AcceptedAsset(mocks.ISTBaseDenom)
	require.True(t, i >= 0)
	istSettings.ReservePortion = istReservePortion
	index.SetAcceptedAsset(istSettings)

	// update index
	_, err = msgServer.GovUpdateRegistry(
		ctx, &metoken.MsgGovUpdateRegistry{
			Authority:   checkers.GovModuleAddr,
			AddIndex:    nil,
			UpdateIndex: []metoken.Index{index},
		},
	)
	require.NoError(t, err)

	// confirm now the balances are unbalanced
	checkBalances(t, ctx, app, k, index.Denom, false, true)

	// move ctx to match rebalance time
	futureCtx := app.NewContext(
		false, tmproto.Header{
			ChainID: ctx.ChainID(),
			Height:  ctx.BlockHeight(),
		},
	).WithBlockTime(time.Now().Add(24 * time.Hour))

	err = app.MetokenKeeperB.Keeper(&futureCtx).RebalanceReserves()
	require.NoError(t, err)

	// confirm the balances are good now
	checkBalances(t, ctx, app, k, index.Denom, true, false)
}

func checkBalances(
	t *testing.T,
	ctx sdk.Context,
	app *app.UmeeApp,
	k keeper.Keeper,
	meTokenDenom string,
	shouldBeBalanced bool,
	supplyToLeverageAllowed bool,
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
		assetSettings, i := index.AcceptedAsset(balance.Denom)
		require.True(t, i >= 0)

		desiredReserves := assetSettings.ReservePortion.MulInt(balance.AvailableSupply()).TruncateInt()
		require.Equal(t, shouldBeBalanced, desiredReserves.Equal(balance.Reserved))

		// check the supply in x/leverage and confirm is the same
		allSupplied, err := app.LeverageKeeper.GetAllSupplied(ctx, meTokenAddr)
		require.NoError(t, err)

		found, assetSupplied := allSupplied.Find(balance.Denom)
		require.Equal(t, supplyToLeverageAllowed, found)
		if supplyToLeverageAllowed {
			require.True(t, assetSupplied.Amount.Equal(balance.Leveraged))
		}

		// confirm balance in x/bank and x/module state is the same
		found, bBalance := bankBalance.Find(balance.Denom)
		require.True(t, found)
		require.True(t, bBalance.Amount.Equal(balance.Interest.Add(balance.Reserved).Add(balance.Fees)))
	}
}
