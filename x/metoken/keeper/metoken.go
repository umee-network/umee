package keeper

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/umee-network/umee/v5/util/store"
	"github.com/umee-network/umee/v5/x/metoken"
)

// GetAllRegisteredIndexes returns all the registered Indexes from the x/metoken
// module's KVStore.
func (k Keeper) GetAllRegisteredIndexes() []metoken.Index {
	return store.MustLoadAll[*metoken.Index](k.store, keyPrefixIndex)
}

// RegisteredIndex gets an Index from the x/metoken module's KVStore, if not found returns an error.
func (k Keeper) RegisteredIndex(meTokenDenom string) (metoken.Index, error) {
	index := store.GetValue[*metoken.Index](k.store, keyIndex(meTokenDenom), "index")

	if index == nil {
		return metoken.Index{}, sdkerrors.ErrNotFound.Wrapf("index %s not found", meTokenDenom)
	}

	return *index, nil
}

// setRegisteredIndex saves a meToken Index with accepted assets and parameters
func (k Keeper) setRegisteredIndex(index metoken.Index) error {
	if err := index.Validate(); err != nil {
		return err
	}

	return store.SetValue(k.store, keyIndex(index.Denom), &index, "index")
}

// GetAllIndexesBalances returns asset balances of every Index
func (k Keeper) GetAllIndexesBalances() []metoken.IndexBalances {
	return store.MustLoadAll[*metoken.IndexBalances](k.store, keyPrefixBalances)
}

// getNextRebalancingTime returns next x/metoken re-balancing time in Milliseconds.
// Returns 0 unix time if the time was not set before.
func (k Keeper) getNextRebalancingTime() time.Time {
	t, _ := store.GetTimeMs(k.store, keyPrefixNextRebalancingTime)
	return t
}

// setNextRebalancingTime next x/metoken re-balancing time in Milliseconds.
func (k Keeper) setNextRebalancingTime(nextRebalancingTime time.Time) {
	store.SetTimeMs(k.store, keyPrefixNextRebalancingTime, nextRebalancingTime)
}

// getNextInterestClaimTime returns next x/metoken interest claiming time in Milliseconds.
// Returns 0 unix time if the time was not set before.
func (k Keeper) getNextInterestClaimTime() time.Time {
	t, _ := store.GetTimeMs(k.store, keyPrefixNextInterestClaimTime)
	return t
}

// setNextInterestClaimTime next x/metoken interest claiming time in Milliseconds.
func (k Keeper) setNextInterestClaimTime(nextInterestClaimTime time.Time) {
	store.SetTimeMs(k.store, keyPrefixNextInterestClaimTime, nextInterestClaimTime)
}

// UpdateIndexes registers `addIndexes` and processes `updateIndexes` to update existing indexes.
func (k Keeper) UpdateIndexes(
	addIndexes []metoken.Index,
	updateIndexes []metoken.Index,
) error {
	registry := k.GetAllRegisteredIndexes()

	registeredIndexes := make(map[string]metoken.Index)
	for _, index := range registry {
		registeredIndexes[index.Denom] = index
	}

	if err := k.addIndexes(addIndexes, registeredIndexes); err != nil {
		return err
	}

	return k.updateIndexes(updateIndexes, registeredIndexes)
}

// addIndexes handles addition of the indexes from the request along with their validations.
func (k Keeper) addIndexes(indexes []metoken.Index, registeredIndexes map[string]metoken.Index) error {
	for _, index := range indexes {
		if _, present := registeredIndexes[index.Denom]; present {
			return sdkerrors.ErrInvalidRequest.Wrapf(
				"add: index with denom %s already exists",
				index.Denom,
			)
		}

		if err := k.validateInLeverage(index); err != nil {
			return err
		}

		if exists := k.hasIndexBalance(index.Denom); exists {
			return sdkerrors.ErrInvalidRequest.Wrapf(
				"can't add index %s - it already exists and is active",
				index.Denom,
			)
		}

		// adding index
		if err := k.setRegisteredIndex(index); err != nil {
			return err
		}

		assetBalances := make([]metoken.AssetBalance, 0)
		for _, aa := range index.AcceptedAssets {
			assetBalances = append(assetBalances, metoken.NewZeroAssetBalance(aa.Denom))
		}

		// adding initial balances for the index
		if err := k.setIndexBalances(
			metoken.NewIndexBalances(
				sdk.NewCoin(
					index.Denom,
					sdkmath.ZeroInt(),
				), assetBalances,
			),
		); err != nil {
			return err
		}
	}

	return nil
}

// updateIndexes handles updates of the indexes from the request along with their validations.
func (k Keeper) updateIndexes(indexes []metoken.Index, registeredIndexes map[string]metoken.Index) error {
	for _, index := range indexes {
		oldIndex, present := registeredIndexes[index.Denom]
		if !present {
			return sdkerrors.ErrNotFound.Wrapf(
				"update: index with denom %s not found",
				index.Denom,
			)
		}

		if oldIndex.Exponent != index.Exponent {
			balances, err := k.IndexBalances(index.Denom)
			if err != nil {
				return err
			}

			if balances.MetokenSupply.IsPositive() {
				return sdkerrors.ErrInvalidRequest.Wrapf(
					"update: index %s exponent cannot be changed when supply is greater than zero",
					index.Denom,
				)
			}
		}

		for _, aa := range oldIndex.AcceptedAssets {
			if exists := index.HasAcceptedAsset(aa.Denom); !exists {
				return sdkerrors.ErrInvalidRequest.Wrapf(
					"update: an asset %s cannot be deleted from an index %s",
					aa.Denom,
					index.Denom,
				)
			}
		}

		if err := k.validateInLeverage(index); err != nil {
			return err
		}

		// updating balances if there is a new accepted asset
		if len(index.AcceptedAssets) > len(oldIndex.AcceptedAssets) {
			balances, err := k.IndexBalances(index.Denom)
			if err != nil {
				return err
			}

			for _, aa := range index.AcceptedAssets {
				if i, _ := balances.AssetBalance(aa.Denom); i < 0 {
					balances.AssetBalances = append(balances.AssetBalances, metoken.NewZeroAssetBalance(aa.Denom))
				}
			}

			if err := k.setIndexBalances(balances); err != nil {
				return err
			}
		}

		if err := k.setRegisteredIndex(index); err != nil {
			return err
		}
	}

	return nil
}

// validateInLeverage validate the existence of every accepted asset in x/leverage
func (k Keeper) validateInLeverage(index metoken.Index) error {
	for _, aa := range index.AcceptedAssets {
		if _, err := k.leverageKeeper.GetTokenSettings(*k.ctx, aa.Denom); err != nil {
			return err
		}
	}

	return nil
}

func ModuleAddr() sdk.AccAddress {
	return authtypes.NewModuleAddress(metoken.ModuleName)
}
