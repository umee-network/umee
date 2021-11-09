package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/umee-network/umee/x/leverage/types"
)

// GetReserveAmount gets the amount reserved of a specified token. On invalid asset, the reserved amount is zero.
func (k Keeper) GetReserveAmount(ctx sdk.Context, denom string) sdk.Int {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateReserveAmountKey(denom)
	amount := sdk.ZeroInt()
	bz := store.Get(key)
	if bz != nil {
		err := amount.Unmarshal(bz)
		if err != nil {
			panic(err)
		}
	}
	if amount.IsNegative() {
		panic("negative reserve amount detected")
	}
	return amount
}

// IncreaseReserves adds an sdk.Coin (denom, amount) to the module's reserve requirements.
func (k Keeper) IncreaseReserves(ctx sdk.Context, coin sdk.Coin) error {
	store := ctx.KVStore(k.storeKey)
	if !k.IsAcceptedToken(ctx, coin.Denom) {
		return sdkerrors.Wrap(types.ErrInvalidAsset, coin.String())
	}

	// Get current amount reserved for this asset type
	reserveKey := types.CreateReserveAmountKey(coin.Denom)
	currentReserve := sdk.ZeroInt()
	bz := store.Get(reserveKey)
	if bz != nil {
		err := currentReserve.Unmarshal(bz)
		if err != nil {
			return err
		}
	}

	// Add the new reserve amount to the current one and save
	bz, err := currentReserve.Add(coin.Amount).Marshal()
	if err != nil {
		return err
	}
	store.Set(reserveKey, bz)
	return nil
}
