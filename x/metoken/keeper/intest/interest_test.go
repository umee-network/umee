package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/umee-network/umee/v5/app"
	"github.com/umee-network/umee/v5/util/coin"
	"github.com/umee-network/umee/v5/x/metoken"
	"github.com/umee-network/umee/v5/x/metoken/mocks"
)

func TestInterestClaiming(t *testing.T) {
	index := mocks.StableIndex(mocks.MeUSDDenom)

	s := initKeeperTestSuite(t, nil, nil)
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

	// create a borrower with 10000 USDT
	borrower := s.newAccount(t, coin.New(mocks.USDTBaseDenom, 10000_000000))

	// swap 1000 USDT, 1000 USDC and 1000 IST to have an initial meUSD balance
	swaps := []*metoken.MsgSwap{
		{
			User:         user.String(),
			Asset:        sdk.NewCoin(mocks.USDTBaseDenom, sdkmath.NewInt(1000_000000)),
			MetokenDenom: mocks.MeUSDDenom,
		},
		{
			User:         user.String(),
			Asset:        sdk.NewCoin(mocks.USDCBaseDenom, sdkmath.NewInt(1000_000000)),
			MetokenDenom: mocks.MeUSDDenom,
		},
		{
			User:         user.String(),
			Asset:        sdk.NewCoin(mocks.ISTBaseDenom, sdkmath.NewInt(1000_000000)),
			MetokenDenom: mocks.MeUSDDenom,
		},
	}

	for _, swap := range swaps {
		_, err := msgServer.Swap(ctx, swap)
		require.NoError(t, err)
	}

	// supply liquidity from borrower and collateralize
	borrowerSupply := sdk.NewCoin(mocks.USDTBaseDenom, sdkmath.NewInt(10000_000000))
	_, err = app.LeverageKeeper.Supply(
		ctx,
		borrower,
		borrowerSupply,
	)
	require.NoError(t, err)

	uTokens, err := app.LeverageKeeper.ExchangeToken(ctx, borrowerSupply)
	require.NoError(t, err)

	err = app.LeverageKeeper.Collateralize(ctx, borrower, uTokens)
	require.NoError(t, err)

	// borrow 110 USDT, 200 USDC and 350 IST
	err = app.LeverageKeeper.Borrow(ctx, borrower, sdk.NewCoin(mocks.USDTBaseDenom, sdkmath.NewInt(110_000000)))
	require.NoError(t, err)
	err = app.LeverageKeeper.Borrow(ctx, borrower, sdk.NewCoin(mocks.USDCBaseDenom, sdkmath.NewInt(200_000000)))
	require.NoError(t, err)
	err = app.LeverageKeeper.Borrow(ctx, borrower, sdk.NewCoin(mocks.ISTBaseDenom, sdkmath.NewInt(350_000000)))
	require.NoError(t, err)

	// confirm meToken account didn't receive any interest yet and the balances in meToken state and x/bank module
	// are the same
	checkInterest(t, ctx, app, index.MetokenMaxSupply.Denom, true)

	// create ctx in the future and generate accrued interest
	err = app.LeverageKeeper.AccrueAllInterest(ctx)
	require.NoError(t, err)
	futureCtx := ctx.WithBlockTime(time.Now().Add(240 * time.Hour))
	err = app.LeverageKeeper.AccrueAllInterest(futureCtx)
	require.NoError(t, err)

	err = app.MetokenKeeperB.Keeper(&ctx).ClaimInterest()
	require.NoError(t, err)

	// confirm meToken account received accrued interest and the balances in meToken state and x/bank module
	// are the same
	checkInterest(t, ctx, app, index.MetokenMaxSupply.Denom, false)
}

func checkInterest(
	t *testing.T,
	ctx sdk.Context,
	app *app.UmeeApp,
	meTokenDenom string,
	zeroInterest bool,
) {
	k := app.MetokenKeeperB.Keeper(&ctx)
	metokenBalances, err := k.MustIndexBalance(meTokenDenom)
	require.NoError(t, err)
	meTokenAddr := authtypes.NewModuleAddress(metoken.ModuleName)
	bankBalance := app.BankKeeper.GetAllBalances(ctx, meTokenAddr)

	for _, balance := range metokenBalances.AssetBalances {
		// check interest
		require.Equal(t, zeroInterest, balance.Interest.IsZero())

		// confirm balance in x/bank and x/module state is the same
		found, bBalance := bankBalance.Find(balance.Denom)
		require.True(t, found)
		require.True(t, bBalance.Amount.Equal(balance.Interest.Add(balance.Reserved).Add(balance.Fees)))
	}
}
