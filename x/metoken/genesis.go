package metoken

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState creates a new GenesisState object
func NewGenesisState(
	params Params,
	registry []Index,
	balances []IndexBalance,
	nextRebalancingTime int64,
	nextInterestClaimTime int64,
) *GenesisState {
	return &GenesisState{
		Params:                params,
		Registry:              registry,
		Balances:              balances,
		NextRebalancingTime:   nextRebalancingTime,
		NextInterestClaimTime: nextInterestClaimTime,
	}
}

// DefaultGenesisState creates a new default GenesisState object
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:                DefaultParams(),
		Registry:              nil,
		Balances:              nil,
		NextRebalancingTime:   0,
		NextInterestClaimTime: 0,
	}
}

// Validate perform basic validation of the GenesisState
func (gs GenesisState) Validate() error {
	for _, index := range gs.Registry {
		if err := index.Validate(); err != nil {
			return err
		}
	}

	for _, balance := range gs.Balances {
		if err := balance.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// NewIndexBalance creates a new IndexBalance object.
func NewIndexBalance(meTokenSupply sdk.Coin, assetBalances []AssetBalance) IndexBalance {
	return IndexBalance{
		MetokenSupply: meTokenSupply,
		AssetBalances: assetBalances,
	}
}

// Validate perform basic validation of the IndexBalance
func (ib IndexBalance) Validate() error {
	if !IsMeToken(ib.MetokenSupply.Denom) {
		return sdkerrors.ErrInvalidRequest.Wrapf(
			"meToken denom %s should have the following format: me<TokenName>",
			ib.MetokenSupply.Denom,
		)
	}

	if err := ib.MetokenSupply.Validate(); err != nil {
		return err
	}

	for _, balance := range ib.AssetBalances {
		if err := balance.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// NewAssetBalance creates a new AssetBalance object
func NewAssetBalance(inLeverage, inReserves, inFees sdk.Coin) AssetBalance {
	return AssetBalance{
		Leveraged: inLeverage,
		Reserved:  inReserves,
		Fees:      inFees,
	}
}

// Validate perform basic validation of the AssetBalance
func (ab AssetBalance) Validate() error {
	if err := ab.Leveraged.Validate(); err != nil {
		return err
	}

	if err := ab.Reserved.Validate(); err != nil {
		return err
	}

	if err := ab.Fees.Validate(); err != nil {
		return err
	}

	if ab.Leveraged.Denom != ab.Reserved.Denom ||
		ab.Leveraged.Denom != ab.Fees.Denom {
		return fmt.Errorf(
			"different assets in the Index balance: %s, %s, %s",
			ab.Leveraged.Denom,
			ab.Reserved.Denom,
			ab.Fees.Denom,
		)
	}

	return nil
}
