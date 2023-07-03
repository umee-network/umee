package ugov

import (
	"github.com/umee-network/umee/v5/util/coin"
)

// DefaultGenesis creates a default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		MinGasPrice: coin.UmeeDec("0.1"),
	}
}

func (gs *GenesisState) Validate() error {
	return gs.MinGasPrice.Validate()
}
