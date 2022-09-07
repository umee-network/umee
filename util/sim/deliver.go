package sim

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	appparams "github.com/umee-network/umee/v3/app/params"
	"github.com/umee-network/umee/v3/util/coin"
)

// GenAndDeliverTxWithRandFees generates a transaction with a random fee and delivers it.
// If gasLimit==0 then appparams default gas limit is used.
func GenAndDeliver(o simulation.OperationInput, gasLimit sdk.Gas) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	if gasLimit == 0 {
		gasLimit = appparams.DefaultGasLimit
	}
	account := o.AccountKeeper.GetAccount(o.Context, o.SimAccount.Address)
	spendable := o.Bankkeeper.SpendableCoins(o.Context, account.GetAddress())

	_, hasNeg := spendable.SafeSub(o.CoinsSpentInMsg...)
	if hasNeg {
		return simtypes.NoOpMsg(o.ModuleName, o.MsgType, "message doesn't leave room for fees"), nil, nil
	}

	fees := coin.NewDecBld(appparams.MinMinGasPrice).
		Scale(int64(gasLimit)).ToCoins()
	return simulation.GenAndDeliverTx(o, fees)
}
