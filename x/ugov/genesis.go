package ugov

import (
	"time"

	"github.com/umee-network/umee/v6/util/coin"
)

// DefaultGenesis creates a default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		MinGasPrice:       coin.UmeeDec("0.1"),
		InflationParams:   DefaultInflationParams(),
		InflationCycleEnd: time.Unix(1, 0),
	}
}

func (gs *GenesisState) Validate() error {
	if err := gs.MinGasPrice.Validate(); err != nil {
		return err
	}

	return gs.InflationParams.Validate()
}
