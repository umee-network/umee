package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

/*
  TODO: Remove or rename this file after writing the proper functions to access governance params.

  Update: Use GetRegisteredToken once it exists, then extract the relevant field.

  Except: GetBorrowInterestEpoch should use x/params, not the token registry.
*/

// GetBorrowInterestEpoch gets the borrow interest epoch.
func (k Keeper) GetBorrowInterestEpoch(ctx sdk.Context) (sdk.Int, error) {
	// TODO: Implement this by retrieving the appropriate governance parameter.
	defaultEpoch := sdk.NewInt(100)
	return defaultEpoch, nil
}
