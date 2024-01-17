package intest

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/util/checkers"
	"github.com/umee-network/umee/v6/util/coin"
	"github.com/umee-network/umee/v6/x/metoken"
	"github.com/umee-network/umee/v6/x/metoken/keeper"
	"github.com/umee-network/umee/v6/x/metoken/mocks"
	otypes "github.com/umee-network/umee/v6/x/oracle/types"
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
			Authority:   checkers.GovModuleAddr,
			AddIndex:    []metoken.Index{index},
			UpdateIndex: nil,
		},
	)
	require := require.New(t)
	require.NoError(err)

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
			"zero amount",
			user,
			sdk.Coin{
				Denom:  mocks.USDTBaseDenom,
				Amount: sdkmath.ZeroInt(),
			},
			"",
			"zero amount",
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
			require.NoError(err)

			prices, err := k.Prices(index)
			require.NoError(err)

			// verify the outputs of swap function
			resp, err := msgServer.Swap(ctx, msg)
			require.NoError(err, tc.name)

			// final state
			fUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(err)

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
			Authority:   checkers.GovModuleAddr,
			AddIndex:    []metoken.Index{index},
			UpdateIndex: nil,
		},
	)
	require := require.New(t)
	require.NoError(err)

	user := s.newAccount(
		t,
		coin.New(mocks.CMSTBaseDenom, 10000_000000),
		coin.New(mocks.WBTCBaseDenom, 1_43100000),
		coin.New(mocks.ETHBaseDenom, 2_876000000000000000),
	)

	tcs := []testCase{
		{
			"valid - first swap 1547 CMST",
			user,
			coin.New(mocks.CMSTBaseDenom, 1547_000000),
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
			require.NoError(err)

			prices, err := k.Prices(index)
			require.NoError(err)

			// verify the outputs of swap function
			if tc.name == "valid - swap 1.435125562353141231 ETH" {
				fmt.Printf("\n")
			}
			resp, err := msgServer.Swap(ctx, msg)
			require.NoError(err, tc.name)

			// final state
			fUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(err)

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
			Authority:   checkers.GovModuleAddr,
			AddIndex:    []metoken.Index{index},
			UpdateIndex: nil,
		},
	)
	require := require.New(t)
	require.NoError(err)

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
			require.NoError(err)

			prices, err := k.Prices(index)
			require.NoError(err)

			// verify the outputs of swap function
			resp, err := msgServer.Swap(ctx, msg)
			require.NoError(err, tc.name)

			// final state
			fUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(err)

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
		index.AcceptedAssets[i].TargetAllocation = sdkmath.LegacyMustNewDecFromStr("0.25")
	}
	index.AcceptedAssets = append(
		index.AcceptedAssets,
		metoken.NewAcceptedAsset(mocks.ETHBaseDenom, sdkmath.LegacyMustNewDecFromStr("0.2"), sdkmath.LegacyMustNewDecFromStr("0.25")),
	)

	_, err = msgServer.GovUpdateRegistry(
		ctx, &metoken.MsgGovUpdateRegistry{
			Authority:   checkers.GovModuleAddr,
			AddIndex:    nil,
			UpdateIndex: []metoken.Index{index},
		},
	)
	require.NoError(err)

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
			require.NoError(err)

			prices, err := k.Prices(index)
			require.NoError(err)

			// verify the outputs of swap function
			resp, err := msgServer.Swap(ctx, msg)
			require.NoError(err, tc.name)

			// final state
			fUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(err)

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
			Authority:   checkers.GovModuleAddr,
			AddIndex:    []metoken.Index{index},
			UpdateIndex: nil,
		},
	)
	require := require.New(t)
	require.NoError(err)

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
			require.NoError(err)

			prices, err := k.Prices(index)
			require.NoError(err)

			// verify the outputs of swap function
			resp, err := msgServer.Swap(ctx, msg)
			require.NoError(err, tc.name)

			// final state
			fUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(err)

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

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	initialPrices := otypes.Prices{
		otypes.Price{
			ExchangeRateTuple: otypes.NewExchangeRateTuple(
				mocks.USDTSymbolDenom,
				mocks.USDTPrice,
			),
			BlockNum: uint64(1),
		}, otypes.Price{
			ExchangeRateTuple: otypes.NewExchangeRateTuple(
				mocks.USDCSymbolDenom,
				mocks.USDCPrice,
			),
			BlockNum: uint64(1),
		}, otypes.Price{
			ExchangeRateTuple: otypes.NewExchangeRateTuple(
				mocks.ISTSymbolDenom,
				sdkmath.LegacyMustNewDecFromStr("0.64"),
			),
			BlockNum: uint64(1),
		},
	}

	oracleMock := mocks.NewMockOracleKeeper(ctrl)
	oracleMock.
		EXPECT().
		AllMedianPrices(gomock.Any()).
		Return(initialPrices).
		AnyTimes()

	kb := keeper.NewKeeperBuilder(
		app.AppCodec(),
		app.GetKey(metoken.ModuleName),
		app.BankKeeper,
		app.LeverageKeeper,
		oracleMock,
		app.UGovKeeperB.EmergencyGroup,
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
			require.NoError(err)

			prices, err := k.Prices(index)
			require.NoError(err)

			// verify the outputs of swap function
			resp, err := msgServer.Swap(ctx, msg)
			require.NoError(err, tc.name)

			// final state
			fUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(err)

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
	oracleMock.
		EXPECT().
		AllMedianPrices(gomock.Any()).
		Return(mocks.ValidPrices()).
		AnyTimes()

	kb = keeper.NewKeeperBuilder(
		app.AppCodec(),
		app.GetKey(metoken.ModuleName),
		app.BankKeeper,
		app.LeverageKeeper,
		oracleMock,
		app.UGovKeeperB.EmergencyGroup,
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
			require.NoError(err)

			prices, err := k.Prices(index)
			require.NoError(err)

			// verify the outputs of swap function
			resp, err := msgServer.Swap(ctx, msg)
			require.NoError(err, tc.name)

			// final state
			fUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(err)

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

func TestMsgServer_Swap_EdgeCase(t *testing.T) {
	index := mocks.NonStableIndex(mocks.MeNonStableDenom)

	s := initTestSuite(t, nil, nil)
	msgServer, ctx, app := s.msgServer, s.ctx, s.app

	_, err := msgServer.GovUpdateRegistry(
		ctx, &metoken.MsgGovUpdateRegistry{
			Authority:   checkers.GovModuleAddr,
			AddIndex:    []metoken.Index{index},
			UpdateIndex: nil,
		},
	)
	require := require.New(t)
	require.NoError(err)

	// create and fund a user with 1 ETH
	user := s.newAccount(
		t,
		coin.New(mocks.ETHBaseDenom, 1_000000000000000000),
	)

	k := app.MetokenKeeperB.Keeper(&ctx)
	iMeTokenBalance, err := k.IndexBalances(mocks.MeNonStableDenom)
	require.NoError(err)

	// 0.00000001 ETH is less than the minimum amount of meToken, so the swap shouldn't change the balances
	_, err = msgServer.Swap(
		ctx,
		&metoken.MsgSwap{
			User: user.String(), Asset: coin.New(mocks.ETHBaseDenom, 10000000000),
			MetokenDenom: mocks.MeNonStableDenom,
		},
	)
	require.ErrorContains(err, "insufficient")

	// the result should be the same as the initial balance
	fMeTokenBalance, err := k.IndexBalances(mocks.MeNonStableDenom)
	require.NoError(err)
	require.Equal(iMeTokenBalance, fMeTokenBalance)
}

func verifySwap(
	t *testing.T, tc testCase, index metoken.Index,
	iUserBalance, fUserBalance, iUTokenSupply, fUTokenSupply sdk.Coins,
	iMeTokenBalance, fMeTokenBalance metoken.IndexBalances,
	prices metoken.IndexPrices, resp metoken.MsgSwapResponse,
) {
	denom, meTokenDenom := tc.asset.Denom, tc.denom

	// initial state
	assetPrice, err := prices.PriceByBaseDenom(denom)
	assert.NilError(t, err)
	iAssetBalance, i := iMeTokenBalance.AssetBalance(denom)
	assert.Check(t, i >= 0)
	assetSupply := iAssetBalance.Leveraged.Add(iAssetBalance.Reserved)

	assetExponentFactorVsUSD, err := metoken.ExponentFactor(assetPrice.Exponent, 0)
	assert.NilError(t, err)
	decAssetSupply := assetExponentFactorVsUSD.MulInt(assetSupply)
	assetValue := decAssetSupply.Mul(assetPrice.Price)

	meTokenExponentFactor, err := metoken.ExponentFactor(prices.Exponent, 0)
	assert.NilError(t, err)
	decMeTokenSupply := meTokenExponentFactor.MulInt(iMeTokenBalance.MetokenSupply.Amount)
	meTokenValue := decMeTokenSupply.Mul(prices.Price)

	// current_allocation = asset_value / total_value
	// swap_delta_allocation = (current_allocation - target_allocation) / target_allocation
	currentAllocation, swapDeltaAllocation := sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec()
	aa, i := index.AcceptedAsset(denom)
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
	exchangeRate := assetPrice.Price.Quo(prices.Price)

	assetExponentFactorVsMeToken, err := metoken.ExponentFactor(assetPrice.Exponent, prices.Exponent)
	assert.NilError(t, err)
	rate := exchangeRate.Mul(assetExponentFactorVsMeToken)

	// expected_metokens = amount_to_swap * exchange_rate * exponent_factor
	expectedMeTokens := sdk.NewCoin(
		meTokenDenom,
		rate.MulInt(amountToSwap).TruncateInt(),
	)

	// calculating reserved and leveraged
	expectedReserved := sdk.NewCoin(
		denom,
		aa.ReservePortion.MulInt(amountToSwap).TruncateInt(),
	)
	expectedLeveraged := sdk.NewCoin(denom, amountToSwap.Sub(expectedReserved.Amount))

	// verify the outputs of swap function
	require := require.New(t)
	require.Equal(expectedFee, resp.Fee, tc.name)
	require.Equal(expectedMeTokens, resp.Returned, tc.name)

	// verify token balance decreased and meToken balance increased by the expected amounts
	require.Equal(
		iUserBalance.Sub(tc.asset).Add(expectedMeTokens),
		fUserBalance,
		tc.name,
		"token balance",
	)
	// verify uToken assetSupply increased by the expected amount
	require.Equal(
		iUTokenSupply.Add(sdk.NewCoin("u/"+expectedLeveraged.Denom, expectedLeveraged.Amount)),
		fUTokenSupply,
		tc.name,
		"uToken assetSupply",
	)
	// reserved + leveraged + fee must be = to total amount supplied by the user for the swap
	require.Equal(expectedReserved.Add(expectedLeveraged).Add(expectedFee), tc.asset)

	// meToken assetSupply is increased by the expected amount
	require.Equal(
		iMeTokenBalance.MetokenSupply.Add(expectedMeTokens),
		fMeTokenBalance.MetokenSupply,
		tc.name,
		"meToken assetSupply",
	)

	fAssetBalance, i := fMeTokenBalance.AssetBalance(denom)
	assert.Check(t, i >= 0)
	require.Equal(
		iAssetBalance.Reserved.Add(expectedReserved.Amount),
		fAssetBalance.Reserved,
		"reserved",
	)
	require.Equal(
		iAssetBalance.Leveraged.Add(expectedLeveraged.Amount),
		fAssetBalance.Leveraged,
		"leveraged",
	)
	require.Equal(
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
			Authority:   checkers.GovModuleAddr,
			AddIndex:    []metoken.Index{index},
			UpdateIndex: nil,
		},
	)
	require := require.New(t)
	require.NoError(err)

	user := s.newAccount(
		t,
		coin.New(mocks.USDTBaseDenom, 1000_000000),
		coin.New(mocks.USDCBaseDenom, 1000_000000),
		coin.New(mocks.ISTBaseDenom, 1000_000000),
	)

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
		require.NoError(err)
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
			"zero amount",
			user,
			sdk.Coin{
				Denom:  mocks.MeUSDDenom,
				Amount: sdkmath.ZeroInt(),
			},
			"",
			"zero amount",
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
			require.NoError(err)

			prices, err := k.Prices(index)
			require.NoError(err)

			// verify the outputs of redeem function
			resp, err := msgServer.Redeem(ctx, msg)
			require.NoError(err, tc.name)

			// final state
			fUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(err)

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
			Authority:   checkers.GovModuleAddr,
			AddIndex:    []metoken.Index{index},
			UpdateIndex: nil,
		},
	)
	require := require.New(t)
	require.NoError(err)

	user := s.newAccount(
		t,
		coin.New(mocks.CMSTBaseDenom, 10000_000000),
		coin.New(mocks.WBTCBaseDenom, 21_43100000),
		coin.New(mocks.ETHBaseDenom, 2_876000000000000000),
	)

	// swap 1547 USDT, 20.57686452 WBTC and 0.7855 ETH to have an initial meNonStable balance
	swaps := []*metoken.MsgSwap{
		{
			User:         user.String(),
			Asset:        sdk.NewCoin(mocks.CMSTBaseDenom, sdkmath.NewInt(1547_000000)),
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
		require.NoError(err)
		require.NotNil(resp)
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
			"valid - redeem 0.1 meNonStable CMST",
			user,
			coin.New(mocks.MeNonStableDenom, 10000000),
			mocks.CMSTBaseDenom,
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
			"valid - redeem 2.182736 meNonStable for CMST - not enough CMST",
			user,
			coin.New(mocks.MeNonStableDenom, 2_18273600),
			mocks.CMSTBaseDenom,
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
			require.NoError(err)

			prices, err := k.Prices(index)
			require.NoError(err)

			// verify the outputs of redeem function
			resp, err := msgServer.Redeem(ctx, msg)
			require.NoError(err, tc.name)

			// final state
			fUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(err)

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
			Authority:   checkers.GovModuleAddr,
			AddIndex:    []metoken.Index{index},
			UpdateIndex: nil,
		},
	)
	require := require.New(t)
	require.NoError(err)

	user := s.newAccount(
		t,
		coin.New(mocks.USDTBaseDenom, 10000_000000),
		coin.New(mocks.USDCBaseDenom, 10000_000000),
		coin.New(mocks.ISTBaseDenom, 10000_000000),
	)

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
		require.NoError(err)
	}

	// after initial swaps USDT price is dropped to 0.73 USD

	initialPrices := otypes.Prices{
		otypes.Price{
			ExchangeRateTuple: otypes.NewExchangeRateTuple(
				mocks.USDTSymbolDenom,
				sdkmath.LegacyMustNewDecFromStr("0.73"),
			),
			BlockNum: uint64(1),
		}, otypes.Price{
			ExchangeRateTuple: otypes.NewExchangeRateTuple(
				mocks.USDCSymbolDenom,
				mocks.USDCPrice,
			),
			BlockNum: uint64(1),
		}, otypes.Price{
			ExchangeRateTuple: otypes.NewExchangeRateTuple(
				mocks.ISTSymbolDenom,
				mocks.ISTPrice,
			),
			BlockNum: uint64(1),
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oracleMock := mocks.NewMockOracleKeeper(ctrl)
	oracleMock.
		EXPECT().
		AllMedianPrices(gomock.Any()).
		Return(initialPrices).
		AnyTimes()

	kb := keeper.NewKeeperBuilder(
		app.AppCodec(),
		app.GetKey(metoken.ModuleName),
		app.BankKeeper,
		app.LeverageKeeper,
		oracleMock,
		app.UGovKeeperB.EmergencyGroup,
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
			require.NoError(err)

			prices, err := k.Prices(index)
			require.NoError(err)

			// verify the outputs of redeem function
			resp, err := msgServer.Redeem(ctx, msg)
			require.NoError(err, tc.name)

			// final state
			fUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(err)

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
	oracleMock.
		EXPECT().
		AllMedianPrices(gomock.Any()).
		Return(mocks.ValidPrices()).
		AnyTimes()

	kb = keeper.NewKeeperBuilder(
		app.AppCodec(),
		app.GetKey(metoken.ModuleName),
		app.BankKeeper,
		app.LeverageKeeper,
		oracleMock,
		app.UGovKeeperB.EmergencyGroup,
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
			require.NoError(err)

			prices, err := k.Prices(index)
			require.NoError(err)

			// verify the outputs of swap function
			resp, err := msgServer.Swap(ctx, msg)
			require.NoError(err, tc.name)

			// final state
			fUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(err)

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
			require.NoError(err)

			prices, err := k.Prices(index)
			require.NoError(err)

			// verify the outputs of redeem function
			resp, err := msgServer.Redeem(ctx, msg)
			require.NoError(err, tc.name)

			// final state
			fUserBalance := app.BankKeeper.GetAllBalances(ctx, tc.addr)
			fUTokenSupply := app.LeverageKeeper.GetAllUTokenSupply(ctx)
			fMeTokenBalance, err := k.IndexBalances(meTokenDenom)
			require.NoError(err)

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

func TestMsgServer_Redeem_EdgeCase(t *testing.T) {
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
	require := require.New(t)
	require.NoError(err)

	user := s.newAccount(
		t,
		coin.New(mocks.USDCBaseDenom, 100_000000),
	)

	_, err = msgServer.Swap(
		ctx,
		&metoken.MsgSwap{
			User: user.String(), Asset: coin.New(mocks.USDCBaseDenom, 1_000000),
			MetokenDenom: mocks.MeUSDDenom,
		},
	)
	require.NoError(err)

	// try to redeem 0.000001 me/USD (the minimum possible amount) for an asset with a higher price than me/USD
	k := app.MetokenKeeperB.Keeper(&ctx)
	iMeTokenBalance, err := k.IndexBalances(mocks.MeUSDDenom)
	require.NoError(err)

	_, err = msgServer.Redeem(
		ctx,
		&metoken.MsgRedeem{User: user.String(), Metoken: coin.New(mocks.MeUSDDenom, 1), AssetDenom: mocks.ISTBaseDenom},
	)
	require.ErrorContains(err, "insufficient")

	// the result should be the same as the initial balance
	// if me/USD amount redeemed is less than the minimum amount possible of an asset, no meTokens should be burned
	fMeTokenBalance, err := k.IndexBalances(mocks.MeUSDDenom)
	require.Equal(iMeTokenBalance, fMeTokenBalance)
}

func verifyRedeem(
	t *testing.T, tc testCase, index metoken.Index,
	iUserBalance, fUserBalance, iUTokenSupply, fUTokenSupply sdk.Coins,
	iMeTokenBalance, fMeTokenBalance metoken.IndexBalances,
	prices metoken.IndexPrices, resp metoken.MsgRedeemResponse,
) {
	// initial state
	assetPrice, err := prices.PriceByBaseDenom(tc.denom)
	assert.NilError(t, err)
	iAssetBalance, i := iMeTokenBalance.AssetBalance(tc.denom)
	assert.Check(t, i >= 0)
	assetSupply := iAssetBalance.Leveraged.Add(iAssetBalance.Reserved)

	assetExponentFactorVsUSD, err := metoken.ExponentFactor(assetPrice.Exponent, 0)
	assert.NilError(t, err)
	decAssetSupply := assetExponentFactorVsUSD.MulInt(assetSupply)
	assetValue := decAssetSupply.Mul(assetPrice.Price)

	meTokenExponentFactor, err := metoken.ExponentFactor(prices.Exponent, 0)
	assert.NilError(t, err)
	decMeTokenSupply := meTokenExponentFactor.MulInt(iMeTokenBalance.MetokenSupply.Amount)
	meTokenValue := decMeTokenSupply.Mul(prices.Price)

	// current_allocation = asset_value / total_value
	// redeem_delta_allocation = (target_allocation - current_allocation) / target_allocation
	currentAllocation, redeemDeltaAllocation := sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec()
	aa, i := index.AcceptedAsset(tc.denom)
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
	redeemExchangeRate := prices.Price.Quo(assetPrice.Price)

	assetExponentFactorVsMeToken, err := metoken.ExponentFactor(prices.Exponent, assetPrice.Exponent)
	assert.NilError(t, err)

	// amount_to_redeem = exchange_rate * metoken_amount
	amountToWithdraw := redeemExchangeRate.MulInt(tc.asset.Amount).Mul(assetExponentFactorVsMeToken).TruncateInt()

	// expected_fee = fee * amount_to_redeem
	expectedFee := sdk.NewCoin(tc.denom, fee.MulInt(amountToWithdraw).TruncateInt())

	// amount_to_redeem = amountToWithdraw - expectedFee
	amountToRedeem := amountToWithdraw.Sub(expectedFee.Amount)

	expectedAssets := sdk.NewCoin(
		tc.denom,
		amountToRedeem,
	)

	// calculating reserved and leveraged
	expectedFromReserves := sdk.NewCoin(
		tc.denom,
		aa.ReservePortion.MulInt(amountToWithdraw).TruncateInt(),
	)
	expectedFromLeverage := sdk.NewCoin(tc.denom, amountToWithdraw.Sub(expectedFromReserves.Amount))

	require := require.New(t)
	// verify the outputs of swap function
	require.Equal(expectedFee, resp.Fee, tc.name, "expectedFee")
	require.Equal(expectedAssets, resp.Returned, tc.name, "expectedAssets")

	// verify meToken balance decreased and asset balance increased by the expected amounts
	require.True(
		iUserBalance.Sub(tc.asset).Add(expectedAssets).Equal(fUserBalance),
		tc.name,
		"token balance",
	)
	// verify uToken assetSupply decreased by the expected amount
	require.True(
		iUTokenSupply.Sub(
			sdk.NewCoin(
				"u/"+expectedFromLeverage.Denom,
				expectedFromLeverage.Amount,
			),
		).Equal(fUTokenSupply),
		tc.name,
		"uToken assetSupply",
	)
	// from_reserves + from_leverage must be = to total amount withdrawn from the modules
	require.True(
		expectedFromReserves.Amount.Add(expectedFromLeverage.Amount).Equal(amountToWithdraw),
		tc.name,
		"total withdraw",
	)

	// meToken assetSupply is decreased by the expected amount
	require.True(
		iMeTokenBalance.MetokenSupply.Sub(tc.asset).IsEqual(fMeTokenBalance.MetokenSupply),
		tc.name,
		"meToken assetSupply",
	)

	fAssetBalance, i := fMeTokenBalance.AssetBalance(tc.denom)
	assert.Check(t, i >= 0)
	require.True(
		iAssetBalance.Reserved.Sub(expectedFromReserves.Amount).Equal(fAssetBalance.Reserved),
		tc.name,
		"reserved",
	)
	require.True(
		iAssetBalance.Leveraged.Sub(expectedFromLeverage.Amount).Equal(fAssetBalance.Leveraged),
		tc.name,
		"leveraged",
	)
	require.True(iAssetBalance.Fees.Add(expectedFee.Amount).Equal(fAssetBalance.Fees), tc.name, "fees")
}

func TestMsgServer_GovSetParams(t *testing.T) {
	s := initTestSuite(t, nil, nil)
	msgServer, ctx := s.msgServer, s.ctx

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
				checkers.GovModuleAddr,
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

	existingIndex := mocks.StableIndex("me/Existing")
	_, err := msgServer.GovUpdateRegistry(
		ctx,
		metoken.NewMsgGovUpdateRegistry(checkers.GovModuleAddr, []metoken.Index{existingIndex}, nil),
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
		sdkmath.NewInt(1_000_000_000_000),
		6,
		mocks.ValidFee(),
		[]metoken.AcceptedAsset{
			metoken.NewAcceptedAsset("notRegisteredDenom", sdkmath.LegacyMustNewDecFromStr("0.2"), sdkmath.LegacyMustNewDecFromStr("1.0")),
		},
	)

	deletedAssetIndex := mocks.NonStableIndex(mocks.MeNonStableDenom)
	aa := []metoken.AcceptedAsset{
		metoken.NewAcceptedAsset(mocks.WBTCSymbolDenom, sdkmath.LegacyMustNewDecFromStr("0.2"), sdkmath.LegacyMustNewDecFromStr("0.5")),
		metoken.NewAcceptedAsset(mocks.ETHSymbolDenom, sdkmath.LegacyMustNewDecFromStr("0.2"), sdkmath.LegacyMustNewDecFromStr("0.5")),
	}
	deletedAssetIndex.AcceptedAssets = aa

	testCases := []struct {
		name   string
		req    *metoken.MsgGovUpdateRegistry
		errMsg string
	}{
		{
			"invalid authority",
			metoken.NewMsgGovUpdateRegistry(
				"umee156hsyuvssklaekm57qy0pcehlfhzpclclaadwq",
				nil,
				[]metoken.Index{existingIndex},
			),
			"unauthorized",
		},
		{
			"invalid - empty add and update indexes",
			metoken.NewMsgGovUpdateRegistry(checkers.GovModuleAddr, nil, nil),
			"empty add and update indexes",
		},
		{
			"invalid - duplicated add indexes",
			metoken.NewMsgGovUpdateRegistry(
				checkers.GovModuleAddr,
				[]metoken.Index{mocks.StableIndex(mocks.MeUSDDenom), mocks.StableIndex(mocks.MeUSDDenom)},
				nil,
			),
			"duplicate addIndex metoken denom",
		},
		{
			"invalid - duplicated update indexes",
			metoken.NewMsgGovUpdateRegistry(
				checkers.GovModuleAddr,
				[]metoken.Index{mocks.StableIndex(mocks.MeUSDDenom)},
				[]metoken.Index{mocks.StableIndex(mocks.MeUSDDenom)},
			),
			"duplicate updateIndex metoken denom",
		},
		{
			"invalid - add index",
			metoken.NewMsgGovUpdateRegistry(
				checkers.GovModuleAddr,
				[]metoken.Index{mocks.StableIndex(mocks.USDTBaseDenom)},
				nil,
			),
			"should have the following format: me/<TokenName>",
		},
		{
			"invalid - existing add index",
			metoken.NewMsgGovUpdateRegistry(checkers.GovModuleAddr, []metoken.Index{existingIndex}, nil),
			"already exists",
		},
		{
			"invalid - index with not registered token",
			metoken.NewMsgGovUpdateRegistry(checkers.GovModuleAddr, []metoken.Index{indexWithNotRegisteredToken}, nil),
			"not a registered Token",
		},
		{
			"valid - add",
			metoken.NewMsgGovUpdateRegistry(
				checkers.GovModuleAddr,
				[]metoken.Index{mocks.NonStableIndex(mocks.MeNonStableDenom)},
				nil,
			),
			"",
		},
		{
			"invalid - update index",
			metoken.NewMsgGovUpdateRegistry(
				checkers.GovModuleAddr,
				nil,
				[]metoken.Index{mocks.StableIndex(mocks.USDTBaseDenom)},
			),
			"should have the following format: me/<TokenName>",
		},
		{
			"invalid - non-existing update index",
			metoken.NewMsgGovUpdateRegistry(checkers.GovModuleAddr, nil, []metoken.Index{mocks.StableIndex("me/NonExisting")}),
			"not found",
		},
		{
			"invalid - update index exponent with balance",
			metoken.NewMsgGovUpdateRegistry(checkers.GovModuleAddr, nil, []metoken.Index{existingIndex}),
			"exponent cannot be changed when supply is greater than zero",
		},
		{
			"invalid - update index deleting an asset",
			metoken.NewMsgGovUpdateRegistry(checkers.GovModuleAddr, nil, []metoken.Index{deletedAssetIndex}),
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
							balance, i := balances.AssetBalance(aa.Denom)
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
							balance, i := balances.AssetBalance(aa.Denom)
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
