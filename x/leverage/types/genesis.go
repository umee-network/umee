package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DefaultGenesis returns the default genesis state of the x/leverage module.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: Params{
			InterestEpoch: sdk.NewInt(100),
		},
		Registry: []Token{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {

	if !gs.Params.InterestEpoch.IsPositive() {
		return sdkerrors.Wrap(ErrInvalidEpoch, gs.Params.InterestEpoch.String())
	}

	for _, token := range gs.Registry {
		if err := token.Validate(); err != nil {
			return err
		}
	}

	return nil
}
