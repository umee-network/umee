package ibctransfer

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewGenesisState(params Params, rateLimits []RateLimit, outflowSum sdk.Dec) *GenesisState {
	return &GenesisState{
		Params:          params,
		RateLimits:      rateLimits,
		TotalOutflowSum: outflowSum,
	}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:          *DefaultParams(),
		RateLimits:      nil,
		TotalOutflowSum: sdk.NewDec(0),
	}
}

// Validate performs basic valida`tion of the interchain accounts GenesisState
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	for _, rateLimits := range gs.RateLimits {
		if err := rateLimits.Validate(); err != nil {
			return err
		}
	}

	if gs.TotalOutflowSum.IsNegative() {
		return fmt.Errorf("total outflow sum shouldn't be negative : %s ", gs.TotalOutflowSum.String())
	}

	return nil
}
