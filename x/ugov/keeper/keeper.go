package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/x/ugov"
)

var _ ugov.Keeper = Keeper{}

// Builder constructs Keeper by perparing all related dependencies (notably the store).
type Builder struct {
	storeKey storetypes.StoreKey
	Cdc      codec.BinaryCodec
}

func NewBuilder(
	cdc codec.BinaryCodec, key storetypes.StoreKey,
) Builder {
	return Builder{
		Cdc:      cdc,
		storeKey: key,
	}
}

func (kb Builder) Keeper(ctx *sdk.Context) ugov.Keeper {
	return Keeper{
		store: ctx.KVStore(kb.storeKey),
		cdc:   kb.Cdc,
	}
}

// functions to downcast Keeper constructor into super types.

func (kb Builder) Params(ctx *sdk.Context) ugov.ParamsKeeper               { return kb.Keeper(ctx) }
func (kb Builder) EmergencyGroup(ctx *sdk.Context) ugov.WithEmergencyGroup { return kb.Keeper(ctx) }

// Keeper provides a light interface for module data access and transformation
type Keeper struct {
	store sdk.KVStore
	cdc   codec.BinaryCodec
}
