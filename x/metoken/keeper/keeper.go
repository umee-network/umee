package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/umee-network/umee/v4/x/metoken"
)

type Keeper struct {
	cdc      codec.Codec
	storeKey storetypes.StoreKey

	//todo: add bank, leverage and oracle keepers here and in the NewKeeper function
}

// NewKeeper  creates a new Keeper object
func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
}

// Logger returns module Logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+metoken.ModuleName)
}

// KVStore returns the module's KVStore
func (k Keeper) KVStore(ctx sdk.Context) sdk.KVStore {
	return ctx.KVStore(k.storeKey)
}
