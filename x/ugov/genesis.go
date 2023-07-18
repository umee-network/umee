package ugov

import (
	"github.com/umee-network/umee/v5/util/coin"
)

// DefaultGenesis creates a default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		MinGasPrice:     coin.UmeeDec("0.1"),
		InflationParams: DefaultInflationParams(),
	}
}

func (gs *GenesisState) Validate() error {
	if err := gs.MinGasPrice.Validate(); err != nil {
		return err
	}

	return gs.InflationParams.Validate()
}
