package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/util/store"
	"github.com/umee-network/umee/v4/x/metoken"
)

// GetAllRegisteredIndexes returns all the registered Indexes from the x/metoken
// module's KVStore.
func (k Keeper) GetAllRegisteredIndexes(ctx sdk.Context) []metoken.Index {
	return store.MustLoadAll[*metoken.Index](k.KVStore(ctx), keyPrefixIndex)
}

// GetRegisteredIndex gets an Index from the x/metoken module's KVStore.
func (k Keeper) GetRegisteredIndex(ctx sdk.Context, meTokenDenom string) metoken.Index {
	index := metoken.Index{}
	ok := store.GetObject(k.KVStore(ctx), k.cdc, keyIndex(meTokenDenom), &index, "balance")

	if !ok {
		return metoken.Index{}
	}

	return index
}

// setRegisteredIndex saves a meToken Index with accepted assets and parameters
func (k Keeper) setRegisteredIndex(ctx sdk.Context, index metoken.Index) error {
	if err := index.Validate(); err != nil {
		return err
	}

	err := store.SetObject(k.KVStore(ctx), k.cdc, keyIndex(index.MetokenMaxSupply.Denom), &index, "index")

	//todo: hooks
	return err
}

// GetAllIndexesBalances returns asset balances of every Index
func (k Keeper) GetAllIndexesBalances(ctx sdk.Context) []metoken.IndexBalance {
	return store.MustLoadAll[*metoken.IndexBalance](k.KVStore(ctx), keyPrefixBalances)
}

// GetIndexBalance gets a Balance from the x/metoken module's KVStore.
func (k Keeper) GetIndexBalance(ctx sdk.Context, meTokenDenom string) metoken.IndexBalance {
	balance := metoken.IndexBalance{}
	ok := store.GetObject(k.KVStore(ctx), k.cdc, keyBalance(meTokenDenom), &balance, "balance")

	if !ok {
		return metoken.IndexBalance{}
	}

	return balance
}

// setIndexBalance saves an Index's Balance
func (k Keeper) setIndexBalance(ctx sdk.Context, balance metoken.IndexBalance) error {
	if err := balance.Validate(); err != nil {
		return err
	}

	err := store.SetObject(k.KVStore(ctx), k.cdc, keyBalance(balance.MetokenSupply.Denom), &balance, "balance")

	//todo: hooks
	return err
}

// getNextRebalancingTime returns next x/metoken re-balancing time
func (k Keeper) getNextRebalancingTime(ctx sdk.Context) int64 {
	return store.GetInt64(
		k.KVStore(ctx),
		keyPrefixNextRebalancingTime, "next re-balancing time",
	)
}

// setNextRebalancingTime next x/metoken re-balancing time
func (k Keeper) setNextRebalancingTime(ctx sdk.Context, nextRebalancingTime int64) error {
	return store.SetInt64(
		k.KVStore(ctx), keyPrefixNextRebalancingTime, nextRebalancingTime,
		"next re-balancing time",
	)
}

// getNextInterestClaimTime returns next x/metoken interest claiming time
func (k Keeper) getNextInterestClaimTime(ctx sdk.Context) int64 {
	return store.GetInt64(
		k.KVStore(ctx),
		keyPrefixNextInterestClaimTime, "next interest claim time",
	)
}

// setNextInterestClaimTime next x/metoken interest claiming time
func (k Keeper) setNextInterestClaimTime(ctx sdk.Context, nextInterestClaimTime int64) error {
	return store.SetInt64(
		k.KVStore(ctx), keyPrefixNextInterestClaimTime, nextInterestClaimTime,
		"next interest claim time",
	)
}
