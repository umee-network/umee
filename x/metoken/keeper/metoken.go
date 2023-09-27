package keeper

import (
	"errors"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/umee-network/umee/v6/util/checkers"
	"github.com/umee-network/umee/v6/util/store"
	"github.com/umee-network/umee/v6/x/metoken"
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
	byEmergencyGroup bool,
) error {
	registry := k.GetAllRegisteredIndexes()

	registeredIndexes := make(map[string]metoken.Index)
	registeredAssets := make(map[string]string)
	for _, index := range registry {
		registeredIndexes[index.Denom] = index
		for _, aa := range index.AcceptedAssets {
			registeredAssets[aa.Denom] = index.Denom
		}
	}

	var errs []error
	if byEmergencyGroup {
		if len(addIndexes) > 0 {
			errs = append(errs, sdkerrors.ErrInvalidRequest.Wrapf("Emergency Group cannot register new indexes"))
		}

		errs = checkers.Merge(errs, validateEGIndexUpdate(updateIndexes, registeredIndexes))

		if len(errs) > 0 {
			return errors.Join(errs...)
		}
	}

	errs = checkers.Merge(errs, k.addIndexes(addIndexes, registeredIndexes, registeredAssets))

	errs = checkers.Merge(errs, k.updateIndexes(updateIndexes, registeredIndexes, registeredAssets))

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// addIndexes handles addition of the indexes from the request along with their validations.
func (k Keeper) addIndexes(
	indexes []metoken.Index,
	registeredIndexes map[string]metoken.Index,
	registeredAssets map[string]string,
) []error {
	var allErrs []error
	for _, index := range indexes {
		var indexErrs []error
		if _, present := registeredIndexes[index.Denom]; present {
			allErrs = append(
				allErrs, sdkerrors.ErrInvalidRequest.Wrapf(
					"add: index with denom %s already exists",
					index.Denom,
				),
			)
			continue
		}

		if exists := k.hasIndexBalance(index.Denom); exists {
			allErrs = append(
				allErrs, sdkerrors.ErrInvalidRequest.Wrapf(
					"can't add index %s - it already exists and is active",
					index.Denom,
				),
			)
		}

		for _, aa := range index.AcceptedAssets {
			if _, present := registeredAssets[aa.Denom]; present {
				indexErrs = append(
					indexErrs, sdkerrors.ErrInvalidRequest.Wrapf(
						"add: asset %s is already accepted in another index",
						aa.Denom,
					),
				)
			}
		}

		indexErrs = checkers.Merge(indexErrs, k.validateInLeverage(index))

		if len(indexErrs) > 0 {
			allErrs = append(allErrs, indexErrs...)
			continue
		}

		// adding index
		if err := k.setRegisteredIndex(index); err != nil {
			return []error{err}
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
			return []error{err}
		}
	}

	if len(allErrs) != 0 {
		return allErrs
	}

	return nil
}

// updateIndexes handles updates of the indexes from the request along with their validations.
func (k Keeper) updateIndexes(
	indexes []metoken.Index,
	registeredIndexes map[string]metoken.Index,
	registeredAssets map[string]string,
) []error {
	var allErrs []error
	for _, index := range indexes {
		var indexErrs []error
		oldIndex, present := registeredIndexes[index.Denom]
		if !present {
			allErrs = append(
				allErrs, sdkerrors.ErrNotFound.Wrapf(
					"update: index with denom %s not found",
					index.Denom,
				),
			)
			continue
		}

		for _, aa := range index.AcceptedAssets {
			if indexDenom, present := registeredAssets[aa.Denom]; present && indexDenom != index.Denom {
				indexErrs = append(
					indexErrs, sdkerrors.ErrInvalidRequest.Wrapf(
						"add: asset %s is already accepted in another index",
						aa.Denom,
					),
				)
			}
		}

		if oldIndex.Exponent != index.Exponent {
			balances, err := k.IndexBalances(index.Denom)
			if err != nil {
				return []error{err}
			}

			if balances.MetokenSupply.IsPositive() {
				indexErrs = append(
					indexErrs, sdkerrors.ErrInvalidRequest.Wrapf(
						"update: index %s exponent cannot be changed when supply is greater than zero",
						index.Denom,
					),
				)
			}
		}

		for _, aa := range oldIndex.AcceptedAssets {
			if exists := index.HasAcceptedAsset(aa.Denom); !exists {
				indexErrs = append(
					indexErrs, sdkerrors.ErrInvalidRequest.Wrapf(
						"update: an asset %s cannot be deleted from an index %s",
						aa.Denom,
						index.Denom,
					),
				)
			}
		}

		indexErrs = checkers.Merge(indexErrs, k.validateInLeverage(index))

		if len(indexErrs) > 0 {
			allErrs = append(allErrs, indexErrs...)
			continue
		}

		// updating balances if there is a new accepted asset
		if len(index.AcceptedAssets) > len(oldIndex.AcceptedAssets) {
			balances, err := k.IndexBalances(index.Denom)
			if err != nil {
				return []error{err}
			}

			for _, aa := range index.AcceptedAssets {
				if _, i := balances.AssetBalance(aa.Denom); i < 0 {
					balances.AssetBalances = append(balances.AssetBalances, metoken.NewZeroAssetBalance(aa.Denom))
				}
			}

			if err := k.setIndexBalances(balances); err != nil {
				return []error{err}
			}
		}

		if err := k.setRegisteredIndex(index); err != nil {
			return []error{err}
		}
	}

	if len(allErrs) != 0 {
		return allErrs
	}

	return nil
}

// validateEGIndexUpdate checks if emergency group can perform updates.
func validateEGIndexUpdate(indexes []metoken.Index, registeredIndexes map[string]metoken.Index) []error {
	var errs []error
	for _, newIndex := range indexes {
		d := newIndex.Denom
		oldIndex, ok := registeredIndexes[d]
		if !ok {
			errs = append(errs, sdkerrors.ErrNotFound.Wrapf("update: index with denom %s not found", d))
			continue
		}

		if newIndex.Exponent != oldIndex.Exponent {
			errs = append(errs, errors.New(d+": exponent cannot be changed"))
		}

		if !newIndex.Fee.Equal(oldIndex.Fee) {
			errs = append(errs, errors.New(d+": fee cannot be changed"))
		}

		if !newIndex.MaxSupply.Equal(oldIndex.MaxSupply) {
			errs = append(errs, errors.New(d+": max_supply cannot be changed"))
		}

		for _, newAsset := range newIndex.AcceptedAssets {
			oldAsset, i := oldIndex.AcceptedAsset(newAsset.Denom)
			if i < 0 {
				errs = append(errs, fmt.Errorf("%s: new asset %s cannot be added", d, newAsset.Denom))
				continue
			}

			if !newAsset.ReservePortion.Equal(oldAsset.ReservePortion) {
				errs = append(errs, fmt.Errorf("%s: reserve_portion of %s cannot be changed", d, newAsset.Denom))
			}
		}
	}

	return errs
}

// validateInLeverage validate the existence of every accepted asset in x/leverage
func (k Keeper) validateInLeverage(index metoken.Index) []error {
	var errs []error
	for _, aa := range index.AcceptedAssets {
		if _, err := k.leverageKeeper.GetTokenSettings(*k.ctx, aa.Denom); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func ModuleAddr() sdk.AccAddress {
	return authtypes.NewModuleAddress(metoken.ModuleName)
}
