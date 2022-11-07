package ibc_rate_limit

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/umee-network/umee/v3/x/ibc-rate-limit/keeper"
)

// EndBlocker implements EndBlock for the x/leverage module.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
