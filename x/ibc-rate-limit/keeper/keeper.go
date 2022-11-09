package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/umee-network/umee/v3/x/ibc-rate-limit/types"
)

type Keeper struct {
	storeKey    storetypes.StoreKey
	cdc         codec.BinaryCodec
	paramSpace  paramtypes.Subspace
	ics4Wrapper types.ICS4Wrapper
	authority   string // the gov module account
}

func NewKeeper(
	cdc codec.BinaryCodec, key storetypes.StoreKey, paramSpace paramtypes.Subspace, ics4Wrapper types.ICS4Wrapper,
	authority string,
) Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:         cdc,
		storeKey:    key,
		paramSpace:  paramSpace,
		ics4Wrapper: ics4Wrapper,
		authority:   authority,
	}
}

// UpdateIBCTansferStatus update the ibc pause status in module params.
func (k Keeper) UpdateIBCTansferStatus(ctx sdk.Context, ibcStatus bool) error {
	var ibcPause bool
	k.paramSpace.Get(ctx, types.KeyIBCPause, &ibcPause)

	if ibcPause == ibcPause {
		return types.ErrIBCPauseStatus
	}

	// update the ibc status
	k.paramSpace.Set(ctx, types.KeyIBCPause, ibcStatus)

	return nil
}
