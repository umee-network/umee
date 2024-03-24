package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Builder struct {
	cdc      codec.Codec
	storeKey storetypes.StoreKey
}

func NewBuilder(cdc codec.Codec, key storetypes.StoreKey) Builder {
	return Builder{cdc: cdc, storeKey: key}
}

func (kb Builder) Keeper(ctx *sdk.Context) Keeper {
	return Keeper{
		ctx: ctx,
	}
}

type Keeper struct {
	ctx *sdk.Context
}
