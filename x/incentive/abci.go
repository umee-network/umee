package incentive

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/umee-network/umee/v2/x/incentive/keeper"
)

// EndBlocker implements EndBlock for the x/incentive module.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	// TODO: Rewards are distributed
	return []abci.ValidatorUpdate{}
}
