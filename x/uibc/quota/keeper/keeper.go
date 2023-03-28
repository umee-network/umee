package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	prefixstore "github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"

	"github.com/umee-network/umee/v4/x/uibc"
)

type Keeper struct {
	storeKey       storetypes.StoreKey
	cdc            codec.BinaryCodec
	leverageKeeper uibc.LeverageKeeper
	oracle         uibc.Oracle
	ics4Wrapper    porttypes.ICS4Wrapper
}

func NewKeeper(
	cdc codec.BinaryCodec, key storetypes.StoreKey, ics4Wrapper porttypes.ICS4Wrapper, leverageKeeper uibc.LeverageKeeper,
	oracleKeeper uibc.Oracle,
) Keeper {
	return Keeper{
		cdc:            cdc,
		storeKey:       key,
		ics4Wrapper:    ics4Wrapper,
		leverageKeeper: leverageKeeper,
		oracle:         oracleKeeper,
	}
}

// PrefixStore creates an new prefix store.
// It will automatically remove provided prefix from keys when using with the iterator.
func (k Keeper) PrefixStore(ctx *sdk.Context, prefix []byte) store.KVStore {
	s := ctx.KVStore(k.storeKey)
	return prefixstore.NewStore(s, prefix)
}
