package sim

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	appparams "github.com/umee-network/umee/v3/app/params"
	"github.com/umee-network/umee/v3/util/coin"
)

// GenAndDeliverTxWithRandFees generates a transaction with a random fee and delivers it.
// If gasLimit==0 then appparams default gas limit is used.
func GenAndDeliver(bk bankkeeper.Keeper, o simulation.OperationInput, gasLimit sdk.Gas,
) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

	if gasLimit == 0 {
		gasLimit = appparams.DefaultGasLimit
	}
	account := o.AccountKeeper.GetAccount(o.Context, o.SimAccount.Address)
	spendable := o.Bankkeeper.SpendableCoins(o.Context, account.GetAddress())

	_, hasNeg := spendable.SafeSub(o.CoinsSpentInMsg...)
	if hasNeg {
		return simtypes.NoOpMsg(o.ModuleName, o.MsgType, "message doesn't leave room for fees"), nil, nil
	}

	fees := coin.NewDecBld(appparams.ProtocolMinGasPrice).
		Scale(int64(gasLimit)).ToCoins()
	if _, hasNeg = spendable.SafeSub(fees...); hasNeg {
		fund := coin.NewDecBld(appparams.ProtocolMinGasPrice).
			Scale(int64(gasLimit * 1000)).ToCoins()
		err := banktestutil.FundAccount(bk, o.Context, o.SimAccount.Address, fund)
		if err != nil {
			return simtypes.NewOperationMsg(o.Msg, false, o.ModuleName, o.Cdc), nil,
				fmt.Errorf("can't fund account [%s] to pay fees; [%w]", o.SimAccount.Address, err)
		}
	}

	return simulation.GenAndDeliverTx(o, fees)
}
