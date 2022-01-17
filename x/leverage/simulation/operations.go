package simulation

import (
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/umee-network/umee/x/leverage/keeper"
	"github.com/umee-network/umee/x/oracle/types"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak types.AccountKeeper, bk keeper.Keeper,
) simulation.WeightedOperations {

	// var weightMsgSend, weightMsgMultiSend int
	// appParams.GetOrGenerate(cdc, OpWeightMsgSend, &weightMsgSend, nil,
	// 	func(_ *rand.Rand) {
	// 		weightMsgSend = simappparams.DefaultWeightMsgSend
	// 	},
	// )

	// appParams.GetOrGenerate(cdc, OpWeightMsgMultiSend, &weightMsgMultiSend, nil,
	// 	func(_ *rand.Rand) {
	// 		weightMsgMultiSend = simappparams.DefaultWeightMsgMultiSend
	// 	},
	// )

	return simulation.WeightedOperations{
		// simulation.NewWeightedOperation(
		// 	weightMsgSend,
		// 	SimulateMsgSend(ak, bk),
		// ),
		// simulation.NewWeightedOperation(
		// 	weightMsgMultiSend,
		// 	SimulateMsgMultiSend(ak, bk),
		// ),
	}
}
