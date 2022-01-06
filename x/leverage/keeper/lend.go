package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/x/leverage/types"
)

// GetLendAPY returns an sdk.Dec of an lend APY returns sdk.ZeroDec if not found.
func (k Keeper) GetLendAPY(ctx sdk.Context, denom string) sdk.Dec {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateLendAPYKey(denom)

	bz := store.Get(key)
	if bz == nil {
		return sdk.ZeroDec()
	}

	var lendAPY sdk.Dec
	if err := lendAPY.Unmarshal(bz); err != nil {
		panic(err)
	}

	return lendAPY
}

// SetLendAPY sets the lend APY of an specific denom.
func (k Keeper) SetLendAPY(ctx sdk.Context, denom string, lendAPY sdk.Dec) error {
	if !k.IsAcceptedToken(ctx, denom) {
		return sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}

	bz, err := lendAPY.Marshal()
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	key := types.CreateLendAPYKey(denom)
	store.Set(key, bz)
	return nil
}
