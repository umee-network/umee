package uibc

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewGenesisState(params Params, quotas sdk.DecCoins, outflowSum sdk.Dec) *GenesisState {
	return &GenesisState{
		Params:          params,
		Quotas:          quotas,
		TotalOutflowSum: outflowSum,
	}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:          DefaultParams(),
		Quotas:          nil,
		TotalOutflowSum: sdk.NewDec(0),
	}
}

// Validate performs basic valida`tion of the interchain accounts GenesisState
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	for _, quota := range gs.Quotas {
		if quota.Amount.IsNil() {
			return sdkerrors.ErrInvalidRequest.Wrap("ibc denom quota must be defined")
		}
		if err := quota.Validate(); err != nil {
			return err
		}
	}

	if gs.TotalOutflowSum.IsNegative() {
		return fmt.Errorf("total outflow sum cannot be negative : %s ", gs.TotalOutflowSum.String())
	}

	return nil
}
