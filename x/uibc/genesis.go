package uibc

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewGenesisState(params Params, outflows sdk.DecCoins, outflowSum sdk.Dec) *GenesisState {
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
		OutflowSum: sdk.NewDec(0),
		InflowSum:  sdk.NewDec(0),
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
		return fmt.Errorf("total outflow sum cannot be negative : %s ", gs.OutflowSum.String())
	}

	if gs.InflowSum.IsNegative() {
		return fmt.Errorf("total inflow sum cannot be negative : %s ", gs.InflowSum.String())
	}

	return nil
}
