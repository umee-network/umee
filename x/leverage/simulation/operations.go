package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/umee-network/umee/v2/x/leverage/keeper"
	"github.com/umee-network/umee/v2/x/leverage/types"
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

// SimulateMsgLendAsset tests and runs a single msg lend where
// an account lends some available assets.
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

// SimulateMsgWithdrawAsset tests and runs a single msg withdraw where
// an account attempts to withdraw some loaned assets.
func SimulateMsgWithdrawAsset(ak simulation.AccountKeeper, bk types.BankKeeper, lk keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		from, withdrawUToken, skip := randomWithdrawFields(r, ctx, accs, bk, lk)
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeWithdrawLoanedAsset, "skip all transfers"), nil, nil
		}

		msg := types.NewMsgWithdrawAsset(from.Address, withdrawUToken)

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

// SimulateMsgBorrowAsset tests and runs a single msg borrow where
// an account attempts to borrow some assets.
func SimulateMsgBorrowAsset(ak simulation.AccountKeeper, bk types.BankKeeper, lk keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		from, token, skip := randomTokenFields(r, ctx, accs, lk)
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeBorrowAsset, "skip all transfers"), nil, nil
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

// SimulateMsgSetCollateralSetting tests and runs a single msg set collateral
// where an account enables or disables a uToken denom as collateral.
func SimulateMsgSetCollateralSetting(
	ak simulation.AccountKeeper,
	bk types.BankKeeper,
	lk keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		from, token, skip := randomTokenFields(r, ctx, accs, lk)
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

// SimulateMsgRepayAsset tests and runs a single msg repay where
// an account repays some borrowed assets.
func SimulateMsgRepayAsset(ak simulation.AccountKeeper, bk types.BankKeeper, lk keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		from, borrowToken, skip := randomBorrowedFields(r, ctx, accs, lk)
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeRepayBorrowedAsset, "skip all transfers"), nil, nil
		}

		msg := types.NewMsgRepayAsset(from.Address, borrowToken)

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

// SimulateMsgLiquidate tests and runs a single msg liquidate where
// one user attempts to liquidate another user's borrow.
func SimulateMsgLiquidate(ak simulation.AccountKeeper, bk types.BankKeeper, lk keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		liquidator, borrower, repaymentToken, rewardDenom, skip := randomLiquidateFields(r, ctx, accs, lk)
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeLiquidate, "skip all transfers"), nil, nil
		}

		msg := types.NewMsgLiquidate(liquidator.Address, borrower.Address, repaymentToken, sdk.NewInt64Coin(rewardDenom, 0))

		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:           nil,
			Msg:           msg,
			MsgType:       types.EventTypeLiquidate,
			Context:       ctx,
			SimAccount:    liquidator,
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

// randomTokenFields returns a random account and an sdk.Coin from all
// the registered tokens with an random amount [0, 150].
// It returns skip=true if no registered token was found.
func randomTokenFields(
	r *rand.Rand, ctx sdk.Context, accs []simtypes.Account, lk keeper.Keeper,
) (acc simtypes.Account, token sdk.Coin, skip bool) {
	acc, _ = simtypes.RandomAcc(r, accs)

	allTokens, err := lk.GetAllRegisteredTokens(ctx)
	if err != nil {
		panic(err)
	}
	if len(allTokens) == 0 {
		return acc, sdk.Coin{}, true
	}

	registeredToken := allTokens[r.Int31n(int32(len(allTokens)))]
	token = sdk.NewCoin(registeredToken.BaseDenom, simtypes.RandomAmount(r, sdk.NewInt(150)))

	return acc, token, false
}

// randomWithdrawFields returns a random account and an sdk.Coin from its uTokens
// (including both collateral and spendable uTokens). It returns skip=true
// if no uTokens were found.
func randomWithdrawFields(
	r *rand.Rand, ctx sdk.Context, accs []simtypes.Account,
	bk types.BankKeeper, lk keeper.Keeper,
) (acc simtypes.Account, withdrawal sdk.Coin, skip bool) {
	acc, _ = simtypes.RandomAcc(r, accs)

	uTokens := getSpendableUTokens(ctx, acc.Address, bk, lk)
	collateral, err := lk.GetBorrowerCollateral(ctx, acc.Address)
	if err != nil {
		panic(err)
	}
	uTokens = uTokens.Add(collateral...)

	uTokens = simtypes.RandSubsetCoins(r, uTokens)

	if uTokens.Empty() {
		return acc, sdk.Coin{}, true
	}

	return acc, randomCoin(r, uTokens), false
}

// getSpendableUTokens returns all uTokens from an account's spendable coins.
func getSpendableUTokens(
	ctx sdk.Context, addr sdk.AccAddress,
	bk types.BankKeeper, lk keeper.Keeper,
) sdk.Coins {
	uTokens := sdk.NewCoins()
	for _, coin := range bk.SpendableCoins(ctx, addr) {
		if lk.IsAcceptedUToken(ctx, coin.Denom) {
			uTokens = uTokens.Add(coin)
		}
	}

	return uTokens
}

// randomBorrowedFields returns a random account and an sdk.Coin from an open borrow position.
// It returns skip=true if no borrow position was open.
func randomBorrowedFields(
	r *rand.Rand, ctx sdk.Context, accs []simtypes.Account, lk keeper.Keeper,
) (acc simtypes.Account, borrowToken sdk.Coin, skip bool) {
	acc, _ = simtypes.RandomAcc(r, accs)
	b, err := lk.GetBorrowerBorrows(ctx, acc.Address)
	if err != nil {
		panic(err)
	}
	borrowTokens := simtypes.RandSubsetCoins(r, b)
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
	r *rand.Rand, ctx sdk.Context, accs []simtypes.Account, lk keeper.Keeper,
) (
	liquidator simtypes.Account,
	borrower simtypes.Account,
	repaymentToken sdk.Coin,
	rewardDenom string,
	skip bool,
) {
	idxLiquidator := r.Intn(len(accs) - 1)

	liquidator = accs[idxLiquidator]
	borrower = accs[idxLiquidator+1]

	collateral, err := lk.GetBorrowerCollateral(ctx, borrower.Address)
	if err != nil {
		panic(err)
	}
	if collateral.Empty() {
		return liquidator, borrower, sdk.Coin{}, "", true
	}

	borrowed, err := lk.GetBorrowerBorrows(ctx, borrower.Address)
	if err != nil {
		panic(err)
	}
	borrowed = simtypes.RandSubsetCoins(r, borrowed)
	if borrowed.Empty() {
		return liquidator, borrower, sdk.Coin{}, "", true
	}

	rewardDenom = lk.FromUTokenToTokenDenom(ctx, randomCoin(r, collateral).Denom)

	return liquidator, borrower, randomCoin(r, borrowed), rewardDenom, false
}
