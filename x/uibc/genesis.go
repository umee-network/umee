package uibc

import (
	fmt "fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewGenesisState(params Params, outflows sdk.DecCoins, outflowSum sdkmath.LegacyDec) *GenesisState {
	return &GenesisState{
		Params:     params,
		Outflows:   outflows,
		OutflowSum: outflowSum,
	}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:     DefaultParams(),
		Inflows:    nil,
		Outflows:   nil,
		OutflowSum: sdkmath.LegacyNewDec(0),
		InflowSum:  sdkmath.LegacyNewDec(0),
	}
}

// Validate performs basic valida`tion of the interchain accounts GenesisState
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	for _, o := range gs.Outflows {
		if o.Amount.IsNil() {
			return sdkerrors.ErrInvalidRequest.Wrap("ibc denom outflow must be defined")
		}
		if err := o.Validate(); err != nil {
			return err
		}
	}

	for _, o := range gs.Inflows {
		if o.Amount.IsNil() {
			return sdkerrors.ErrInvalidRequest.Wrap("ibc denom inflow must be defined")
		}
		if err := o.Validate(); err != nil {
			return err
		}
	}

	if gs.OutflowSum.IsNegative() {
		return fmt.Errorf("outflow sum cannot be negative : %s ", gs.OutflowSum.String())
	}

	if gs.InflowSum.IsNegative() {
		return fmt.Errorf("inflow sum cannot be negative : %s ", gs.InflowSum.String())
	}

	return nil
}
