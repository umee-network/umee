package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/umee-network/umee/x/leverage/types"
)

// Default simulation operation weights for leverage messages
const (
	DefaultWeightMsgLendAsset int = 100
	OpWeightMsgLendAsset          = "op_weight_msg_lend_asset"
)

// WeightedOperations returns all the operations from the leverage module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak types.AccountKeeper, bk types.BankKeeper,
) simulation.WeightedOperations {

	var weightMsgLend int
	appParams.GetOrGenerate(cdc, OpWeightMsgLendAsset, &weightMsgLend, nil,
		func(_ *rand.Rand) {
			weightMsgLend = DefaultWeightMsgLendAsset
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgLend,
			SimulateMsgLendAsset(ak, bk),
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

// randomSendFields returns the sender and recipient simulation accounts as well
// as the transferred amount.
func randomSendFields(
	r *rand.Rand, ctx sdk.Context, accs []simtypes.Account, bk types.BankKeeper,
) (simtypes.Account, sdk.Coins, bool) {
	from, _ := simtypes.RandomAcc(r, accs)

	spendableBalances := bk.SpendableCoins(ctx, from.Address)

	sendCoins := simtypes.RandSubsetCoins(r, spendableBalances)
	if sendCoins.Empty() {
		return from, nil, true
	}

	return from, sendCoins, false
}
