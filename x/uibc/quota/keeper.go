package quota

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	prefixstore "github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/x/ugov"
	"github.com/umee-network/umee/v6/x/uibc"
)

// Builder constructs Keeper by perparing all related dependencies (notably the store).
type Builder struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec
	leverage uibc.Leverage
	oracle   uibc.Oracle
	ugov     ugov.EmergencyGroupBuilder
}

func NewBuilder(
	cdc codec.BinaryCodec, key storetypes.StoreKey,
	leverage uibc.Leverage, oracle uibc.Oracle, ugov ugov.EmergencyGroupBuilder,
) Builder {
	return Builder{
		cdc:      cdc,
		storeKey: key,
		leverage: leverage,
		oracle:   oracle,
		ugov:     ugov,
	}
}

func (kb Builder) Keeper(ctx *sdk.Context) Keeper {
	return Keeper{
		store:     ctx.KVStore(kb.storeKey),
		leverage:  kb.leverage,
		oracle:    kb.oracle,
		ugov:      kb.ugov(ctx),
		cdc:       kb.cdc,
		blockTime: ctx.BlockTime(),

		ctx: ctx,
	}
}

type Keeper struct {
	// Note: ideally we use a ligther interface here to directly use cosmos-db/DB
	// however we will need to wait probably until Cosmos SDK 0.48
	// We can have multiple stores if needed
	store    sdk.KVStore
	leverage uibc.Leverage
	oracle   uibc.Oracle
	ugov     ugov.WithEmergencyGroup

	/**
	if Keeper methods depends on sdk.Context, then we should add those dependencies directly,
	or provide them as function arguments.
	   gasMeter    sdk.GasMeter

	Sometimes, all types don't have Any. In that case we don't codec, and those types can be
	serialized directly using `bz := protoObject.Marshal()`.
	*/
	blockTime time.Time
	cdc       codec.BinaryCodec

	// TODO: ctx should be removed when we migrate leverage and oracle
	ctx *sdk.Context
}

// PrefixStore creates an new prefix store.
// It will automatically remove provided prefix from keys when using with the iterator.
func (k Keeper) PrefixStore(prefix []byte) store.KVStore {
	return prefixstore.NewStore(k.store, prefix)
}
