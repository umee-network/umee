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
	OpWeightMsgLendAsset              = "op_weight_msg_lend_asset"
	OpWeightMsgWithdrawAsset          = "op_weight_msg_withdraw_asset"
)

// WeightedOperations returns all the operations from the leverage module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak types.AccountKeeper, bk types.BankKeeper,
	lk keeper.Keeper,
) simulation.WeightedOperations {

	var (
		weightMsgLend     int
		weightMsgWithdraw int
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

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgLend,
			SimulateMsgLendAsset(ak, bk),
		),
		simulation.NewWeightedOperation(
			weightMsgWithdraw,
			SimulateMsgWithdrawAsset(ak, bk, lk),
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
		from, coins, skip := randomSendFields(r, ctx, accs, bk)
		if coins == nil {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeLoanAsset, "Coins is nil"), nil, nil
		}
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeLoanAsset, "skip all transfers"), nil, nil
		}

		coin := coins[r.Int31n(int32(coins.Len()))]
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
		from, coins, skip := randomWithdrawFields(r, ctx, accs, lk)
		if coins == nil {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeWithdrawLoanedAsset, "Coins is nil"), nil, nil
		}
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeWithdrawLoanedAsset, "skip all transfers"), nil, nil
		}

		coin := coins[r.Int31n(int32(coins.Len()))]
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

// randomSendFields returns an random account and coins to spend in the simulation.
func randomSendFields(
	r *rand.Rand, ctx sdk.Context, accs []simtypes.Account, bk types.BankKeeper,
) (acc simtypes.Account, spendableCoins sdk.Coins, skip bool) {
	acc, _ = simtypes.RandomAcc(r, accs)

	spendableBalances := bk.SpendableCoins(ctx, acc.Address)

	spendableCoins = simtypes.RandSubsetCoins(r, spendableBalances)
	if spendableCoins.Empty() {
		return acc, nil, true
	}

	return acc, spendableCoins, false
}

// randomWithdrawFields returns an random account and coins to withdraw in the simulation.
func randomWithdrawFields(
	r *rand.Rand, ctx sdk.Context, accs []simtypes.Account, lk keeper.Keeper,
) (acc simtypes.Account, withdrawCoins sdk.Coins, skip bool) {
	acc, _ = simtypes.RandomAcc(r, accs)

	collateralBalances := lk.GetBorrowerCollateral(ctx, acc.Address)

	withdrawCoins = simtypes.RandSubsetCoins(r, collateralBalances)
	if withdrawCoins.Empty() {
		return acc, nil, true
	}

	return acc, withdrawCoins, false
}
