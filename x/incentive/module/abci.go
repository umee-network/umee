package module

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/umee-network/umee/v4/x/incentive/keeper"
)

// EndBlocker implements EndBlock for the x/incentive module.
func EndBlocker(_ sdk.Context, _ keeper.Keeper) []abci.ValidatorUpdate {
	// TODO: Programs change in status, and reward distribution amounts are updated
	return []abci.ValidatorUpdate{}
}
