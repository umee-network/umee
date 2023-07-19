package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	prefixstore "github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/umee-network/umee/v5/util/store"
	"github.com/umee-network/umee/v5/x/incentive"
)

type Keeper struct {
	cdc            codec.Codec
	storeKey       storetypes.StoreKey
	bankKeeper     incentive.BankKeeper
	leverageKeeper incentive.LeverageKeeper
}

func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	bk incentive.BankKeeper,
	lk incentive.LeverageKeeper,
) Keeper {
	return Keeper{
		cdc:            cdc,
		storeKey:       storeKey,
		bankKeeper:     bk,
		leverageKeeper: lk,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+incentive.ModuleName)
}

// ModuleBalance returns the amount of a given token held in the x/incentive module account
func (k Keeper) ModuleBalance(ctx sdk.Context, denom string) sdk.Coin {
	amount := k.bankKeeper.SpendableCoins(ctx, authtypes.NewModuleAddress(incentive.ModuleName)).AmountOf(denom)
	return sdk.NewCoin(denom, amount)
}

// KVStore returns the module's KVStore
func (k Keeper) KVStore(ctx sdk.Context) sdk.KVStore {
	return ctx.KVStore(k.storeKey)
}

func (k Keeper) prefixStore(ctx sdk.Context, prefix []byte) sdk.KVStore {
	return prefixstore.NewStore(ctx.KVStore(k.storeKey), prefix)
}

func (k Keeper) setObject(ctx sdk.Context, key []byte, object codec.ProtoMarshaler, errField string) error {
	return store.SetObject(ctx.KVStore(k.storeKey), k.cdc, key, object, errField)
}
