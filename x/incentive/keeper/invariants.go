package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInvariants registers empty incentive module invariants
func RegisterInvariants(_ sdk.InvariantRegistry, _ Keeper) {}
