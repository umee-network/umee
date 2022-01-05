package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/x/leverage/types"
)

// GetLendApy returns an sdk.Dec of an lend APY
// returns sdk.ZeroDec if not found
func (k Keeper) GetLendApy(ctx sdk.Context, denom string) sdk.Dec {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateLendAPYKey(denom)

	apyBytes := store.Get(key)
	if apyBytes == nil {
		return sdk.ZeroDec()
	}

	var apy sdk.Dec
	err := apy.Unmarshal(apyBytes)
	if err != nil {
		panic(err)
	}

	return apy
}

// SetLendApy sets the lend APY of an specific denom
func (k Keeper) SetLendApy(ctx sdk.Context, denom string, apy sdk.Dec) error {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateLendAPYKey(denom)

	apyBytes, err := apy.Marshal()
	if err != nil {
		return err
	}

	store.Set(key, apyBytes)
	return nil
}
