package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/umee-network/umee/x/leverage/types"
	oracletypes "github.com/umee-network/umee/x/oracle/types"
)

// Default simulation operation weights for leverage messages
const (
	DefaultWeightMsgLendAsset int = 100
	OpWeightMsgLendAsset          = "op_weight_msg_lend_asset"
)

// WeightedOperations returns all the operations from the leverage module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak oracletypes.AccountKeeper, bk types.BankKeeper,
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
func SimulateMsgLendAsset(ak oracletypes.AccountKeeper, bk types.BankKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		from, coins, skip := randomSendFields(r, ctx, accs, bk, ak)
		if coins == nil {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeLoanAsset, "Coins is nil"), nil, nil
		}
		coin := coins[r.Int31n(int32(coins.Len()))]

		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeLoanAsset, "skip all transfers"), nil, nil
		}

		msg := types.NewMsgLendAsset(from.Address, coin)

		err := sendMsgLendAsset(r, app, ak, bk, msg, ctx, chainID, []cryptotypes.PrivKey{from.PrivKey})
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.EventTypeLoanAsset, "invalid transfers"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, "", nil), nil, nil
	}
}

// randomSendFields returns the sender and recipient simulation accounts as well
// as the transferred amount.
func randomSendFields(
	r *rand.Rand, ctx sdk.Context, accs []simtypes.Account, bk types.BankKeeper, ak oracletypes.AccountKeeper,
) (simtypes.Account, sdk.Coins, bool) {
	from, _ := simtypes.RandomAcc(r, accs)

	acc := ak.GetAccount(ctx, from.Address)
	if acc == nil {
		return from, nil, true
	}

	accBalances := bk.GetAllBalances(ctx, acc.GetAddress())

	sendCoins := simtypes.RandSubsetCoins(r, accBalances)
	if sendCoins.Empty() {
		return from, nil, true
	}

	return from, sendCoins, false
}

// sendMsgLendAsset sends a transaction with a MsgLendAsset from a provided random account.
func sendMsgLendAsset(
	r *rand.Rand, app *baseapp.BaseApp, ak oracletypes.AccountKeeper,
	bk types.BankKeeper, msg *types.MsgLendAsset, ctx sdk.Context, chainID string, privkeys []cryptotypes.PrivKey,
) error {

	var (
		fees sdk.Coins
		err  error
	)

	from, err := sdk.AccAddressFromBech32(msg.Lender)
	if err != nil {
		return err
	}

	account := ak.GetAccount(ctx, from)
	spendableCoins := bk.GetAllBalances(ctx, account.GetAddress())

	coins, hasNeg := spendableCoins.SafeSub([]sdk.Coin{msg.Amount})
	if !hasNeg {
		fees, err = simtypes.RandomFees(r, ctx, coins)
		if err != nil {
			return err
		}
	}
	txGen := simappparams.MakeTestEncodingConfig().TxConfig
	tx, err := helpers.GenTx(
		txGen,
		[]sdk.Msg{msg},
		fees,
		helpers.DefaultGenTxGas,
		chainID,
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		privkeys...,
	)
	if err != nil {
		return err
	}

	_, _, err = app.Deliver(txGen.TxEncoder(), tx)
	if err != nil {
		return err
	}

	return nil
}
