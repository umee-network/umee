package types

// DefaultGenesis returns the default genesis state of the x/leverage module.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: Params{
			InterestEpoch:                defaultInterestEpoch,
			CompleteLiquidationThreshold: defaultCompleteLiquidationThreshold,
			MinimumCloseFactor:           defaultMinimumCloseFactor,
		},
		Registry: []Token{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	for _, token := range gs.Registry {
		if err := token.Validate(); err != nil {
			return err
		}
	}

	return nil
}
