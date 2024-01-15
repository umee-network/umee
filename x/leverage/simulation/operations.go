package simulation

import (
	"math/rand"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	appparams "github.com/umee-network/umee/v6/app/params"
	"github.com/umee-network/umee/v6/util/coin"
	umeesim "github.com/umee-network/umee/v6/util/sim"
	"github.com/umee-network/umee/v6/x/leverage/keeper"
	"github.com/umee-network/umee/v6/x/leverage/types"
)

// Default simulation operation weights for leverage messages
const (
	DefaultWeightMsgSupply            int = 100
	DefaultWeightMsgWithdraw          int = 85
	DefaultWeightMsgBorrow            int = 80
	DefaultWeightMsgCollateralize     int = 65
	DefaultWeightMsgDecollateralize   int = 60
	DefaultWeightMsgRepay             int = 70
	DefaultWeightMsgLiquidate         int = 75
	OperationWeightMsgSupply              = "op_weight_msg_supply"
	OperationWeightMsgWithdraw            = "op_weight_msg_withdraw"
	OperationWeightMsgBorrow              = "op_weight_msg_borrow"
	OperationWeightMsgCollateralize       = "op_weight_msg_collateralize"
	OperationWeightMsgDecollateralize     = "op_weight_msg_decollateralize"
	OperationWeightMsgRepay               = "op_weight_msg_repay"
	OperationWeightMsgLiquidate           = "op_weight_msg_liquidate"
)

// WeightedOperations returns all the operations from the leverage module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak types.AccountKeeper, bk bankkeeper.Keeper,
	lk keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgSupply          int
		weightMsgWithdraw        int
		weightMsgBorrow          int
		weightMsgCollateralize   int
		weightMsgDecollateralize int
		weightMsgRepay           int
		weightMsgLiquidate       int
	)
	appParams.GetOrGenerate(OperationWeightMsgSupply, &weightMsgSupply, nil,
		func(*rand.Rand) {
			weightMsgSupply = DefaultWeightMsgSupply
		},
	)
	appParams.GetOrGenerate(OperationWeightMsgWithdraw, &weightMsgWithdraw, nil,
		func(*rand.Rand) {
			weightMsgWithdraw = DefaultWeightMsgWithdraw
		},
	)
	appParams.GetOrGenerate(OperationWeightMsgBorrow, &weightMsgBorrow, nil,
		func(*rand.Rand) {
			weightMsgBorrow = DefaultWeightMsgBorrow
		},
	)
	appParams.GetOrGenerate(OperationWeightMsgCollateralize, &weightMsgCollateralize, nil,
		func(*rand.Rand) {
			weightMsgCollateralize = DefaultWeightMsgCollateralize
		},
	)
	appParams.GetOrGenerate(OperationWeightMsgDecollateralize, &weightMsgDecollateralize, nil,
		func(*rand.Rand) {
			weightMsgDecollateralize = DefaultWeightMsgDecollateralize
		},
	)
	appParams.GetOrGenerate(OperationWeightMsgRepay, &weightMsgRepay, nil,
		func(*rand.Rand) {
			weightMsgRepay = DefaultWeightMsgRepay
		},
	)
	appParams.GetOrGenerate(OperationWeightMsgLiquidate, &weightMsgLiquidate, nil,
		func(*rand.Rand) {
			weightMsgLiquidate = DefaultWeightMsgLiquidate
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgSupply,
			SimulateMsgSupply(ak, bk),
		),
		simulation.NewWeightedOperation(
			weightMsgCollateralize,
			SimulateMsgCollateralize(ak, bk),
		),
		simulation.NewWeightedOperation(
			weightMsgBorrow,
			SimulateMsgBorrow(ak, bk, lk),
		),
		simulation.NewWeightedOperation(
			weightMsgLiquidate,
			SimulateMsgLiquidate(ak, bk, lk),
		),
		simulation.NewWeightedOperation(
			weightMsgRepay,
			SimulateMsgRepay(ak, bk, lk),
		),
		simulation.NewWeightedOperation(
			weightMsgDecollateralize,
			SimulateMsgDecollateralize(ak, bk, lk),
		),
		simulation.NewWeightedOperation(
			weightMsgWithdraw,
			SimulateMsgWithdraw(ak, bk, lk),
		),
	}
}

// SimulateMsgSupply tests and runs a single msg supply where
// an account supplies some available assets.
func SimulateMsgSupply(ak simulation.AccountKeeper, bk bankkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		from, c, skip := randomSupplyFields(r, ctx, accs, bk)
		if skip {
			typename := sdk.MsgTypeURL(new(types.MsgSupply))
			return simtypes.NoOpMsg(types.ModuleName, typename, "skip all transfers"), nil, nil
		}

		msg := types.NewMsgSupply(from.Address, c)
		return deliver(r, app, ctx, ak, bk, from, msg, sdk.NewCoins(c))
	}
}

// SimulateMsgWithdraw tests and runs a single msg withdraw where
// an account attempts to withdraw some supplied assets.
func SimulateMsgWithdraw(ak simulation.AccountKeeper, bk bankkeeper.Keeper, lk keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		from, withdrawUToken, skip := randomWithdrawFields(r, ctx, accs, bk, lk)
		if skip {
			typename := sdk.MsgTypeURL(new(types.MsgWithdraw))
			return simtypes.NoOpMsg(types.ModuleName, typename, "skip all transfers"), nil, nil
		}

		msg := types.NewMsgWithdraw(from.Address, withdrawUToken)
		return deliver(r, app, ctx, ak, bk, from, msg, sdk.NewCoins(withdrawUToken))
	}
}

// SimulateMsgBorrow tests and runs a single msg borrow where
// an account attempts to borrow some assets.
func SimulateMsgBorrow(ak simulation.AccountKeeper, bk bankkeeper.Keeper, lk keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		from, token, skip := randomBorrowFields(r, ctx, accs, lk)
		if skip {
			typename := sdk.MsgTypeURL(new(types.MsgBorrow))
			return simtypes.NoOpMsg(types.ModuleName, typename, "skip all transfers"), nil, nil
		}

		msg := types.NewMsgBorrow(from.Address, token)
		return deliver(r, app, ctx, ak, bk, from, msg, nil)
	}
}

// SimulateMsgCollateralize tests and runs a single msg which adds
// some collateral to a user.
func SimulateMsgCollateralize(
	ak simulation.AccountKeeper,
	bk bankkeeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		from, collateral, skip := randomCollateralizeFields(r, ctx, accs, bk)
		if skip {
			typename := sdk.MsgTypeURL(new(types.MsgCollateralize))
			return simtypes.NoOpMsg(types.ModuleName, typename, "skip all transfers"), nil, nil
		}

		msg := types.NewMsgCollateralize(from.Address, collateral)
		return deliver(r, app, ctx, ak, bk, from, msg, sdk.NewCoins(collateral))
	}
}

// SimulateMsgDecollateralize tests and runs a single msg which removes
// some collateral from a user.
func SimulateMsgDecollateralize(
	ak simulation.AccountKeeper,
	bk bankkeeper.Keeper,
	lk keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		from, collateral, skip := randomDecollateralizeFields(r, ctx, accs, lk)
		if skip {
			typename := sdk.MsgTypeURL(new(types.MsgDecollateralize))
			return simtypes.NoOpMsg(types.ModuleName, typename, "skip all transfers"), nil, nil
		}

		msg := types.NewMsgDecollateralize(from.Address, collateral)
		return deliver(r, app, ctx, ak, bk, from, msg, nil)
	}
}

// SimulateMsgRepay tests and runs a single msg repay where
// an account repays some borrowed assets.
func SimulateMsgRepay(ak simulation.AccountKeeper, bk bankkeeper.Keeper, lk keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		from, repayToken, skip := randomRepayFields(r, ctx, accs, lk)
		if skip {
			typename := sdk.MsgTypeURL(new(types.MsgRepay))
			return simtypes.NoOpMsg(types.ModuleName, typename, "skip all transfers"), nil, nil
		}

		msg := types.NewMsgRepay(from.Address, repayToken)
		return deliver(r, app, ctx, ak, bk, from, msg, sdk.NewCoins(repayToken))
	}
}

// SimulateMsgLiquidate tests and runs a single msg liquidate where
// one user attempts to liquidate another user's borrow.
func SimulateMsgLiquidate(ak simulation.AccountKeeper, bk bankkeeper.Keeper, lk keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		liquidator, borrower, repaymentToken, rewardDenom, skip := randomLiquidateFields(r, ctx, accs, lk)
		if skip {
			typename := sdk.MsgTypeURL(new(types.MsgLiquidate))
			return simtypes.NoOpMsg(types.ModuleName, typename, "skip all transfers"), nil, nil
		}

		msg := types.NewMsgLiquidate(liquidator.Address, borrower.Address, repaymentToken, rewardDenom)
		return deliver(r, app, ctx, ak, bk, liquidator, msg, sdk.NewCoins(repaymentToken))
	}
}

// randomCoin get random coin from coins
func randomCoin(r *rand.Rand, coins sdk.Coins) sdk.Coin {
	if coins.Empty() {
		return sdk.Coin{}
	}
	return coins[r.Int31n(int32(coins.Len()))]
}

// getSpendableTokens returns all non-uTokens from an account's spendable coins.
func getSpendableTokens(ctx sdk.Context, addr sdk.AccAddress, bk types.BankKeeper) sdk.Coins {
	tokens := sdk.NewCoins()
	for _, c := range bk.SpendableCoins(ctx, addr) {
		if !coin.HasUTokenPrefix(c.Denom) {
			tokens = tokens.Add(c)
		}
	}

	return tokens
}

// getSpendableUTokens returns all uTokens from an account's spendable coins.
func getSpendableUTokens(ctx sdk.Context, addr sdk.AccAddress, bk types.BankKeeper) sdk.Coins {
	uTokens := sdk.NewCoins()
	for _, c := range bk.SpendableCoins(ctx, addr) {
		if coin.HasUTokenPrefix(c.Denom) {
			uTokens = uTokens.Add(c)
		}
	}

	return uTokens
}

// randomSupplyFields returns a random account and a non-uToken from its spendable
// coins. It returns skip=true if the account has zero spendable base tokens.
func randomSupplyFields(
	r *rand.Rand, ctx sdk.Context, accs []simtypes.Account, bk types.BankKeeper,
) (acc simtypes.Account, spendableToken sdk.Coin, skip bool) {
	acc, _ = simtypes.RandomAcc(r, accs)
	tokens := getSpendableTokens(ctx, acc.Address, bk)
	spendableTokens := simtypes.RandSubsetCoins(r, tokens)
	if spendableTokens.Empty() {
		return acc, sdk.Coin{}, true
	}

	return acc, randomCoin(r, spendableTokens), false
}

// randomWithdrawFields returns a random account and an sdk.Coin from its uTokens
// (including both collateral and spendable uTokens). It returns skip=true
// if no uTokens were found.
func randomWithdrawFields(
	r *rand.Rand, ctx sdk.Context, accs []simtypes.Account,
	bk types.BankKeeper, lk keeper.Keeper,
) (acc simtypes.Account, withdrawal sdk.Coin, skip bool) {
	acc, _ = simtypes.RandomAcc(r, accs)
	uTokens := getSpendableUTokens(ctx, acc.Address, bk)
	uTokens = uTokens.Add(lk.GetBorrowerCollateral(ctx, acc.Address)...)
	uTokens = simtypes.RandSubsetCoins(r, uTokens)
	if uTokens.Empty() {
		return acc, sdk.Coin{}, true
	}

	return acc, randomCoin(r, uTokens), false
}

// randomCollateralizeFields returns a random account and a uToken from its spendable
// coins. It returns skip=true if the account has zero spendable uTokens.
func randomCollateralizeFields(
	r *rand.Rand, ctx sdk.Context, accs []simtypes.Account, bk types.BankKeeper,
) (acc simtypes.Account, spendableToken sdk.Coin, skip bool) {
	acc, _ = simtypes.RandomAcc(r, accs)
	uTokens := getSpendableUTokens(ctx, acc.Address, bk)
	uTokens = simtypes.RandSubsetCoins(r, uTokens)
	if uTokens.Empty() {
		return acc, sdk.Coin{}, true
	}

	return acc, randomCoin(r, uTokens), false
}

// randomDecollateralizeFields returns a random account and a uToken from its collateral
// coins. It returns skip=true if the account has zero collateral.
func randomDecollateralizeFields(
	r *rand.Rand, ctx sdk.Context, accs []simtypes.Account, lk keeper.Keeper,
) (acc simtypes.Account, spendableToken sdk.Coin, skip bool) {
	acc, _ = simtypes.RandomAcc(r, accs)
	uTokens := lk.GetBorrowerCollateral(ctx, acc.Address)
	uTokens = simtypes.RandSubsetCoins(r, uTokens)
	if uTokens.Empty() {
		return acc, sdk.Coin{}, true
	}

	return acc, randomCoin(r, uTokens), false
}

// randomBorrowFields returns a random account and an sdk.Coin from all
// the registered tokens with an random amount [0, 10^6].
// It returns skip=true if no registered token was found.
func randomBorrowFields(
	r *rand.Rand, ctx sdk.Context, accs []simtypes.Account, lk keeper.Keeper,
) (acc simtypes.Account, token sdk.Coin, skip bool) {
	acc, _ = simtypes.RandomAcc(r, accs)
	allTokens := lk.GetAllRegisteredTokens(ctx)
	if len(allTokens) == 0 {
		return acc, sdk.Coin{}, true
	}

	registeredToken := allTokens[r.Int31n(int32(len(allTokens)))]
	token = sdk.NewCoin(registeredToken.BaseDenom, simtypes.RandomAmount(r, sdkmath.NewInt(1_000000)))

	return acc, token, false
}

// randomRepayFields returns a random account and an sdk.Coin from an open borrow position.
// It returns skip=true if no borrow position was open.
func randomRepayFields(
	r *rand.Rand, ctx sdk.Context, accs []simtypes.Account, lk keeper.Keeper,
) (acc simtypes.Account, borrowToken sdk.Coin, skip bool) {
	acc, _ = simtypes.RandomAcc(r, accs)

	borrowTokens := simtypes.RandSubsetCoins(r, lk.GetBorrowerBorrows(ctx, acc.Address))
	if borrowTokens.Empty() {
		return acc, sdk.Coin{}, true
	}

	return acc, randomCoin(r, borrowTokens), false
}

// randomLiquidateFields returns two random accounts to be used as a liquidator
// and a borrower in a MsgLiquidate transaction, as well as a random sdk.Coin
// from the borrower's borrows and a random sdk.Coin from the borrower's collateral.
// It returns skip=true if no collateral is found.
func randomLiquidateFields(
	r *rand.Rand, _ sdk.Context, accs []simtypes.Account, _ keeper.Keeper,
) (
	liquidator simtypes.Account,
	borrower simtypes.Account,
	repaymentToken sdk.Coin, //nolint
	rewardDenom string,
	skip bool,
) {
	// note: liquidator and borrower might even be the same account
	liquidator, _ = simtypes.RandomAcc(r, accs)
	borrower, _ = simtypes.RandomAcc(r, accs)
	// TODO: evaluate whether we want liquidations in sims when we are enabling them for leverage
	return liquidator, borrower, sdk.Coin{}, "", true
}

func deliver(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, ak simulation.AccountKeeper,
	bk bankkeeper.Keeper, from simtypes.Account, msg sdk.Msg, coins sdk.Coins,
) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	cfg := testutil.MakeTestEncodingConfig()
	o := simulation.OperationInput{
		R:               r,
		App:             app,
		TxGen:           cfg.TxConfig,
		Cdc:             cfg.Codec.(*codec.ProtoCodec),
		Msg:             msg,
		Context:         ctx,
		SimAccount:      from,
		AccountKeeper:   ak,
		Bankkeeper:      bk,
		ModuleName:      types.ModuleName,
		CoinsSpentInMsg: coins,
	}

	// note: leverage operations are more expensive!
	return umeesim.GenAndDeliver(bk, o, appparams.DefaultGasLimit*50)
}
