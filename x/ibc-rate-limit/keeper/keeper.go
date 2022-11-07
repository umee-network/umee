package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
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
	return Keeper{
		cdc:         cdc,
		storeKey:    key,
		paramSpace:  paramSpace,
		ics4Wrapper: ics4Wrapper,
		authority:   authority,
	}
}
