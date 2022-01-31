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
	DefaultWeightMsgLendAsset       int = 100
	DefaultWeightMsgWithdrawAsset   int = 85
	DefaultWeightMsgBorrowAsset     int = 80
	DefaultWeightMsgSetCollateral   int = 60
	DefaultWeightMsgRepayAsset      int = 70
	DefaultWeightMsgLiquidate       int = 75
	OperationWeightMsgLendAsset         = "op_weight_msg_lend_asset"
	OperationWeightMsgWithdrawAsset     = "op_weight_msg_withdraw_asset"
	OperationWeightMsgBorrowAsset       = "op_weight_msg_borrow_asset"
	OperationWeightMsgSetCollateral     = "op_weight_msg_set_collateral"
	OperationWeightMsgRepayAsset        = "op_weight_msg_repay_asset"
	OperationWeightMsgLiquidate         = "op_weight_msg_liquidate"
)

// WeightedOperations returns all the operations from the leverage module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak types.AccountKeeper, bk types.BankKeeper,
	lk keeper.Keeper,
) simulation.WeightedOperations {

	var (
		weightMsgLend          int
		weightMsgWithdraw      int
		weightMsgBorrow        int
		weightMsgSetCollateral int
		weightMsgRepayAsset    int
		weightMsgLiquidate     int
	)
	appParams.GetOrGenerate(cdc, OperationWeightMsgLendAsset, &weightMsgLend, nil,
		func(_ *rand.Rand) {
			weightMsgLend = DefaultWeightMsgLendAsset
		},
	)
	appParams.GetOrGenerate(cdc, OperationWeightMsgWithdrawAsset, &weightMsgWithdraw, nil,
		func(_ *rand.Rand) {
			weightMsgWithdraw = DefaultWeightMsgWithdrawAsset
		},
	)
	appParams.GetOrGenerate(cdc, OperationWeightMsgBorrowAsset, &weightMsgBorrow, nil,
		func(_ *rand.Rand) {
			weightMsgBorrow = DefaultWeightMsgBorrowAsset
		},
	)
	appParams.GetOrGenerate(cdc, OperationWeightMsgSetCollateral, &weightMsgSetCollateral, nil,
		func(_ *rand.Rand) {
			weightMsgSetCollateral = DefaultWeightMsgSetCollateral
		},
	)
	appParams.GetOrGenerate(cdc, OperationWeightMsgRepayAsset, &weightMsgRepayAsset, nil,
		func(_ *rand.Rand) {
			weightMsgRepayAsset = DefaultWeightMsgRepayAsset
		},
	)
	appParams.GetOrGenerate(cdc, OperationWeightMsgLiquidate, &weightMsgLiquidate, nil,
		func(_ *rand.Rand) {
			weightMsgLiquidate = DefaultWeightMsgLiquidate
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
		simulation.NewWeightedOperation(
			weightMsgSetCollateral,
			SimulateMsgSetCollateralSetting(ak, bk, lk),
		),
		simulation.NewWeightedOperation(
			weightMsgRepayAsset,
			SimulateMsgRepayAsset(ak, bk, lk),
		),
		simulation.NewWeightedOperation(
			weightMsgLiquidate,
			SimulateMsgLiquidate(ak, bk, lk),
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
		from, coin, skip := randomSpendableFields(r, ctx, accs, bk)
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
		from, uToken, skip := randomCollateralFields(r, ctx, accs, lk)
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeWithdrawLoanedAsset, "skip all transfers"), nil, nil
		}

		msg := types.NewMsgWithdrawAsset(from.Address, uToken)

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
		from, uToken, skip := randomCollateralFields(r, ctx, accs, lk)
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeBorrowAsset, "skip all transfers"), nil, nil
		}

		token, err := lk.ExchangeUToken(ctx, uToken)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeBorrowAsset, "error exchange uToken"), nil, err
		}
		msg := types.NewMsgBorrowAsset(from.Address, token)

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

// SimulateMsgSetCollateralSetting tests and runs a single msg send where
// an account set some denom as collateral.
func SimulateMsgSetCollateralSetting(
	ak simulation.AccountKeeper,
	bk types.BankKeeper,
	lk keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		from, token, skip := randomSpendableFields(r, ctx, accs, bk)
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeSetCollateralSetting, "skip all transfers"), nil, nil
		}

		uDenom := lk.FromTokenToUTokenDenom(ctx, token.Denom)
		enable := lk.GetCollateralSetting(ctx, from.Address.Bytes(), uDenom)
		msg := types.NewMsgSetCollateral(from.Address, uDenom, !enable)

		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:           nil,
			Msg:           msg,
			MsgType:       types.EventTypeSetCollateralSetting,
			Context:       ctx,
			SimAccount:    from,
			AccountKeeper: ak,
			Bankkeeper:    bk,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgRepayAsset tests and runs a single msg send where
// an account repay some asset borrowed.
func SimulateMsgRepayAsset(ak simulation.AccountKeeper, bk types.BankKeeper, lk keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		from, uToken, skip := randomCollateralFields(r, ctx, accs, lk)
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeRepayBorrowedAsset, "skip all transfers"), nil, nil
		}

		token, err := lk.ExchangeUToken(ctx, uToken)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeRepayBorrowedAsset, "error exchange uToken"), nil, err
		}
		msg := types.NewMsgRepayAsset(from.Address, token)

		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:           nil,
			Msg:           msg,
			MsgType:       types.EventTypeRepayBorrowedAsset,
			Context:       ctx,
			SimAccount:    from,
			AccountKeeper: ak,
			Bankkeeper:    bk,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgLiquidate tests and runs a single msg send where
// some asset borrowed is liquidated.
func SimulateMsgLiquidate(ak simulation.AccountKeeper, bk types.BankKeeper, lk keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		liquidator, borrower, repaymentToken, uRewardToken, skip := randomLiquidateFields(r, ctx, accs, lk)
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeLiquidate, "skip all transfers"), nil, nil
		}

		msg := types.NewMsgLiquidate(liquidator.Address, borrower.Address, repaymentToken, uRewardToken.Denom)

		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:           nil,
			Msg:           msg,
			MsgType:       types.EventTypeLiquidate,
			Context:       ctx,
			SimAccount:    borrower,
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

// randomSpendableFields returns a random account and an sdk.Coin from its spendable
// coins. It returns skip=true if the account has zero spendable coins.
func randomSpendableFields(
	r *rand.Rand, ctx sdk.Context, accs []simtypes.Account, bk types.BankKeeper,
) (acc simtypes.Account, spendableToken sdk.Coin, skip bool) {
	acc, _ = simtypes.RandomAcc(r, accs)

	spendableBalances := bk.SpendableCoins(ctx, acc.Address)

	spendableTokens := simtypes.RandSubsetCoins(r, spendableBalances)
	if spendableTokens.Empty() {
		return acc, sdk.Coin{}, true
	}

	return acc, randomCoin(r, spendableTokens), false
}

// randomCollateralFields returns a random account and an sdk.Coin from its collateral.
// It returns skip=true if no collateral was found.
func randomCollateralFields(
	r *rand.Rand, ctx sdk.Context, accs []simtypes.Account, lk keeper.Keeper,
) (acc simtypes.Account, withdrawToken sdk.Coin, skip bool) {
	acc, _ = simtypes.RandomAcc(r, accs)

	uRewardTokens := lk.GetBorrowerCollateral(ctx, acc.Address)
	if uRewardTokens.Empty() {
		return acc, sdk.Coin{}, true
	}

	return acc, randomCoin(r, uRewardTokens), false
}

// randomLiquidateFields returns two random accounts to be used as a liquidator
// and a borrower in a MsgLiquidate transaction, as well as a random sdk.Coin
// open borrower position and a random sdk.Coin from the borrower's collateral.
// It returns skip=true if no collateral is found.
func randomLiquidateFields(
	r *rand.Rand, ctx sdk.Context, accs []simtypes.Account, lk keeper.Keeper,
) (
	liquidator simtypes.Account,
	borrower simtypes.Account,
	repaymentToken sdk.Coin,
	uRewardToken sdk.Coin,
	skip bool,
) {
	idxLiquidator := r.Intn(len(accs) - 1)

	liquidator = accs[idxLiquidator]
	borrower = accs[idxLiquidator+1]

	uRewardTokens := lk.GetBorrowerCollateral(ctx, borrower.Address)
	if uRewardTokens.Empty() {
		return liquidator, borrower, sdk.Coin{}, sdk.Coin{}, true
	}

	repaymentTokens := lk.GetBorrowerBorrows(ctx, borrower.Address)
	if uRewardTokens.Empty() {
		return liquidator, borrower, sdk.Coin{}, sdk.Coin{}, true
	}

	return liquidator, borrower, randomCoin(r, repaymentTokens), randomCoin(r, uRewardTokens), false
}
