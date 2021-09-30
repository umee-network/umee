package types

// DefaultGenesis returns the default genesis state of the x/leverage module.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: Params{},
		Assets: []Asset{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	for _, asset := range gs.Assets {
		if err := asset.Validate(); err != nil {
			return err
		}
	}

	return nil
}
