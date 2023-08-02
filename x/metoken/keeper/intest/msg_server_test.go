package intest

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/util/coin"
	"github.com/umee-network/umee/v5/x/metoken"
	"github.com/umee-network/umee/v5/x/metoken/keeper"
	"github.com/umee-network/umee/v5/x/metoken/mocks"
	otypes "github.com/umee-network/umee/v5/x/oracle/types"
)

type testCase struct {
	name   string
	addr   sdk.AccAddress
	asset  sdk.Coin
	denom  string
	errMsg string
}

func TestMsgServer_Swap(t *testing.T) {
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

	// create and fund a user with 100 USDT, 1000 USDC and 2000000 IST
	user := s.newAccount(
		t,
		coin.New(mocks.USDTBaseDenom, 100_000000),
		coin.New(mocks.USDCBaseDenom, 1000_000000),
		coin.New(mocks.ISTBaseDenom, 2_000_000_000000),
	)

	tcs := []testCase{
		{
			"invalid user address",
			sdk.AccAddress{},
			sdk.Coin{},
			"",
			"empty address string is not allowed",
		},
		{
			"invalid invalid asset",
			user,
			sdk.Coin{
				Denom:  "???",
				Amount: sdkmath.ZeroInt(),
			},
			"",
			"invalid denom",
		},
		{
			"invalid meToken denom",
			user,
			coin.New(mocks.USDTBaseDenom, 100_000000),
			"???",
			"invalid denom",
		},
		{
			"valid - swap 73 usdt",
			user,
			coin.New(mocks.USDTBaseDenom, 73_000000),
			mocks.MeUSDDenom,
			"",
		},
		{
			"valid - swap 750 usdc",
			user,
			coin.New(mocks.USDCBaseDenom, 750_000000),
			mocks.MeUSDDenom,
			"",
		},
		{
			"valid - swap 1876 ist",
			user,
			coin.New(mocks.ISTBaseDenom, 1876_000000),
			mocks.MeUSDDenom,
			"",
		},
		{
			"invalid - index not found",
			user,
			coin.New(mocks.ISTBaseDenom, 100_000000),
			"me/EUR",
			"index me/EUR not found",
		},
		{
			"invalid - asset not present en balances",
			user,
			coin.New(mocks.WBTCBaseDenom, 100_000000),
			mocks.MeUSDDenom,
			"is not accepted in the index",
		},
		{
			"max supply reached",
			user,
			coin.New(mocks.ISTBaseDenom, 1_900_000_000000),
			mocks.MeUSDDenom,
			"reaching the max supply",
		},
	}

	for _, tc := range tcs {
		msg := &metoken.MsgSwap{
			User:         tc.addr.String(),
			Asset:        tc.asset,
			MetokenDenom: tc.denom,
		}
		if len(tc.errMsg) > 0 {
			_, err := msgServer.Swap(ctx, msg)
			assert.ErrorContains(t, err, tc.errMsg, tc.name)
		} else {
			meTokenDenom := tc.denom
			k := app.MetokenKeeperB.Keeper(&ctx)

			// initial state
			iUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			iUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			iMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(t, err)

			prices, err := k.Prices(index)
			require.NoError(t, err)

			// verify the outputs of swap function
			resp, err := msgServer.Swap(ctx, msg)
			require.NoError(t, err, tc.name)

			// final state
			fUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(t, err)

			verifySwap(
				t,
				tc,
				index,
				iUserBalance,
				fUserBalance,
				iUTokenSupply,
				fUTokenSupply,
				iMeTokenBalance,
				fMeTokenBalance,
				prices,
				*resp,
			)
		}
	}
}

func TestMsgServer_Swap_NonStableAssets_DiffExponents(t *testing.T) {
	index := mocks.NonStableIndex(mocks.MeNonStableDenom)

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

	// create and fund a user with 10000 USDT, 1.431 WBTC, 2.876 ETH
	user := s.newAccount(
		t,
		coin.New(mocks.USDTBaseDenom, 10000_000000),
		coin.New(mocks.WBTCBaseDenom, 1_43100000),
		coin.New(mocks.ETHBaseDenom, 2_876000000000000000),
	)

	tcs := []testCase{
		{
			"valid - first swap 1547 USDT",
			user,
			coin.New(mocks.USDTBaseDenom, 1547_000000),
			mocks.MeNonStableDenom,
			"",
		},
		{
			"valid - swap 0.57686452 WBTC",
			user,
			coin.New(mocks.WBTCBaseDenom, 57686452),
			mocks.MeNonStableDenom,
			"",
		},
		{
			"valid - swap 1.435125562353141231 ETH",
			user,
			coin.New(mocks.ETHBaseDenom, 1_435125562353141231),
			mocks.MeNonStableDenom,
			"",
		},
	}

	for _, tc := range tcs {
		msg := &metoken.MsgSwap{
			User:         tc.addr.String(),
			Asset:        tc.asset,
			MetokenDenom: tc.denom,
		}
		if len(tc.errMsg) > 0 {
			_, err := msgServer.Swap(ctx, msg)
			assert.ErrorContains(t, err, tc.errMsg)
		} else {
			meTokenDenom := tc.denom
			k := app.MetokenKeeperB.Keeper(&ctx)

			// initial state
			iUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			iUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			iMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(t, err)

			prices, err := k.Prices(index)
			require.NoError(t, err)

			// verify the outputs of swap function
			resp, err := msgServer.Swap(ctx, msg)
			require.NoError(t, err, tc.name)

			// final state
			fUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(t, err)

			verifySwap(
				t,
				tc,
				index,
				iUserBalance,
				fUserBalance,
				iUTokenSupply,
				fUTokenSupply,
				iMeTokenBalance,
				fMeTokenBalance,
				prices,
				*resp,
			)
		}
	}
}

func TestMsgServer_Swap_AfterAddingAssetToIndex(t *testing.T) {
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

	// create and fund a user with 1000 USDT, 5000 USDC, 20000 IST and 7.674 ETH
	user := s.newAccount(
		t,
		coin.New(mocks.USDTBaseDenom, 10000_000000),
		coin.New(mocks.USDCBaseDenom, 5000_000000),
		coin.New(mocks.ISTBaseDenom, 20000_000000),
		coin.New(mocks.ETHBaseDenom, 7_674000000000000000),
	)

	tcs := []testCase{
		{
			"valid - first swap 7546 USDT",
			user,
			coin.New(mocks.USDTBaseDenom, 7546_000000),
			mocks.MeUSDDenom,
			"",
		},
		{
			"valid - swap 1432.77 USDC",
			user,
			coin.New(mocks.USDCBaseDenom, 1432_770000),
			mocks.MeUSDDenom,
			"",
		},
		{
			"valid - swap 15600.82 IST",
			user,
			coin.New(mocks.ISTBaseDenom, 15600_820000),
			mocks.MeUSDDenom,
			"",
		},
	}

	for _, tc := range tcs {
		msg := &metoken.MsgSwap{
			User:         tc.addr.String(),
			Asset:        tc.asset,
			MetokenDenom: tc.denom,
		}
		if len(tc.errMsg) > 0 {
			_, err := msgServer.Swap(ctx, msg)
			assert.ErrorContains(t, err, tc.errMsg)
		} else {
			meTokenDenom := tc.denom
			k := app.MetokenKeeperB.Keeper(&ctx)

			// initial state
			iUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			iUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			iMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(t, err)

			prices, err := k.Prices(index)
			require.NoError(t, err)

			// verify the outputs of swap function
			resp, err := msgServer.Swap(ctx, msg)
			require.NoError(t, err, tc.name)

			// final state
			fUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(t, err)

			verifySwap(
				t,
				tc,
				index,
				iUserBalance,
				fUserBalance,
				iUTokenSupply,
				fUTokenSupply,
				iMeTokenBalance,
				fMeTokenBalance,
				prices,
				*resp,
			)
		}
	}

	// after initial swaps ETH is added to the index and target_allocation is recalculated
	for i := 0; i < len(index.AcceptedAssets); i++ {
		index.AcceptedAssets[i].TargetAllocation = sdk.MustNewDecFromStr("0.25")
	}
	index.AcceptedAssets = append(
		index.AcceptedAssets,
		metoken.NewAcceptedAsset(mocks.ETHBaseDenom, sdk.MustNewDecFromStr("0.2"), sdk.MustNewDecFromStr("0.25")),
	)

	_, err = msgServer.GovUpdateRegistry(
		ctx, &metoken.MsgGovUpdateRegistry{
			Authority:   app.GovKeeper.GetGovernanceAccount(s.ctx).GetAddress().String(),
			AddIndex:    nil,
			UpdateIndex: []metoken.Index{index},
		},
	)
	require.NoError(t, err)

	tcs = []testCase{
		{
			"valid - swap 1243.56 USDT",
			user,
			coin.New(mocks.USDTBaseDenom, 1243_560000),
			mocks.MeUSDDenom,
			"",
		},
		{
			"valid - swap 6.312 ETH",
			user,
			coin.New(mocks.ETHBaseDenom, 6_312000000000000000),
			mocks.MeUSDDenom,
			"",
		},
		{
			"valid - swap 0.312 ETH",
			user,
			coin.New(mocks.ETHBaseDenom, 312000000000000000),
			mocks.MeUSDDenom,
			"",
		},
		{
			"valid - swap 2700 USDC",
			user,
			coin.New(mocks.USDCBaseDenom, 2700_000000),
			mocks.MeUSDDenom,
			"",
		},
		{
			"valid - swap 1000.07 IST",
			user,
			coin.New(mocks.ISTBaseDenom, 1000_070000),
			mocks.MeUSDDenom,
			"",
		},
	}

	for _, tc := range tcs {
		msg := &metoken.MsgSwap{
			User:         tc.addr.String(),
			Asset:        tc.asset,
			MetokenDenom: tc.denom,
		}
		if len(tc.errMsg) > 0 {
			_, err := msgServer.Swap(ctx, msg)
			assert.ErrorContains(t, err, tc.errMsg)
		} else {
			meTokenDenom := tc.denom
			k := app.MetokenKeeperB.Keeper(&ctx)

			// initial state
			iUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			iUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			iMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(t, err)

			prices, err := k.Prices(index)
			require.NoError(t, err)

			// verify the outputs of swap function
			resp, err := msgServer.Swap(ctx, msg)
			require.NoError(t, err, tc.name)

			// final state
			fUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(t, err)

			verifySwap(
				t,
				tc,
				index,
				iUserBalance,
				fUserBalance,
				iUTokenSupply,
				fUTokenSupply,
				iMeTokenBalance,
				fMeTokenBalance,
				prices,
				*resp,
			)
		}
	}
}

func TestMsgServer_Swap_Depegging(t *testing.T) {
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

	// create and fund a user with 10000 USDT, 10000 USDC, 10000 IST
	user := s.newAccount(
		t,
		coin.New(mocks.USDTBaseDenom, 10000_000000),
		coin.New(mocks.USDCBaseDenom, 10000_000000),
		coin.New(mocks.ISTBaseDenom, 10000_000000),
	)

	tcs := []testCase{
		{
			"valid - first swap 343.055 IST",
			user,
			coin.New(mocks.ISTBaseDenom, 343_055000),
			mocks.MeUSDDenom,
			"",
		},
		{
			"valid - swap 12.77 IST",
			user,
			coin.New(mocks.ISTBaseDenom, 12_770000),
			mocks.MeUSDDenom,
			"",
		},
		{
			"valid - swap 48.33 USDC",
			user,
			coin.New(mocks.USDCBaseDenom, 48_330000),
			mocks.MeUSDDenom,
			"",
		},
	}

	for _, tc := range tcs {
		msg := &metoken.MsgSwap{
			User:         tc.addr.String(),
			Asset:        tc.asset,
			MetokenDenom: tc.denom,
		}
		if len(tc.errMsg) > 0 {
			_, err := msgServer.Swap(ctx, msg)
			assert.ErrorContains(t, err, tc.errMsg)
		} else {
			meTokenDenom := tc.denom
			k := app.MetokenKeeperB.Keeper(&ctx)

			// initial state
			iUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			iUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			iMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(t, err)

			prices, err := k.Prices(index)
			require.NoError(t, err)

			// verify the outputs of swap function
			resp, err := msgServer.Swap(ctx, msg)
			require.NoError(t, err, tc.name)

			// final state
			fUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(t, err)

			verifySwap(
				t,
				tc,
				index,
				iUserBalance,
				fUserBalance,
				iUTokenSupply,
				fUTokenSupply,
				iMeTokenBalance,
				fMeTokenBalance,
				prices,
				*resp,
			)
		}
	}

	// after initial swaps IST price is dropped to 0.64 USD
	oracleMock := mocks.NewMockOracleKeeper()
	oracleMock.AllMedianPricesFunc.SetDefaultHook(
		func(ctx sdk.Context) otypes.Prices {
			prices := otypes.Prices{}
			median := otypes.Price{
				ExchangeRateTuple: otypes.NewExchangeRateTuple(
					mocks.USDTSymbolDenom,
					mocks.USDTPrice,
				),
				BlockNum: uint64(1),
			}
			prices = append(prices, median)

			median = otypes.Price{
				ExchangeRateTuple: otypes.NewExchangeRateTuple(
					mocks.USDCSymbolDenom,
					mocks.USDCPrice,
				),
				BlockNum: uint64(1),
			}
			prices = append(prices, median)

			median = otypes.Price{
				ExchangeRateTuple: otypes.NewExchangeRateTuple(
					mocks.ISTSymbolDenom,
					sdk.MustNewDecFromStr("0.64"),
				),
				BlockNum: uint64(1),
			}
			prices = append(prices, median)

			return prices
		},
	)

	kb := keeper.NewKeeperBuilder(
		app.AppCodec(),
		app.GetKey(metoken.ModuleName),
		app.BankKeeper,
		app.LeverageKeeper,
		oracleMock,
	)
	app.MetokenKeeperB = kb
	msgServer = keeper.NewMsgServerImpl(app.MetokenKeeperB)

	tcs = []testCase{
		{
			"valid - swap 1243.56 USDT",
			user,
			coin.New(mocks.USDTBaseDenom, 1243_560000),
			mocks.MeUSDDenom,
			"",
		},
		{
			"valid - swap 2000 IST",
			user,
			coin.New(mocks.ISTBaseDenom, 2000_000000),
			mocks.MeUSDDenom,
			"",
		},
		{
			"valid - swap 1000.07 IST",
			user,
			coin.New(mocks.ISTBaseDenom, 1000_070000),
			mocks.MeUSDDenom,
			"",
		},
		{
			"valid - swap 753.011 USDC",
			user,
			coin.New(mocks.USDCBaseDenom, 753_011000),
			mocks.MeUSDDenom,
			"",
		},
	}

	for _, tc := range tcs {
		msg := &metoken.MsgSwap{
			User:         tc.addr.String(),
			Asset:        tc.asset,
			MetokenDenom: tc.denom,
		}
		if len(tc.errMsg) > 0 {
			_, err := msgServer.Swap(ctx, msg)
			assert.ErrorContains(t, err, tc.errMsg)
		} else {
			meTokenDenom := tc.denom
			k := app.MetokenKeeperB.Keeper(&ctx)

			// initial state
			iUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			iUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			iMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(t, err)

			prices, err := k.Prices(index)
			require.NoError(t, err)

			// verify the outputs of swap function
			resp, err := msgServer.Swap(ctx, msg)
			require.NoError(t, err, tc.name)

			// final state
			fUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(t, err)

			verifySwap(
				t,
				tc,
				index,
				iUserBalance,
				fUserBalance,
				iUTokenSupply,
				fUTokenSupply,
				iMeTokenBalance,
				fMeTokenBalance,
				prices,
				*resp,
			)
		}
	}

	// after some swaps with depegged price, all is back to normal
	oracleMock.AllMedianPricesFunc.SetDefaultHook(mocks.ValidPricesFunc())

	kb = keeper.NewKeeperBuilder(
		app.AppCodec(),
		app.GetKey(metoken.ModuleName),
		app.BankKeeper,
		app.LeverageKeeper,
		oracleMock,
	)
	app.MetokenKeeperB = kb
	msgServer = keeper.NewMsgServerImpl(app.MetokenKeeperB)

	tcs = []testCase{
		{
			"valid - swap 312.04 USDT",
			user,
			coin.New(mocks.USDTBaseDenom, 312_040000),
			mocks.MeUSDDenom,
			"",
		},
		{
			"valid - swap 145 IST",
			user,
			coin.New(mocks.ISTBaseDenom, 145_000000),
			mocks.MeUSDDenom,
			"",
		},
		{
			"valid - swap 255.478 USDT",
			user,
			coin.New(mocks.USDTBaseDenom, 255_478000),
			mocks.MeUSDDenom,
			"",
		},
		{
			"valid - swap 753.011 USDC",
			user,
			coin.New(mocks.USDCBaseDenom, 753_011000),
			mocks.MeUSDDenom,
			"",
		},
	}

	for _, tc := range tcs {
		msg := &metoken.MsgSwap{
			User:         tc.addr.String(),
			Asset:        tc.asset,
			MetokenDenom: tc.denom,
		}
		if len(tc.errMsg) > 0 {
			_, err := msgServer.Swap(ctx, msg)
			assert.ErrorContains(t, err, tc.errMsg)
		} else {
			meTokenDenom := tc.denom
			k := app.MetokenKeeperB.Keeper(&ctx)

			// initial state
			iUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			iUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			iMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(t, err)

			prices, err := k.Prices(index)
			require.NoError(t, err)

			// verify the outputs of swap function
			resp, err := msgServer.Swap(ctx, msg)
			require.NoError(t, err, tc.name)

			// final state
			fUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(t, err)

			verifySwap(
				t,
				tc,
				index,
				iUserBalance,
				fUserBalance,
				iUTokenSupply,
				fUTokenSupply,
				iMeTokenBalance,
				fMeTokenBalance,
				prices,
				*resp,
			)
		}
	}
}

func verifySwap(
	t *testing.T, tc testCase, index metoken.Index,
	iUserBalance, fUserBalance, iUTokenSupply, fUTokenSupply sdk.Coins,
	iMeTokenBalance, fMeTokenBalance metoken.IndexBalances,
	prices metoken.IndexPrices, resp metoken.MsgSwapResponse,
) {
	denom, meTokenDenom := tc.asset.Denom, tc.denom

	// initial state
	assetPrice, err := prices.Price(denom)
	assert.NilError(t, err)
	i, iAssetBalance := iMeTokenBalance.AssetBalance(denom)
	assert.Check(t, i >= 0)
	assetSupply := iAssetBalance.Leveraged.Add(iAssetBalance.Reserved)
	meTokenPrice, err := prices.Price(meTokenDenom)
	assert.NilError(t, err)

	assetExponentFactorVsUSD, err := metoken.ExponentFactor(assetPrice.Exponent, 0)
	assert.NilError(t, err)
	decAssetSupply := assetExponentFactorVsUSD.MulInt(assetSupply)
	assetValue := decAssetSupply.Mul(assetPrice.Price)

	meTokenExponentFactor, err := metoken.ExponentFactor(meTokenPrice.Exponent, 0)
	assert.NilError(t, err)
	decMeTokenSupply := meTokenExponentFactor.MulInt(iMeTokenBalance.MetokenSupply.Amount)
	meTokenValue := decMeTokenSupply.Mul(meTokenPrice.Price)

	// current_allocation = asset_value / total_value
	// swap_delta_allocation = (current_allocation - target_allocation) / target_allocation
	currentAllocation, swapDeltaAllocation := sdk.ZeroDec(), sdk.ZeroDec()
	i, aa := index.AcceptedAsset(denom)
	assert.Check(t, i >= 0)
	targetAllocation := aa.TargetAllocation

	if assetSupply.IsZero() {
		swapDeltaAllocation = currentAllocation.Sub(targetAllocation)
	} else {
		currentAllocation = assetValue.Quo(meTokenValue)
		swapDeltaAllocation = currentAllocation.Sub(targetAllocation).Quo(targetAllocation)
	}

	// fee = delta_allocation * balanced_fee + balanced_fee
	fee := swapDeltaAllocation.Mul(index.Fee.BalancedFee).Add(index.Fee.BalancedFee)
	if fee.LT(index.Fee.MinFee) {
		fee = index.Fee.MinFee
	}
	if fee.GT(index.Fee.MaxFee) {
		fee = index.Fee.MaxFee
	}

	// if current_allocation = 0, fee = min_fee
	if !currentAllocation.IsPositive() {
		fee = index.Fee.MinFee
	}

	// expected_fee = fee * amount
	expectedFee := sdk.NewCoin(denom, fee.MulInt(tc.asset.Amount).TruncateInt())

	// amount_to_swap = swap_amount - fee
	amountToSwap := tc.asset.Amount.Sub(expectedFee.Amount)

	// swap_exchange_rate = asset_price / metoken_price
	exchangeRate := assetPrice.Price.Quo(meTokenPrice.Price)

	assetExponentFactorVsMeToken, err := prices.ExponentFactor(denom, meTokenDenom)
	assert.NilError(t, err)

	// expected_metokens = amount_to_swap * exchange_rate * exponent_factor
	expectedMeTokens := sdk.NewCoin(
		meTokenDenom,
		exchangeRate.MulInt(amountToSwap).Mul(assetExponentFactorVsMeToken).TruncateInt(),
	)

	// calculating reserved and leveraged
	expectedReserved := sdk.NewCoin(
		denom,
		aa.ReservePortion.MulInt(amountToSwap).TruncateInt(),
	)
	expectedLeveraged := sdk.NewCoin(denom, amountToSwap.Sub(expectedReserved.Amount))

	// verify the outputs of swap function
	require.Equal(t, expectedFee, resp.Fee, tc.name)
	require.Equal(t, expectedMeTokens, resp.Returned, tc.name)

	// verify token balance decreased and meToken balance increased by the expected amounts
	require.Equal(
		t,
		iUserBalance.Sub(tc.asset).Add(expectedMeTokens),
		fUserBalance,
		tc.name,
		"token balance",
	)
	// verify uToken assetSupply increased by the expected amount
	require.Equal(
		t,
		iUTokenSupply.Add(sdk.NewCoin("u/"+expectedLeveraged.Denom, expectedLeveraged.Amount)),
		fUTokenSupply,
		tc.name,
		"uToken assetSupply",
	)
	// reserved + leveraged + fee must be = to total amount supplied by the user for the swap
	require.Equal(t, expectedReserved.Add(expectedLeveraged).Add(expectedFee), tc.asset)

	// meToken assetSupply is increased by the expected amount
	require.Equal(
		t,
		iMeTokenBalance.MetokenSupply.Add(expectedMeTokens),
		fMeTokenBalance.MetokenSupply,
		tc.name,
		"meToken assetSupply",
	)

	i, fAssetBalance := fMeTokenBalance.AssetBalance(denom)
	assert.Check(t, i >= 0)
	require.Equal(
		t,
		iAssetBalance.Reserved.Add(expectedReserved.Amount),
		fAssetBalance.Reserved,
		"reserved",
	)
	require.Equal(
		t,
		iAssetBalance.Leveraged.Add(expectedLeveraged.Amount),
		fAssetBalance.Leveraged,
		"leveraged",
	)
	require.Equal(
		t,
		iAssetBalance.Fees.Add(expectedFee.Amount),
		fAssetBalance.Fees,
		"fee",
	)
}

func TestMsgServer_Redeem(t *testing.T) {
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

	tcs := []testCase{
		{
			"invalid user address",
			sdk.AccAddress{},
			sdk.Coin{},
			"",
			"empty address string is not allowed",
		},
		{
			"invalid meToken",
			user,
			sdk.Coin{
				Denom:  "???",
				Amount: sdkmath.ZeroInt(),
			},
			"",
			"invalid denom",
		},
		{
			"invalid asset denom",
			user,
			coin.New(mocks.MeUSDDenom, 100_000000),
			"???",
			"invalid denom",
		},
		{
			"valid - redemption 155.9876 meUSD for IST",
			user,
			coin.New(mocks.MeUSDDenom, 155_987600),
			mocks.ISTBaseDenom,
			"",
		},
		{
			"valid - redemption 750.56 meUSD for USDC - but not enough USDC",
			user,
			coin.New(mocks.MeUSDDenom, 750_560000),
			mocks.USDCBaseDenom,
			"not enough",
		},
		{
			"valid - redemption 187 meUSD for ist",
			user,
			coin.New(mocks.MeUSDDenom, 187_000000),
			mocks.ISTBaseDenom,
			"",
		},
		{
			"valid - redemption 468.702000 meUSD for USDT",
			user,
			coin.New(mocks.MeUSDDenom, 468_702000),
			mocks.USDTBaseDenom,
			"",
		},
		{
			"invalid - index doesn't exist",
			user,
			coin.New("me/EUR", 100_000000),
			"me/EUR",
			"index me/EUR not found",
		},
		{
			"valid - redemption 1000.13 meUSD for IST - but not enough IST",
			user,
			coin.New(mocks.MeUSDDenom, 1000_130000),
			mocks.ISTBaseDenom,
			"not enough",
		},
	}

	for _, tc := range tcs {
		msg := &metoken.MsgRedeem{
			User:       tc.addr.String(),
			Metoken:    tc.asset,
			AssetDenom: tc.denom,
		}
		if len(tc.errMsg) > 0 {
			_, err := msgServer.Redeem(ctx, msg)
			assert.ErrorContains(t, err, tc.errMsg, tc.name)
		} else {
			meTokenDenom := tc.asset.Denom
			k := app.MetokenKeeperB.Keeper(&ctx)

			// initial state
			iUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			iUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			iMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(t, err)

			prices, err := k.Prices(index)
			require.NoError(t, err)

			// verify the outputs of redeem function
			resp, err := msgServer.Redeem(ctx, msg)
			require.NoError(t, err, tc.name)

			// final state
			fUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(t, err)

			verifyRedeem(
				t,
				tc,
				index,
				iUserBalance,
				fUserBalance,
				iUTokenSupply,
				fUTokenSupply,
				iMeTokenBalance,
				fMeTokenBalance,
				prices,
				*resp,
			)
		}
	}
}

func TestMsgServer_Redeem_NonStableAssets_DiffExponents(t *testing.T) {
	index := mocks.NonStableIndex(mocks.MeNonStableDenom)

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

	// create and fund a user with 10000 USDT, 1.431 WBTC, 2.876 ETH
	user := s.newAccount(
		t,
		coin.New(mocks.USDTBaseDenom, 10000_000000),
		coin.New(mocks.WBTCBaseDenom, 21_43100000),
		coin.New(mocks.ETHBaseDenom, 2_876000000000000000),
	)

	// swap 1547 USDT, 20.57686452 WBTC and 0.7855 ETH to have an initial meNonStable balance
	swaps := []*metoken.MsgSwap{
		{
			User:         user.String(),
			Asset:        sdk.NewCoin(mocks.USDTBaseDenom, sdkmath.NewInt(1547_000000)),
			MetokenDenom: mocks.MeNonStableDenom,
		},
		{
			User:         user.String(),
			Asset:        sdk.NewCoin(mocks.WBTCBaseDenom, sdkmath.NewInt(20_57686452)),
			MetokenDenom: mocks.MeNonStableDenom,
		},
		{
			User:         user.String(),
			Asset:        sdk.NewCoin(mocks.ETHBaseDenom, sdkmath.NewInt(785500000000000000)),
			MetokenDenom: mocks.MeNonStableDenom,
		},
	}

	for _, swap := range swaps {
		resp, err := msgServer.Swap(ctx, swap)
		require.NoError(t, err)
		require.NotNil(t, resp)
	}

	tcs := []testCase{
		{
			"valid - redeem 2.182736 meNonStable for WBTC",
			user,
			coin.New(mocks.MeNonStableDenom, 2_18273600),
			mocks.WBTCBaseDenom,
			"",
		},
		{
			"valid - redeem 0.05879611 meNonStable ETH",
			user,
			coin.New(mocks.MeNonStableDenom, 5879611),
			mocks.ETHBaseDenom,
			"",
		},
		{
			"valid - redeem 0.1 meNonStable USDT",
			user,
			coin.New(mocks.MeNonStableDenom, 10000000),
			mocks.USDTBaseDenom,
			"",
		},
		{
			"valid - redeem 12 meNonStable for WBTC",
			user,
			coin.New(mocks.MeNonStableDenom, 12_00000000),
			mocks.WBTCBaseDenom,
			"",
		},
		{
			"valid - redeem 2.182736 meNonStable for USDT - not enough USDT",
			user,
			coin.New(mocks.MeNonStableDenom, 2_18273600),
			mocks.USDTBaseDenom,
			"not enough",
		},
	}

	for _, tc := range tcs {
		msg := &metoken.MsgRedeem{
			User:       tc.addr.String(),
			Metoken:    tc.asset,
			AssetDenom: tc.denom,
		}
		if len(tc.errMsg) > 0 {
			_, err := msgServer.Redeem(ctx, msg)
			assert.ErrorContains(t, err, tc.errMsg)
		} else {
			meTokenDenom := tc.asset.Denom
			k := app.MetokenKeeperB.Keeper(&ctx)

			// initial state
			iUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			iUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			iMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(t, err)

			prices, err := k.Prices(index)
			require.NoError(t, err)

			// verify the outputs of redeem function
			resp, err := msgServer.Redeem(ctx, msg)
			require.NoError(t, err, tc.name)

			// final state
			fUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(t, err)

			verifyRedeem(
				t,
				tc,
				index,
				iUserBalance,
				fUserBalance,
				iUTokenSupply,
				fUTokenSupply,
				iMeTokenBalance,
				fMeTokenBalance,
				prices,
				*resp,
			)
		}
	}
}

func TestMsgServer_Redeem_Depegging(t *testing.T) {
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

	// create and fund a user with 10000 USDT, 10000 USDC, 10000 IST
	user := s.newAccount(
		t,
		coin.New(mocks.USDTBaseDenom, 10000_000000),
		coin.New(mocks.USDCBaseDenom, 10000_000000),
		coin.New(mocks.ISTBaseDenom, 10000_000000),
	)

	// swap 5000 USDT, 3500 USDC an initial meUSD balance
	swaps := []*metoken.MsgSwap{
		{
			User:         user.String(),
			Asset:        sdk.NewCoin(mocks.USDTBaseDenom, sdkmath.NewInt(5000_000000)),
			MetokenDenom: mocks.MeUSDDenom,
		},
		{
			User:         user.String(),
			Asset:        sdk.NewCoin(mocks.USDCBaseDenom, sdkmath.NewInt(3500_000000)),
			MetokenDenom: mocks.MeUSDDenom,
		},
	}

	for _, swap := range swaps {
		_, err := msgServer.Swap(ctx, swap)
		require.NoError(t, err)
	}

	// after initial swaps USDT price is dropped to 0.73 USD
	oracleMock := mocks.NewMockOracleKeeper()
	oracleMock.AllMedianPricesFunc.SetDefaultHook(
		func(ctx sdk.Context) otypes.Prices {
			prices := otypes.Prices{}
			median := otypes.Price{
				ExchangeRateTuple: otypes.NewExchangeRateTuple(
					mocks.USDTSymbolDenom,
					sdk.MustNewDecFromStr("0.73"),
				),
				BlockNum: uint64(1),
			}
			prices = append(prices, median)

			median = otypes.Price{
				ExchangeRateTuple: otypes.NewExchangeRateTuple(
					mocks.USDCSymbolDenom,
					mocks.USDCPrice,
				),
				BlockNum: uint64(1),
			}
			prices = append(prices, median)

			median = otypes.Price{
				ExchangeRateTuple: otypes.NewExchangeRateTuple(
					mocks.ISTSymbolDenom,
					mocks.ISTPrice,
				),
				BlockNum: uint64(1),
			}
			prices = append(prices, median)

			return prices
		},
	)

	kb := keeper.NewKeeperBuilder(
		app.AppCodec(),
		app.GetKey(metoken.ModuleName),
		app.BankKeeper,
		app.LeverageKeeper,
		oracleMock,
	)
	app.MetokenKeeperB = kb
	msgServer = keeper.NewMsgServerImpl(app.MetokenKeeperB)

	tcs := []testCase{
		{
			"valid - redeem 100 meUSD for IST",
			user,
			coin.New(mocks.MeUSDDenom, 100_000000),
			mocks.ISTBaseDenom,
			"not enough",
		},
		{
			"valid - redeem 1000 meUSD for USDC",
			user,
			coin.New(mocks.MeUSDDenom, 1000_000000),
			mocks.USDCBaseDenom,
			"",
		},
		{
			"valid - redeem 2500 meUSD for USDT",
			user,
			coin.New(mocks.MeUSDDenom, 2500_000000),
			mocks.USDTBaseDenom,
			"",
		},
		{
			"valid - redeem 2200 meUSD for USDC",
			user,
			coin.New(mocks.MeUSDDenom, 2200_000000),
			mocks.USDCBaseDenom,
			"",
		},
	}

	for _, tc := range tcs {
		msg := &metoken.MsgRedeem{
			User:       tc.addr.String(),
			Metoken:    tc.asset,
			AssetDenom: tc.denom,
		}
		if len(tc.errMsg) > 0 {
			_, err := msgServer.Redeem(ctx, msg)
			assert.ErrorContains(t, err, tc.errMsg)
		} else {
			meTokenDenom := tc.asset.Denom
			k := app.MetokenKeeperB.Keeper(&ctx)

			// initial state
			iUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			iUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			iMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(t, err)

			prices, err := k.Prices(index)
			require.NoError(t, err)

			// verify the outputs of redeem function
			resp, err := msgServer.Redeem(ctx, msg)
			require.NoError(t, err, tc.name)

			// final state
			fUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(t, err)

			verifyRedeem(
				t,
				tc,
				index,
				iUserBalance,
				fUserBalance,
				iUTokenSupply,
				fUTokenSupply,
				iMeTokenBalance,
				fMeTokenBalance,
				prices,
				*resp,
			)
		}
	}

	// after some redemptions with depegged price, all is back to normal
	oracleMock.AllMedianPricesFunc.SetDefaultHook(mocks.ValidPricesFunc())

	kb = keeper.NewKeeperBuilder(
		app.AppCodec(),
		app.GetKey(metoken.ModuleName),
		app.BankKeeper,
		app.LeverageKeeper,
		oracleMock,
	)
	app.MetokenKeeperB = kb
	msgServer = keeper.NewMsgServerImpl(app.MetokenKeeperB)

	tcs = []testCase{
		{
			"valid - swap 312.04 USDT",
			user,
			coin.New(mocks.USDTBaseDenom, 312_040000),
			mocks.MeUSDDenom,
			"",
		},
		{
			"valid - swap 145 IST",
			user,
			coin.New(mocks.ISTBaseDenom, 145_000000),
			mocks.MeUSDDenom,
			"",
		},
		{
			"valid - swap 255.478 USDT",
			user,
			coin.New(mocks.USDTBaseDenom, 255_478000),
			mocks.MeUSDDenom,
			"",
		},
		{
			"valid - swap 753.011 USDC",
			user,
			coin.New(mocks.USDCBaseDenom, 753_011000),
			mocks.MeUSDDenom,
			"",
		},
	}

	for _, tc := range tcs {
		msg := &metoken.MsgSwap{
			User:         tc.addr.String(),
			Asset:        tc.asset,
			MetokenDenom: tc.denom,
		}
		if len(tc.errMsg) > 0 {
			_, err := msgServer.Swap(ctx, msg)
			assert.ErrorContains(t, err, tc.errMsg)
		} else {
			meTokenDenom := tc.denom
			k := app.MetokenKeeperB.Keeper(&ctx)

			// initial state
			iUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			iUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			iMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(t, err)

			prices, err := k.Prices(index)
			require.NoError(t, err)

			// verify the outputs of swap function
			resp, err := msgServer.Swap(ctx, msg)
			require.NoError(t, err, tc.name)

			// final state
			fUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(t, err)

			verifySwap(
				t,
				tc,
				index,
				iUserBalance,
				fUserBalance,
				iUTokenSupply,
				fUTokenSupply,
				iMeTokenBalance,
				fMeTokenBalance,
				prices,
				*resp,
			)
		}
	}

	// redeem all the available meUSD in the balance to confirm we have sufficient liquidity to do it
	tcs = []testCase{
		{
			"valid - redeem 1584.411571 meUSD for USDT",
			user,
			coin.New(mocks.MeUSDDenom, 1584_411571),
			mocks.USDTBaseDenom,
			"",
		},
		{
			"valid - redeem 934.139005 meUSD for USDC",
			user,
			coin.New(mocks.MeUSDDenom, 934_139005),
			mocks.USDCBaseDenom,
			"",
		},
		{
			"valid - redeem 118.125618 meUSD for IST",
			user,
			coin.New(mocks.MeUSDDenom, 118_125618),
			mocks.ISTBaseDenom,
			"",
		},
	}

	for _, tc := range tcs {
		msg := &metoken.MsgRedeem{
			User:       tc.addr.String(),
			Metoken:    tc.asset,
			AssetDenom: tc.denom,
		}
		if len(tc.errMsg) > 0 {
			_, err := msgServer.Redeem(ctx, msg)
			assert.ErrorContains(t, err, tc.errMsg)
		} else {
			meTokenDenom := tc.asset.Denom
			k := app.MetokenKeeperB.Keeper(&ctx)

			// initial state
			iUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			iUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			iMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(t, err)

			prices, err := k.Prices(index)
			require.NoError(t, err)

			// verify the outputs of redeem function
			resp, err := msgServer.Redeem(ctx, msg)
			require.NoError(t, err, tc.name)

			// final state
			fUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(t, err)

			verifyRedeem(
				t,
				tc,
				index,
				iUserBalance,
				fUserBalance,
				iUTokenSupply,
				fUTokenSupply,
				iMeTokenBalance,
				fMeTokenBalance,
				prices,
				*resp,
			)
		}
	}
}

func verifyRedeem(
	t *testing.T, tc testCase, index metoken.Index,
	iUserBalance, fUserBalance, iUTokenSupply, fUTokenSupply sdk.Coins,
	iMeTokenBalance, fMeTokenBalance metoken.IndexBalances,
	prices metoken.IndexPrices, resp metoken.MsgRedeemResponse,
) {
	meTokenDenom, denom := tc.asset.Denom, tc.denom

	// initial state
	assetPrice, err := prices.Price(denom)
	assert.NilError(t, err)
	i, iAssetBalance := iMeTokenBalance.AssetBalance(denom)
	assert.Check(t, i >= 0)
	assetSupply := iAssetBalance.Leveraged.Add(iAssetBalance.Reserved)
	meTokenPrice, err := prices.Price(meTokenDenom)
	assert.NilError(t, err)

	assetExponentFactorVsUSD, err := metoken.ExponentFactor(assetPrice.Exponent, 0)
	assert.NilError(t, err)
	decAssetSupply := assetExponentFactorVsUSD.MulInt(assetSupply)
	assetValue := decAssetSupply.Mul(assetPrice.Price)

	meTokenExponentFactor, err := metoken.ExponentFactor(meTokenPrice.Exponent, 0)
	assert.NilError(t, err)
	decMeTokenSupply := meTokenExponentFactor.MulInt(iMeTokenBalance.MetokenSupply.Amount)
	meTokenValue := decMeTokenSupply.Mul(meTokenPrice.Price)

	// current_allocation = asset_value / total_value
	// redeem_delta_allocation = (target_allocation - current_allocation) / target_allocation
	currentAllocation, redeemDeltaAllocation := sdk.ZeroDec(), sdk.ZeroDec()
	i, aa := index.AcceptedAsset(denom)
	assert.Check(t, i >= 0)
	targetAllocation := aa.TargetAllocation
	if assetSupply.IsZero() {
		redeemDeltaAllocation = targetAllocation
	} else {
		currentAllocation = assetValue.Quo(meTokenValue)
		redeemDeltaAllocation = targetAllocation.Sub(currentAllocation).Quo(targetAllocation)
	}

	// fee = delta_allocation * balanced_fee + balanced_fee
	fee := redeemDeltaAllocation.Mul(index.Fee.BalancedFee).Add(index.Fee.BalancedFee)
	if fee.LT(index.Fee.MinFee) {
		fee = index.Fee.MinFee
	}
	if fee.GT(index.Fee.MaxFee) {
		fee = index.Fee.MaxFee
	}
	// redeem_exchange_rate = metoken_price / asset_price
	redeemExchangeRate := meTokenPrice.Price.Quo(assetPrice.Price)

	assetExponentFactorVsMeToken, err := prices.ExponentFactor(meTokenDenom, denom)
	assert.NilError(t, err)

	// amount_to_redeem = exchange_rate * metoken_amount
	amountToWithdraw := redeemExchangeRate.MulInt(tc.asset.Amount).Mul(assetExponentFactorVsMeToken).TruncateInt()

	// expected_fee = fee * amount_to_redeem
	expectedFee := sdk.NewCoin(denom, fee.MulInt(amountToWithdraw).TruncateInt())

	// amount_to_redeem = amountToWithdraw - expectedFee
	amountToRedeem := amountToWithdraw.Sub(expectedFee.Amount)

	expectedAssets := sdk.NewCoin(
		denom,
		amountToRedeem,
	)

	// calculating reserved and leveraged
	expectedFromReserves := sdk.NewCoin(
		denom,
		aa.ReservePortion.MulInt(amountToWithdraw).TruncateInt(),
	)
	expectedFromLeverage := sdk.NewCoin(denom, amountToWithdraw.Sub(expectedFromReserves.Amount))

	// verify the outputs of swap function
	require.Equal(t, expectedFee, resp.Fee, tc.name, "expectedFee")
	require.Equal(t, expectedAssets, resp.Returned, tc.name, "expectedAssets")

	// verify meToken balance decreased and asset balance increased by the expected amounts
	require.True(
		t,
		iUserBalance.Sub(tc.asset).Add(expectedAssets).IsEqual(fUserBalance),
		tc.name,
		"token balance",
	)
	// verify uToken assetSupply decreased by the expected amount
	require.True(
		t,
		iUTokenSupply.Sub(
			sdk.NewCoin(
				"u/"+expectedFromLeverage.Denom,
				expectedFromLeverage.Amount,
			),
		).IsEqual(fUTokenSupply),
		tc.name,
		"uToken assetSupply",
	)
	// from_reserves + from_leverage must be = to total amount withdrawn from the modules
	require.True(
		t,
		expectedFromReserves.Amount.Add(expectedFromLeverage.Amount).Equal(amountToWithdraw),
		tc.name,
		"total withdraw",
	)

	// meToken assetSupply is decreased by the expected amount
	require.True(
		t,
		iMeTokenBalance.MetokenSupply.Sub(tc.asset).IsEqual(fMeTokenBalance.MetokenSupply),
		tc.name,
		"meToken assetSupply",
	)

	i, fAssetBalance := fMeTokenBalance.AssetBalance(denom)
	assert.Check(t, i >= 0)
	require.True(
		t,
		iAssetBalance.Reserved.Sub(expectedFromReserves.Amount).Equal(fAssetBalance.Reserved),
		tc.name,
		"reserved",
	)
	require.True(
		t,
		iAssetBalance.Leveraged.Sub(expectedFromLeverage.Amount).Equal(fAssetBalance.Leveraged),
		tc.name,
		"leveraged",
	)
	require.True(t, iAssetBalance.Fees.Add(expectedFee.Amount).Equal(fAssetBalance.Fees), tc.name, "fees")
}

func TestMsgServer_GovSetParams(t *testing.T) {
	s := initTestSuite(t, nil, nil)
	msgServer, ctx, app := s.msgServer, s.ctx, s.app

	testCases := []struct {
		name   string
		req    *metoken.MsgGovSetParams
		errMsg string
	}{
		{
			"invalid authority",
			metoken.NewMsgGovSetParams("invalid_authority", metoken.Params{}),
			"invalid_authority",
		},
		{
			"valid",
			metoken.NewMsgGovSetParams(
				app.GovKeeper.GetGovernanceAccount(s.ctx).GetAddress().String(),
				metoken.DefaultParams(),
			),
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(
			tc.name, func(t *testing.T) {
				_, err := msgServer.GovSetParams(ctx, tc.req)
				if len(tc.errMsg) > 0 {
					assert.ErrorContains(t, err, tc.errMsg)
				} else {
					assert.NilError(t, err)
				}
			},
		)
	}
}

func TestMsgServer_GovUpdateRegistry(t *testing.T) {
	s := initTestSuite(t, nil, nil)
	msgServer, ctx, app := s.msgServer, s.ctx, s.app
	govAddr := app.GovKeeper.GetGovernanceAccount(s.ctx).GetAddress().String()

	existingIndex := mocks.StableIndex("me/Existing")
	_, err := msgServer.GovUpdateRegistry(
		ctx,
		metoken.NewMsgGovUpdateRegistry(govAddr, []metoken.Index{existingIndex}, nil),
	)
	require.NoError(t, err)

	user := s.newAccount(
		t,
		coin.New(mocks.USDTBaseDenom, 10000_000000),
	)

	// swap 5000 USDT for existing index
	swap := &metoken.MsgSwap{
		User:         user.String(),
		Asset:        sdk.NewCoin(mocks.USDTBaseDenom, sdkmath.NewInt(5000_000000)),
		MetokenDenom: existingIndex.Denom,
	}

	_, err = msgServer.Swap(ctx, swap)
	require.NoError(t, err)
	existingIndex.Exponent = 10

	indexWithNotRegisteredToken := metoken.NewIndex(
		"me/NotRegistered",
		sdk.NewInt(1_000_000_000_000),
		6,
		mocks.ValidFee(),
		[]metoken.AcceptedAsset{
			metoken.NewAcceptedAsset("notRegisteredDenom", sdk.MustNewDecFromStr("0.2"), sdk.MustNewDecFromStr("1.0")),
		},
	)

	deletedAssetIndex := mocks.StableIndex(mocks.MeUSDDenom)
	aa := []metoken.AcceptedAsset{
		metoken.NewAcceptedAsset(mocks.USDTBaseDenom, sdk.MustNewDecFromStr("0.2"), sdk.MustNewDecFromStr("0.5")),
		metoken.NewAcceptedAsset(mocks.USDCBaseDenom, sdk.MustNewDecFromStr("0.2"), sdk.MustNewDecFromStr("0.5")),
	}
	deletedAssetIndex.AcceptedAssets = aa

	testCases := []struct {
		name   string
		req    *metoken.MsgGovUpdateRegistry
		errMsg string
	}{
		{
			"invalid authority",
			metoken.NewMsgGovUpdateRegistry("invalid_authority", nil, nil),
			"invalid_authority",
		},
		{
			"invalid - empty add and update indexes",
			metoken.NewMsgGovUpdateRegistry(govAddr, nil, nil),
			"empty add and update indexes",
		},
		{
			"invalid - duplicated add indexes",
			metoken.NewMsgGovUpdateRegistry(
				govAddr,
				[]metoken.Index{mocks.StableIndex(mocks.MeUSDDenom), mocks.StableIndex(mocks.MeUSDDenom)},
				nil,
			),
			"duplicate addIndex metoken denom",
		},
		{
			"invalid - duplicated update indexes",
			metoken.NewMsgGovUpdateRegistry(
				govAddr,
				[]metoken.Index{mocks.StableIndex(mocks.MeUSDDenom)},
				[]metoken.Index{mocks.StableIndex(mocks.MeUSDDenom)},
			),
			"duplicate updateIndex metoken denom",
		},
		{
			"invalid - add index",
			metoken.NewMsgGovUpdateRegistry(
				govAddr,
				[]metoken.Index{mocks.StableIndex(mocks.USDTBaseDenom)},
				nil,
			),
			"should have the following format: me/<TokenName>",
		},
		{
			"invalid - existing add index",
			metoken.NewMsgGovUpdateRegistry(govAddr, []metoken.Index{existingIndex}, nil),
			"already exists",
		},
		{
			"invalid - index with not registered token",
			metoken.NewMsgGovUpdateRegistry(govAddr, []metoken.Index{indexWithNotRegisteredToken}, nil),
			"not a registered Token",
		},
		{
			"valid - add",
			metoken.NewMsgGovUpdateRegistry(govAddr, []metoken.Index{mocks.StableIndex(mocks.MeUSDDenom)}, nil),
			"",
		},
		{
			"invalid - update index",
			metoken.NewMsgGovUpdateRegistry(
				govAddr,
				nil,
				[]metoken.Index{mocks.StableIndex(mocks.USDTBaseDenom)},
			),
			"should have the following format: me/<TokenName>",
		},
		{
			"invalid - non-existing update index",
			metoken.NewMsgGovUpdateRegistry(govAddr, nil, []metoken.Index{mocks.StableIndex("me/NonExisting")}),
			"not found",
		},
		{
			"invalid - update index exponent with balance",
			metoken.NewMsgGovUpdateRegistry(govAddr, nil, []metoken.Index{existingIndex}),
			"exponent cannot be changed when supply is greater than zero",
		},
		{
			"invalid - update index deleting an asset",
			metoken.NewMsgGovUpdateRegistry(govAddr, nil, []metoken.Index{deletedAssetIndex}),
			"cannot be deleted from an index",
		},
	}

	for _, tc := range testCases {
		t.Run(
			tc.name, func(t *testing.T) {
				_, err := msgServer.GovUpdateRegistry(ctx, tc.req)
				if len(tc.errMsg) > 0 {
					assert.ErrorContains(t, err, tc.errMsg)
				} else {
					assert.NilError(t, err)
					for _, addIndex := range tc.req.AddIndex {
						index, err := app.MetokenKeeperB.Keeper(&ctx).RegisteredIndex(
							addIndex.Denom,
						)
						require.NoError(t, err)
						assert.DeepEqual(t, addIndex, index)

						balances, found := app.MetokenKeeperB.Keeper(&ctx).IndexBalances(
							addIndex.Denom,
						)
						assert.Check(t, found)
						for _, aa := range addIndex.AcceptedAssets {
							i, balance := balances.AssetBalance(aa.Denom)
							assert.Check(t, i >= 0)
							assert.Check(t, balance.Fees.Equal(sdkmath.ZeroInt()))
							assert.Check(t, balance.Interest.Equal(sdkmath.ZeroInt()))
							assert.Check(t, balance.Leveraged.Equal(sdkmath.ZeroInt()))
							assert.Check(t, balance.Reserved.Equal(sdkmath.ZeroInt()))
						}
					}

					for _, updateIndex := range tc.req.AddIndex {
						index, err := app.MetokenKeeperB.Keeper(&ctx).RegisteredIndex(updateIndex.Denom)
						require.NoError(t, err)
						assert.DeepEqual(t, updateIndex, index)

						balances, found := app.MetokenKeeperB.Keeper(&ctx).IndexBalances(updateIndex.Denom)
						assert.Check(t, found)
						for _, aa := range updateIndex.AcceptedAssets {
							i, balance := balances.AssetBalance(aa.Denom)
							assert.Check(t, i >= 0)
							assert.Check(t, balance.Fees.Equal(sdkmath.ZeroInt()))
							assert.Check(t, balance.Interest.Equal(sdkmath.ZeroInt()))
							assert.Check(t, balance.Leveraged.Equal(sdkmath.ZeroInt()))
							assert.Check(t, balance.Reserved.Equal(sdkmath.ZeroInt()))
						}
					}
				}
			},
		)
	}
}
