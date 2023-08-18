package metoken

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewGenesisState creates a new GenesisState object
func NewGenesisState(
	params Params,
	registry []Index,
	balances []IndexBalances,
	nextRebalancingTime time.Time,
	nextInterestClaimTime time.Time,
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
		NextRebalancingTime:   time.Time{},
		NextInterestClaimTime: time.Time{},
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

// NewIndexBalances creates a new IndexBalances object.
func NewIndexBalances(meTokenSupply sdk.Coin, assetBalances []AssetBalance) IndexBalances {
	return IndexBalances{
		MetokenSupply: meTokenSupply,
		AssetBalances: assetBalances,
	}
}

// Validate perform basic validation of the IndexBalances
func (ib IndexBalances) Validate() error {
	if !IsMeToken(ib.MetokenSupply.Denom) {
		return sdkerrors.ErrInvalidRequest.Wrapf(
			"meToken denom %s should have the following format: me/<TokenName>",
			ib.MetokenSupply.Denom,
		)
	}

	if err := ib.MetokenSupply.Validate(); err != nil {
		return err
	}

	existingBalances := make(map[string]struct{})
	for _, balance := range ib.AssetBalances {
		if _, present := existingBalances[balance.Denom]; present {
			return fmt.Errorf("duplicated balance %s in the Index: %s", balance.Denom, ib.MetokenSupply.Denom)
		}
		existingBalances[balance.Denom] = struct{}{}

		if err := balance.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// AssetBalance returns an asset balance and its index from balances, given a specific denom.
// If it isn't present, -1 as index.
func (ib IndexBalances) AssetBalance(denom string) (AssetBalance, int) {
	for i, balance := range ib.AssetBalances {
		if balance.Denom == denom {
			return balance, i
		}
	}

	return AssetBalance{}, -1
}

// SetAssetBalance overrides an asset balance if exists in the list, otherwise add it to the list.
func (ib *IndexBalances) SetAssetBalance(balance AssetBalance) {
	_, i := ib.AssetBalance(balance.Denom)

	if i < 0 {
		ib.AssetBalances = append(ib.AssetBalances, balance)
		return
	}

	ib.AssetBalances[i] = balance
}

// NewZeroAssetBalance creates a new AssetBalance object with all balances in zero.
func NewZeroAssetBalance(denom string) AssetBalance {
	return AssetBalance{
		Denom:     denom,
		Leveraged: sdkmath.ZeroInt(),
		Reserved:  sdkmath.ZeroInt(),
		Fees:      sdkmath.ZeroInt(),
		Interest:  sdkmath.ZeroInt(),
	}
}

// NewAssetBalance creates a new AssetBalance object.
func NewAssetBalance(denom string, inLeverage, inReserves, inFees, inInterest sdkmath.Int) AssetBalance {
	return AssetBalance{
		Denom:     denom,
		Leveraged: inLeverage,
		Reserved:  inReserves,
		Fees:      inFees,
		Interest:  inInterest,
	}
}

// Validate perform basic validation of the AssetBalance
func (ab AssetBalance) Validate() error {
	if err := sdk.ValidateDenom(ab.Denom); err != nil {
		return err
	}
	if ab.Leveraged.IsNegative() {
		return sdkerrors.ErrInvalidRequest.Wrapf("leveraged asset balance cannot be negative")
	}
	if ab.Reserved.IsNegative() {
		return sdkerrors.ErrInvalidRequest.Wrapf("reserved asset balance cannot be negative")
	}
	if ab.Fees.IsNegative() {
		return sdkerrors.ErrInvalidRequest.Wrapf("fees asset balance cannot be negative")
	}
	if ab.Interest.IsNegative() {
		return sdkerrors.ErrInvalidRequest.Wrapf("interest asset balance cannot be negative")
	}

	return nil
}

// AvailableSupply returns reserved plus leveraged
func (ab AssetBalance) AvailableSupply() sdkmath.Int {
	return ab.Reserved.Add(ab.Leveraged)
}
