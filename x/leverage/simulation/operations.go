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
	DefaultWeightMsgSupply            int = 100
	DefaultWeightMsgWithdraw          int = 85
	DefaultWeightMsgBorrow            int = 80
	DefaultWeightMsgCollateralize     int = 60
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
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak types.AccountKeeper, bk types.BankKeeper,
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
	appParams.GetOrGenerate(cdc, OperationWeightMsgSupply, &weightMsgSupply, nil,
		func(_ *rand.Rand) {
			weightMsgSupply = DefaultWeightMsgSupply
		},
	)
	appParams.GetOrGenerate(cdc, OperationWeightMsgWithdraw, &weightMsgWithdraw, nil,
		func(_ *rand.Rand) {
			weightMsgWithdraw = DefaultWeightMsgWithdraw
		},
	)
	appParams.GetOrGenerate(cdc, OperationWeightMsgBorrow, &weightMsgBorrow, nil,
		func(_ *rand.Rand) {
			weightMsgBorrow = DefaultWeightMsgBorrow
		},
	)
	appParams.GetOrGenerate(cdc, OperationWeightMsgCollateralize, &weightMsgCollateralize, nil,
		func(_ *rand.Rand) {
			weightMsgCollateralize = DefaultWeightMsgCollateralize
		},
	)
	appParams.GetOrGenerate(cdc, OperationWeightMsgDecollateralize, &weightMsgDecollateralize, nil,
		func(_ *rand.Rand) {
			weightMsgDecollateralize = DefaultWeightMsgDecollateralize
		},
	)
	appParams.GetOrGenerate(cdc, OperationWeightMsgRepay, &weightMsgRepay, nil,
		func(_ *rand.Rand) {
			weightMsgRepay = DefaultWeightMsgRepay
		},
	)
	appParams.GetOrGenerate(cdc, OperationWeightMsgLiquidate, &weightMsgLiquidate, nil,
		func(_ *rand.Rand) {
			weightMsgLiquidate = DefaultWeightMsgLiquidate
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgSupply,
			SimulateMsgSupply(ak, bk),
		),
		simulation.NewWeightedOperation(
			weightMsgWithdraw,
			SimulateMsgWithdraw(ak, bk, lk),
		),
		simulation.NewWeightedOperation(
			weightMsgBorrow,
			SimulateMsgBorrow(ak, bk, lk),
		),
		simulation.NewWeightedOperation(
			weightMsgCollateralize,
			SimulateMsgCollateralize(ak, bk, lk),
		),
		simulation.NewWeightedOperation(
			weightMsgDecollateralize,
			SimulateMsgDecollateralize(ak, bk, lk),
		),
		simulation.NewWeightedOperation(
			weightMsgRepay,
			SimulateMsgRepay(ak, bk, lk),
		),
		simulation.NewWeightedOperation(
			weightMsgLiquidate,
			SimulateMsgLiquidate(ak, bk, lk),
		),
	}
}

// SimulateMsgSupply tests and runs a single msg supply where
// an account supplies some available assets.
func SimulateMsgSupply(ak simulation.AccountKeeper, bk types.BankKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		from, coin, skip := randomSupplyFields(r, ctx, accs, bk)
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeSupply, "skip all transfers"), nil, nil
		}

		msg := types.NewMsgSupply(from.Address, coin)

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         types.EventTypeSupply,
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

// SimulateMsgWithdraw tests and runs a single msg withdraw where
// an account attempts to withdraw some supplied assets.
func SimulateMsgWithdraw(ak simulation.AccountKeeper, bk types.BankKeeper, lk keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		from, withdrawUToken, skip := randomWithdrawFields(r, ctx, accs, bk, lk)
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeWithdraw, "skip all transfers"), nil, nil
		}

		msg := types.NewMsgWithdraw(from.Address, withdrawUToken)

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         types.EventTypeWithdraw,
			Context:         ctx,
			SimAccount:      from,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: sdk.NewCoins(withdrawUToken),
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgBorrow tests and runs a single msg borrow where
// an account attempts to borrow some assets.
func SimulateMsgBorrow(ak simulation.AccountKeeper, bk types.BankKeeper, lk keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		from, token, skip := randomBorrowFields(r, ctx, accs, lk)
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeBorrow, "skip all transfers"), nil, nil
		}

		msg := types.NewMsgBorrow(from.Address, token)

		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:           nil,
			Msg:           msg,
			MsgType:       types.EventTypeBorrow,
			Context:       ctx,
			SimAccount:    from,
			AccountKeeper: ak,
			Bankkeeper:    bk,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgCollateralize tests and runs a single msg which adds
// some collateral to a user.
func SimulateMsgCollateralize(
	ak simulation.AccountKeeper,
	bk types.BankKeeper,
	lk keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		from, collateral, skip := randomCollateralizeFields(r, ctx, accs, bk)
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeCollateralize, "skip all transfers"), nil, nil
		}

		msg := types.NewMsgCollateralize(from.Address, collateral)

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         types.EventTypeCollateralize,
			Context:         ctx,
			SimAccount:      from,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: sdk.NewCoins(collateral),
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgDecollateralize tests and runs a single msg which removes
// some collateral from a user.
func SimulateMsgDecollateralize(
	ak simulation.AccountKeeper,
	bk types.BankKeeper,
	lk keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		from, collateral, skip := randomDecollateralizeFields(r, ctx, accs, lk)
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeDecollateralize, "skip all transfers"), nil, nil
		}

		msg := types.NewMsgDecollateralize(from.Address, collateral)

		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:           nil,
			Msg:           msg,
			MsgType:       types.EventTypeDecollateralize,
			Context:       ctx,
			SimAccount:    from,
			AccountKeeper: ak,
			Bankkeeper:    bk,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgRepay tests and runs a single msg repay where
// an account repays some borrowed assets.
func SimulateMsgRepay(ak simulation.AccountKeeper, bk types.BankKeeper, lk keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		from, repayToken, skip := randomRepayFields(r, ctx, accs, lk)
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeRepay, "skip all transfers"), nil, nil
		}

		msg := types.NewMsgRepay(from.Address, repayToken)

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         types.EventTypeRepay,
			Context:         ctx,
			SimAccount:      from,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: sdk.NewCoins(repayToken),
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

		msg := types.NewMsgLiquidate(liquidator.Address, borrower.Address, repaymentToken, rewardDenom)

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         types.EventTypeLiquidate,
			Context:         ctx,
			SimAccount:      liquidator,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: sdk.NewCoins(repaymentToken),
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

// getSpendableTokens returns all non-uTokens from an account's spendable coins.
func getSpendableTokens(ctx sdk.Context, addr sdk.AccAddress, bk types.BankKeeper) sdk.Coins {
	tokens := sdk.NewCoins()
	for _, coin := range bk.SpendableCoins(ctx, addr) {
		if !types.HasUTokenPrefix(coin.Denom) {
			tokens = tokens.Add(coin)
		}
	}

	return tokens
}

// getSpendableUTokens returns all uTokens from an account's spendable coins.
func getSpendableUTokens(ctx sdk.Context, addr sdk.AccAddress, bk types.BankKeeper) sdk.Coins {
	uTokens := sdk.NewCoins()
	for _, coin := range bk.SpendableCoins(ctx, addr) {
		if types.HasUTokenPrefix(coin.Denom) {
			uTokens = uTokens.Add(coin)
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
// the registered tokens with an random amount [0, 150].
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
	token = sdk.NewCoin(registeredToken.BaseDenom, simtypes.RandomAmount(r, sdk.NewInt(150)))

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

	collateral := lk.GetBorrowerCollateral(ctx, borrower.Address)
	if collateral.Empty() {
		return liquidator, borrower, sdk.Coin{}, "", true
	}

	borrowed := lk.GetBorrowerBorrows(ctx, borrower.Address)

	borrowed = simtypes.RandSubsetCoins(r, borrowed)
	if borrowed.Empty() {
		return liquidator, borrower, sdk.Coin{}, "", true
	}

	rewardDenom = types.ToTokenDenom(randomCoin(r, collateral).Denom)

	return liquidator, borrower, randomCoin(r, borrowed), rewardDenom, false
}
