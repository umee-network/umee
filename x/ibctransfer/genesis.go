package ibctransfer

func NewGenesisState(params Params, rateLimits []RateLimit) *GenesisState {
	return &GenesisState{
		Params:     params,
		RateLimits: rateLimits,
	}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:     *DefaultParams(),
		RateLimits: nil,
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

	return nil
}
