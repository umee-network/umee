package uibc

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewGenesisState(params Params, quotas []Quota, outflowSum sdk.Dec) *GenesisState {
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
		if err := quota.Validate(); err != nil {
			return err
		}
	}

	if gs.TotalOutflowSum.IsNegative() {
		return fmt.Errorf("total outflow sum shouldn't be negative : %s ", gs.TotalOutflowSum.String())
	}

	return nil
}
