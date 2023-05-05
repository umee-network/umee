package metoken

import (
	"fmt"
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// MeTokenPrefix defines the meToken denomination prefix for all meToken Indexes.
	MeTokenPrefix = "me"
)

// HasMeTokenPrefix detects the meToken prefix on a denom.
func IsMeToken(denom string) bool {
	return strings.HasPrefix(denom, MeTokenPrefix)
}

// NewIndex creates a new Index object
func NewIndex(maxSupply sdk.Coin, fee Fee, acceptedAssets []AcceptedAsset) Index {
	return Index{
		MetokenMaxSupply: maxSupply,
		Fee:              fee,
		AcceptedAssets:   acceptedAssets,
	}
}

// Validate perform basic validation of the Index
func (i Index) Validate() error {
	if err := i.MetokenMaxSupply.Validate(); err != nil {
		return err
	}

	if !IsMeToken(i.MetokenMaxSupply.Denom) {
		return sdkerrors.ErrInvalidRequest.Wrapf(
			"meToken denom %s should have the following format: me<TokenName>",
			i.MetokenMaxSupply.Denom,
		)
	}

	if err := i.Fee.Validate(); err != nil {
		return err
	}

	totalAllocation := sdk.ZeroDec()
	existingAssets := make(map[string]struct{})
	for _, asset := range i.AcceptedAssets {
		if _, present := existingAssets[asset.AssetDenom]; present {
			return fmt.Errorf("duplicated accepted asset in the Index: %s", asset.AssetDenom)
		}
		existingAssets[asset.AssetDenom] = struct{}{}

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

// NewAcceptedAsset creates a new AcceptedAsset object
func NewAcceptedAsset(assetDenom string, reservePortion, targetAllocation sdk.Dec) AcceptedAsset {
	return AcceptedAsset{
		AssetDenom:       assetDenom,
		ReservePortion:   reservePortion,
		TargetAllocation: targetAllocation,
	}
}

// Validate perform basic validation of the AcceptedAsset
func (aa AcceptedAsset) Validate() error {
	if err := sdk.ValidateDenom(aa.AssetDenom); err != nil {
		return err
	}

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
