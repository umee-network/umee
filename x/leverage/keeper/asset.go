package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/x/leverage/types"
)

// SetAsset stores an asset into the x/leverage module's KVStore.
func (k Keeper) SetAsset(ctx sdk.Context, asset types.Asset) {
	store := ctx.KVStore(k.storeKey)
	assetKey := types.CreateAssetKey(asset.BaseTokenDenom)

	bz, err := asset.Marshal()
	if err != nil {
		panic(fmt.Sprintf("failed to encode asset: %s", err))
	}

	store.Set(assetKey, bz)
}
