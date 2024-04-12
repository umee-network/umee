package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/x/auction"
	"github.com/umee-network/umee/v6/x/ugov"
)

type Builder struct {
	cdc      codec.Codec
	storeKey storetypes.StoreKey
	bank     auction.BankKeeper
	ugov     ugov.EmergencyGroupBuilder
}

func NewBuilder(cdc codec.Codec,
	key storetypes.StoreKey,
	b auction.BankKeeper,
	ugov ugov.EmergencyGroupBuilder) Builder {

	return Builder{cdc: cdc, storeKey: key, bank: b, ugov: ugov}
}

func (kb Builder) Keeper(ctx *sdk.Context) Keeper {
	return Keeper{
		ctx:  ctx,
		bank: kb.bank,
		ugov: kb.ugov(ctx),
	}
}

type Keeper struct {
	ctx  *sdk.Context
	bank auction.BankKeeper
	ugov ugov.WithEmergencyGroup
}
