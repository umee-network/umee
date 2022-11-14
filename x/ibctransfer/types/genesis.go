package types

func NewGenesisState(params Params) *GenesisState {
	return &GenesisState{
		Params: params,
	}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: *DefaultParams(),
	}
}

// Validate performs basic valida`tion of the interchain accounts GenesisState
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	return nil
}
