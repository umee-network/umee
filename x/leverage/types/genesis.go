package types

import fmt "fmt"

// DefaultGenesis returns the default genesis state of the x/leverage module.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: Params{
			InterestEpoch: 100,
		},
		Registry: []Token{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {

	if gs.Params.InterestEpoch <= 0 {
		return fmt.Errorf("interest epoch must be positive: %d", gs.Params.InterestEpoch)
	}

	for _, token := range gs.Registry {
		if err := token.Validate(); err != nil {
			return err
		}
	}

	return nil
}
