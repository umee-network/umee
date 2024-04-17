package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/x/auction"
	"github.com/umee-network/umee/v6/x/ugov"
)

type SubAccounts struct {
	Rewards    []byte
	RewardsBid []byte
}

type Builder struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	bank     auction.BankKeeper
	ugov     ugov.EmergencyGroupBuilder
	accs     SubAccounts
}

func NewBuilder(cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	accs SubAccounts,
	b auction.BankKeeper,
	ugov ugov.EmergencyGroupBuilder) Builder {

	return Builder{cdc: cdc, storeKey: key, accs: accs,
		bank: b, ugov: ugov}
}

func (kb Builder) Keeper(ctx *sdk.Context) Keeper {
	return Keeper{
		store: ctx.KVStore(kb.storeKey),
		cdc:   kb.cdc,
		accs:  kb.accs,
		bank:  kb.bank,
		ugov:  kb.ugov(ctx),

		ctx: ctx,
	}
}

type Keeper struct {
	store sdk.KVStore
	cdc   codec.BinaryCodec
	accs  SubAccounts
	bank  auction.BankKeeper
	ugov  ugov.WithEmergencyGroup

	ctx *sdk.Context
}

func (k Keeper) sendCoins(from, to sdk.AccAddress, amount sdk.Coins) error {
	return k.bank.SendCoins(*k.ctx, from, to, amount)
}
