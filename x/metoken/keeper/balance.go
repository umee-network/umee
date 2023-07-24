package keeper

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/umee-network/umee/v5/util/store"
	"github.com/umee-network/umee/v5/x/metoken"
)

// IndexBalances returns Index Token supply, if it's not found returns an error.
func (k Keeper) IndexBalances(meTokenDenom string) (metoken.IndexBalances, error) {
	balance := store.GetValue[*metoken.IndexBalances](k.store, keyBalance(meTokenDenom), "balance")
	if balance == nil {
		return metoken.IndexBalances{}, sdkerrors.ErrNotFound.Wrapf("balance for index %s not found", meTokenDenom)
	}

	return *balance, nil
}

// setIndexBalances saves an Index's Balance
func (k Keeper) setIndexBalances(balance metoken.IndexBalances) error {
	if err := balance.Validate(); err != nil {
		return err
	}

	return store.SetValue(k.store, keyBalance(balance.MetokenSupply.Denom), &balance, "balance")
}

// hasIndexBalance returns true when Index exists.
func (k Keeper) hasIndexBalance(meTokenDenom string) bool {
	balance := store.GetValue[*metoken.IndexBalances](k.store, keyBalance(meTokenDenom), "balance")
	return balance != nil
}
