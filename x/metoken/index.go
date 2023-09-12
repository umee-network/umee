package metoken

import (
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// MeTokenPrefix defines the meToken denomination prefix for all meToken Indexes.
	MeTokenPrefix = "me/"
)

// IsMeToken detects the meToken prefix on a denom.
func IsMeToken(denom string) bool {
	return strings.HasPrefix(denom, MeTokenPrefix)
}

// NewIndex creates a new Index object
func NewIndex(denom string, maxSupply sdkmath.Int, exponent uint32, fee Fee, acceptedAssets []AcceptedAsset) Index {
	return Index{
		Denom:          denom,
		MaxSupply:      maxSupply,
		Exponent:       exponent,
		Fee:            fee,
		AcceptedAssets: acceptedAssets,
	}
}

// Validate perform basic validation of the Index
func (i Index) Validate() error {
	if !IsMeToken(i.Denom) {
		return sdkerrors.ErrInvalidRequest.Wrapf(
			"meToken denom %s should have the following format: me/<TokenName>",
			i.Denom,
		)
	}

	if i.MaxSupply.IsNegative() {
		return sdkerrors.ErrInvalidRequest.Wrapf(
			"maxSupply cannot be negative for %s",
			i.Denom,
		)
	}

	if err := i.Fee.Validate(); err != nil {
		return err
	}

	totalAllocation := sdk.ZeroDec()
	existingAssets := make(map[string]struct{})
	for _, asset := range i.AcceptedAssets {
		if _, present := existingAssets[asset.Denom]; present {
			return fmt.Errorf("duplicated accepted asset %s in the Index: %s", asset.Denom, i.Denom)
		}
		existingAssets[asset.Denom] = struct{}{}

		if err := sdk.ValidateDenom(asset.Denom); err != nil {
			return err
		}

		if err := asset.Validate(); err != nil {
			return err
		}
		totalAllocation = totalAllocation.Add(asset.TargetAllocation)
	}

	if !totalAllocation.Equal(sdk.OneDec()) {
		return sdkerrors.ErrInvalidRequest.Wrapf(
			"total allocation %s of all the accepted assets should be 1.0",
			totalAllocation.String(),
		)
	}

	return nil
}

// NewFee creates a new Fee object
func NewFee(minFee, balancedFee, maxFee sdk.Dec) Fee {
	return Fee{
		MinFee:      minFee,
		BalancedFee: balancedFee,
		MaxFee:      maxFee,
	}
}

// Validate perform basic validation of the Fee
func (f Fee) Validate() error {
	if f.MinFee.IsNegative() || f.MinFee.GT(sdk.OneDec()) {
		return sdkerrors.ErrInvalidRequest.Wrapf("min_fee %s should be between 0.0 and 1.0", f.MinFee.String())
	}

	if f.BalancedFee.IsNegative() || f.BalancedFee.GT(sdk.OneDec()) {
		return sdkerrors.ErrInvalidRequest.Wrapf(
			"balanced_fee %s should be between 0.0 and 1.0",
			f.BalancedFee.String(),
		)
	}

	// BalancedFee must be always greater than MinFee for correct incentivizing and disincentivizing the allocation
	// of every asset in the index
	if f.BalancedFee.LTE(f.MinFee) {
		return sdkerrors.ErrInvalidRequest.Wrapf(
			"balanced_fee %s should be greater than min_fee %s",
			f.BalancedFee.String(), f.MinFee.String(),
		)
	}

	if f.MaxFee.IsNegative() || f.MaxFee.GT(sdk.OneDec()) {
		return sdkerrors.ErrInvalidRequest.Wrapf("max_fee %s should be between 0.0 and 1.0", f.MaxFee.String())
	}

	// MaxFee must be always greater than BalancedFee for correct incentivizing and disincentivizing the allocation
	// of every asset in the index
	if f.MaxFee.LTE(f.BalancedFee) {
		return sdkerrors.ErrInvalidRequest.Wrapf(
			"max_fee %s should be greater than balanced_fee %s",
			f.MaxFee.String(), f.BalancedFee.String(),
		)
	}

	return nil
}

// CalculateFee based on its settings and allocation deviation.
func (f Fee) CalculateFee(allocationDeviation sdk.Dec) sdk.Dec {
	fee := allocationDeviation.Mul(f.BalancedFee).Add(f.BalancedFee)

	if fee.LT(f.MinFee) {
		return f.MinFee
	}

	if fee.GT(f.MaxFee) {
		return f.MaxFee
	}

	return fee
}

// NewAcceptedAsset creates a new AcceptedAsset object
func NewAcceptedAsset(denom string, reservePortion, targetAllocation sdk.Dec) AcceptedAsset {
	return AcceptedAsset{
		Denom:            denom,
		ReservePortion:   reservePortion,
		TargetAllocation: targetAllocation,
	}
}

// Validate perform basic validation of the AcceptedAsset
func (aa AcceptedAsset) Validate() error {
	if aa.TargetAllocation.IsNegative() || aa.TargetAllocation.GT(sdk.OneDec()) {
		return sdkerrors.ErrInvalidRequest.Wrapf(
			"target_allocation %s should be between 0.0 and 1.0",
			aa.TargetAllocation.String(),
		)
	}

	if aa.ReservePortion.IsNegative() || aa.ReservePortion.GT(sdk.OneDec()) {
		return sdkerrors.ErrInvalidRequest.Wrapf(
			"reserve_portion %s should be between 0.0 and 1.0",
			aa.ReservePortion.String(),
		)
	}

	return nil
}

// AcceptedAsset returns an accepted asset and its index in the list, given a specific denom. If it isn't present,
// returns -1.
func (i Index) AcceptedAsset(denom string) (AcceptedAsset, int) {
	for index, aa := range i.AcceptedAssets {
		if aa.Denom == denom {
			return aa, index
		}
	}
	return AcceptedAsset{}, -1
}

// HasAcceptedAsset returns true if an accepted asset is present in the index. Otherwise returns false.
func (i Index) HasAcceptedAsset(denom string) bool {
	for _, aa := range i.AcceptedAssets {
		if aa.Denom == denom {
			return true
		}
	}
	return false
}

// SetAcceptedAsset overrides an accepted asset if exists in the list, otherwise add it to the list.
func (i *Index) SetAcceptedAsset(acceptedAsset AcceptedAsset) {
	_, index := i.AcceptedAsset(acceptedAsset.Denom)
	if index < 0 {
		i.AcceptedAssets = append(i.AcceptedAssets, acceptedAsset)
		return
	}
	i.AcceptedAssets[index] = acceptedAsset
}
