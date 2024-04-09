package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/x/auction"
)

type Builder struct {
	cdc      codec.Codec
	storeKey storetypes.StoreKey
	bank     auction.BankKeeper
}

func NewBuilder(cdc codec.Codec, key storetypes.StoreKey, b auction.BankKeeper) Builder {
	return Builder{cdc: cdc, storeKey: key, bank: b}
}

func (kb Builder) Keeper(ctx *sdk.Context) Keeper {
	return Keeper{
		ctx: ctx,
	}
}

type Keeper struct {
	ctx *sdk.Context
}
