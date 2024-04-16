package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	prefixstore "github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/x/auction"
	"github.com/umee-network/umee/v6/x/ugov"
)

type Builder struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	bank     auction.BankKeeper
	ugov     ugov.EmergencyGroupBuilder
}

func NewBuilder(cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	b auction.BankKeeper,
	ugov ugov.EmergencyGroupBuilder) Builder {

	return Builder{cdc: cdc, storeKey: key, bank: b, ugov: ugov}
}

func (kb Builder) Keeper(ctx *sdk.Context) Keeper {
	return Keeper{
		store: ctx.KVStore(kb.storeKey),
		cdc:   kb.cdc,
		bank:  kb.bank,
		ugov:  kb.ugov(ctx),

		ctx: ctx,
	}
}

type Keeper struct {
	store sdk.KVStore
	cdc   codec.BinaryCodec
	bank  auction.BankKeeper
	ugov  ugov.WithEmergencyGroup

	// TODO: ctx should be removed when we migrate leverage and oracle
	ctx *sdk.Context
}

// PrefixStore creates an new prefix store.
// It will automatically remove provided prefix from keys when using with the iterator.
func (k Keeper) PrefixStore(prefix []byte) store.KVStore {
	return prefixstore.NewStore(k.store, prefix)
}
