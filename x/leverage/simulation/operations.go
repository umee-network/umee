package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/umee-network/umee/x/leverage/keeper"
	"github.com/umee-network/umee/x/leverage/types"
)

// Default simulation operation weights for leverage messages
const (
	DefaultWeightMsgLendAsset     int = 100
	DefaultWeightMsgWithdrawAsset int = 85
	DefaultWeightMsgBorrowAsset   int = 80
	OpWeightMsgLendAsset              = "op_weight_msg_lend_asset"
	OpWeightMsgWithdrawAsset          = "op_weight_msg_withdraw_asset"
	OpWeightMsgBorrowAsset            = "op_weight_msg_borrow_asset"
)

// WeightedOperations returns all the operations from the leverage module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak types.AccountKeeper, bk types.BankKeeper,
	lk keeper.Keeper,
) simulation.WeightedOperations {

	var (
		weightMsgLend     int
		weightMsgWithdraw int
		weightMsgBorrow   int
	)
	appParams.GetOrGenerate(cdc, OpWeightMsgLendAsset, &weightMsgLend, nil,
		func(_ *rand.Rand) {
			weightMsgLend = DefaultWeightMsgLendAsset
		},
	)
	appParams.GetOrGenerate(cdc, OpWeightMsgWithdrawAsset, &weightMsgWithdraw, nil,
		func(_ *rand.Rand) {
			weightMsgWithdraw = DefaultWeightMsgWithdrawAsset
		},
	)
	appParams.GetOrGenerate(cdc, OpWeightMsgBorrowAsset, &weightMsgBorrow, nil,
		func(_ *rand.Rand) {
			weightMsgBorrow = DefaultWeightMsgBorrowAsset
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgLend,
			SimulateMsgLendAsset(ak, bk),
		),
		simulation.NewWeightedOperation(
			weightMsgWithdraw,
			SimulateMsgWithdrawAsset(ak, bk, lk),
		),
		simulation.NewWeightedOperation(
			weightMsgBorrow,
			SimulateMsgBorrowAsset(ak, bk, lk),
		),
	}
}

// SimulateMsgLendAsset tests and runs a single msg send where
// an account lends some available asset.
func SimulateMsgLendAsset(ak simulation.AccountKeeper, bk types.BankKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		from, coin, skip := randomSendFields(r, ctx, accs, bk)
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeLoanAsset, "skip all transfers"), nil, nil
		}

		msg := types.NewMsgLendAsset(from.Address, coin)

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         types.EventTypeLoanAsset,
			Context:         ctx,
			SimAccount:      from,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: sdk.NewCoins(coin),
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgWithdrawAsset tests and runs a single msg send where
// an account withdraw some lended asset.
func SimulateMsgWithdrawAsset(ak simulation.AccountKeeper, bk types.BankKeeper, lk keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		from, coin, skip := randomCollateralFields(r, ctx, accs, lk)
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeWithdrawLoanedAsset, "skip all transfers"), nil, nil
		}

		msg := types.NewMsgWithdrawAsset(from.Address, coin)

		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:           nil,
			Msg:           msg,
			MsgType:       types.EventTypeWithdrawLoanedAsset,
			Context:       ctx,
			SimAccount:    from,
			AccountKeeper: ak,
			Bankkeeper:    bk,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgBorrowAsset tests and runs a single msg send where
// an account borrow some asset.
func SimulateMsgBorrowAsset(ak simulation.AccountKeeper, bk types.BankKeeper, lk keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		from, coin, skip := randomCollateralFields(r, ctx, accs, lk)
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeBorrowAsset, "skip all transfers"), nil, nil
		}

		msg := types.NewMsgBorrowAsset(from.Address, coin)

		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:           nil,
			Msg:           msg,
			MsgType:       types.EventTypeBorrowAsset,
			Context:       ctx,
			SimAccount:    from,
			AccountKeeper: ak,
			Bankkeeper:    bk,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// randomCoin get random coin from coins
func randomCoin(r *rand.Rand, coins sdk.Coins) sdk.Coin {
	if coins.Empty() {
		return sdk.Coin{}
	}
	return coins[r.Int31n(int32(coins.Len()))]
}

// randomSendFields returns an random account and coins to spend in the simulation.
func randomSendFields(
	r *rand.Rand, ctx sdk.Context, accs []simtypes.Account, bk types.BankKeeper,
) (acc simtypes.Account, spendableCoin sdk.Coin, skip bool) {
	acc, _ = simtypes.RandomAcc(r, accs)

	spendableBalances := bk.SpendableCoins(ctx, acc.Address)

	spendableCoins := simtypes.RandSubsetCoins(r, spendableBalances)
	if spendableCoins.Empty() {
		return acc, sdk.Coin{}, true
	}

	return acc, randomCoin(r, spendableCoins), false
}

// randomCollateralFields returns an random account and coins to withdraw in the simulation.
func randomCollateralFields(
	r *rand.Rand, ctx sdk.Context, accs []simtypes.Account, lk keeper.Keeper,
) (acc simtypes.Account, withdrawCoin sdk.Coin, skip bool) {
	acc, _ = simtypes.RandomAcc(r, accs)

	collateralBalances := lk.GetBorrowerCollateral(ctx, acc.Address)

	withdrawCoins := simtypes.RandSubsetCoins(r, collateralBalances)
	if withdrawCoins.Empty() {
		return acc, sdk.Coin{}, true
	}

	return acc, randomCoin(r, withdrawCoins), false
}
