package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/x/leverage/types"
)

// iterate through all keys with a given prefix using a provided function.
// If the provided function returns an error, iteration stops and the error
// is returned.
func (k Keeper) iterate(ctx sdk.Context, prefix []byte, cb func(key, val []byte) error) error {
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key, val := iter.Key(), iter.Value()

		if err := cb(key, val); err != nil {
			return err
		}
	}

	return nil
}

// GetAllBadDebts gets bad debt instances across all borrowers.
func (k Keeper) GetAllBadDebts(ctx sdk.Context) []types.BadDebt {
	prefix := types.KeyPrefixBadDebt
	badDebts := []types.BadDebt{}

	iterator := func(key, val []byte) error {
		addr := types.AddressFromKey(key, prefix)
		denom := types.DenomFromKeyWithAddress(key, prefix)

		badDebts = append(badDebts, types.NewBadDebt(addr.String(), denom))

		return nil
	}

	err := k.iterate(ctx, prefix, iterator)
	if err != nil {
		panic(err)
	}

	return badDebts
}
