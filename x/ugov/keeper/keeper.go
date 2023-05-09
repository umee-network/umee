package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Builder constructs Keeper by perparing all related dependencies (notably the store).
type Builder struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec
}

func NewKeeperBuilder(
	cdc codec.BinaryCodec, key storetypes.StoreKey) Builder {
	return Builder{
		cdc:      cdc,
		storeKey: key,
	}
}

func (kb Builder) Keeper(ctx *sdk.Context) Keeper {
	return Keeper{
		store: ctx.KVStore(kb.storeKey),
		cdc:   kb.cdc,

		ctx: ctx,
	}
}

// Keeper provides a light interface for module data access and transformation
type Keeper struct {
	store sdk.KVStore
	cdc   codec.BinaryCodec

	// TODO: ctx should be removed when we migrate leverage and oracle
	ctx *sdk.Context
}
