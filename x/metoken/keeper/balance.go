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

func (k Keeper) setIndexBalances(balance metoken.IndexBalances) error {
	if err := balance.Validate(); err != nil {
		return err
	}

	return store.SetValue(k.store, keyBalance(balance.MetokenSupply.Denom), &balance, "balance")
}

func (k Keeper) hasIndexBalance(meTokenDenom string) bool {
	balance := store.GetValue[*metoken.IndexBalances](k.store, keyBalance(meTokenDenom), "balance")
	return balance != nil
}

// updateBalances of the assets of an Index and save them.
func (k Keeper) updateBalances(balances metoken.IndexBalances, updatedBalances []metoken.AssetBalance) error {
	if len(updatedBalances) > 0 {
		for _, balance := range updatedBalances {
			balances.SetAssetBalance(balance)
		}
		err := k.setIndexBalances(balances)
		if err != nil {
			return err
		}
	}

	return nil
}
