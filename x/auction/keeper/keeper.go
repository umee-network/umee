package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/x/auction"
	"github.com/umee-network/umee/v6/x/ugov"
)

type SubAccounts struct {
	// Account used to collect rewards
	RewardsCollect []byte
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

//
// helper functions
//

func (k Keeper) sendCoins(from, to sdk.AccAddress, amounts sdk.Coins) error {
	return k.bank.SendCoins(*k.ctx, from, to, amounts)
}
func (k Keeper) sendFromModule(to sdk.AccAddress, amounts ...sdk.Coin) error {
	return k.bank.SendCoinsFromModuleToAccount(*k.ctx, auction.ModuleName, to, amounts)
}
func (k Keeper) sendToModule(from sdk.AccAddress, amounts ...sdk.Coin) error {
	return k.bank.SendCoinsFromAccountToModule(*k.ctx, from, auction.ModuleName, amounts)
}
