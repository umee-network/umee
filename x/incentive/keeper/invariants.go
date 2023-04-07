package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/x/incentive"
)

const (
	routeNextID = "next-id"
)

// RegisterInvariants registers the incentive module invariants
func RegisterInvariants(_ sdk.InvariantRegistry, _ Keeper) {
	// ir.RegisterRoute(incentive.ModuleName, routeNextID, NextIDInvariant(k))
}

// NextIDInvariant checks that next ID is nonzero
func NextIDInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// TODO:
		// we can also verify that
		// next ID = len(get all programs) + 1
		// and that all program IDs are < next ID
		// and that no program IDs repeat

		nid := k.getNextProgramID(ctx)
		broken := nid == 0

		return sdk.FormatInvariant(
			incentive.ModuleName, routeNextID,
			fmt.Sprintf("invalid next ID: %d", nid),
		), broken
	}
}
