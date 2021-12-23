package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/x/leverage/types"
)

// GetCollateralAmount returns an sdk.Coin representing how much of a given denom the
// x/leverage module account currently holds as collateral for a given borrower.
func (k Keeper) GetCollateralAmount(ctx sdk.Context, borrowerAddr sdk.AccAddress, denom string) sdk.Coin {
	store := ctx.KVStore(k.storeKey)
	collateral := sdk.NewCoin(denom, sdk.ZeroInt())
	key := types.CreateCollateralAmountKey(borrowerAddr, denom)

	if bz := store.Get(key); bz != nil {
		err := collateral.Amount.Unmarshal(bz)
		if err != nil {
			panic(err)
		}
	}

	return collateral
}

// SetCollateralAmount sets the amount of a given denom the x/leverage module account
// currently holds as collateral for a given borrower. If the amount is zero, any
// stored value is cleared. A negative amount or invalid coin causes an error.
// This function does not move coins to or from the module account.
func (k Keeper) SetCollateralAmount(ctx sdk.Context, borrowerAddr sdk.AccAddress, collateral sdk.Coin) error {
	if !collateral.IsValid() {
		return sdkerrors.Wrap(types.ErrInvalidAsset, collateral.String())
	}

	bz, err := collateral.Amount.Marshal()
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	key := types.CreateCollateralAmountKey(borrowerAddr, collateral.Denom)

	if collateral.Amount.IsZero() {
		store.Delete(key)
	} else {
		store.Set(key, bz)
	}
	return nil
}

// GetBorrowerCollateral returns an sdk.Coins containing all of a borrower's collateral.
func (k Keeper) GetBorrowerCollateral(ctx sdk.Context, borrowerAddr sdk.AccAddress) sdk.Coins {
	store := ctx.KVStore(k.storeKey)
	prefix := types.CreateCollateralAmountKeyNoDenom(borrowerAddr)

	totalCollateral := sdk.NewCoins()

	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		// key is prefix | denom | 0x00
		key, val := iter.Key(), iter.Value()
		denom := string(key[len(prefix) : len(key)-1]) // remove prefix and null-terminator

		var amount sdk.Int
		if err := amount.Unmarshal(val); err != nil {
			// improperly marshaled collateral amount should never happen
			panic(err)
		}

		// for each collateral coin found, add it to totalCollateral
		totalCollateral = totalCollateral.Add(sdk.NewCoin(denom, amount))
	}

	totalCollateral.Sort()
	return totalCollateral
}

// GetEligibleLiquidationTargets returns a list of borrower addresses eligible for liquidation
func (k Keeper) GetEligibleLiquidationTargets(ctx sdk.Context) ([]sdk.AccAddress, error) {
	// TODO: Complete this to support liquidation tool
	// ref: https://github.com/umee-network/umee/issues/229
	return nil, nil
}
