package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/umee-network/umee/v3/x/incentive/types"
)

type Keeper struct {
	cdc            codec.Codec
	storeKey       storetypes.StoreKey
	bankKeeper     types.BankKeeper
	leverageKeeper types.LeverageKeeper
}

func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	bk types.BankKeeper,
	lk types.LeverageKeeper,
) Keeper {
	return Keeper{
		cdc:            cdc,
		storeKey:       storeKey,
		bankKeeper:     bk,
		leverageKeeper: lk,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// ModuleBalance returns the amount of a given token held in the x/incentive module account
func (k Keeper) ModuleBalance(ctx sdk.Context, denom string) sdk.Int {
	return k.bankKeeper.SpendableCoins(ctx, authtypes.NewModuleAddress(types.ModuleName)).AmountOf(denom)
}

// Claim claims any pending rewards belonging to an address.
func (k Keeper) Claim(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coins, error) {
	return sdk.Coins{}, types.ErrNotImplemented
}
