package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/x/leverage/types"
)

// GetLendAPY returns an sdk.Dec of an lend APY
// returns sdk.ZeroDec if not found
func (k Keeper) GetLendAPY(ctx sdk.Context, denom string) sdk.Dec {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateLendAPYKey(denom)

	bytesAPY := store.Get(key)
	if bytesAPY == nil {
		return sdk.ZeroDec()
	}

	var apy sdk.Dec
	err := apy.Unmarshal(bytesAPY)
	if err != nil {
		panic(err)
	}

	return apy
}

// SetLendAPY sets the lend APY of an specific denom
func (k Keeper) SetLendAPY(ctx sdk.Context, denom string, lendAPY sdk.Dec) error {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateLendAPYKey(denom)

	bytesAPY, err := lendAPY.Marshal()
	if err != nil {
		return err
	}

	store.Set(key, bytesAPY)
	return nil
}
