package types

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewGenesisState creates a new GenesisState object
func NewGenesisState(
	params Params,
	tokens []Token,
	borrows []Borrow,
	collateralSettings []CollateralSetting,
	collateral []Collateral,
	reserves sdk.Coins,
	lastInterestTime int64,
	exchangeRates []ExchangeRate,
	badDebts []BadDebt,
	borrowAPYs []APY,
	lendAPYs []APY,
) *GenesisState {

	return &GenesisState{
		Params:             params,
		Registry:           tokens,
		Borrows:            borrows,
		CollateralSettings: collateralSettings,
		Collateral:         collateral,
		LastInterestTime:   lastInterestTime,
		ExchangeRates:      exchangeRates,
		BadDebts:           badDebts,
		BorrowRates:        borrowAPYs,
		LendRates:          lendAPYs,
	}
}

// DefaultGenesis returns the default genesis state of the x/leverage module.
func DefaultGenesis() *GenesisState {

	return &GenesisState{
		Params:           DefaultParams(),
		Registry:         []Token{},
		LastInterestTime: 0,
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

	for _, borrow := range gs.Borrows {
		if _, err := sdk.AccAddressFromBech32(borrow.Address); err != nil {
			return err
		}

		if err := borrow.Amount.Validate(); err != nil {
			return err
		}
	}

	for _, setting := range gs.CollateralSettings {
		if _, err := sdk.AccAddressFromBech32(setting.Address); err != nil {
			return err
		}

		if err := sdk.ValidateDenom(setting.Denom); err != nil {
			return err
		}
	}

	for _, collateral := range gs.Collateral {
		if _, err := sdk.AccAddressFromBech32(collateral.Address); err != nil {
			return err
		}

		if err := collateral.Amount.Validate(); err != nil {
			return err
		}
	}

	if err := gs.Reserves.Validate(); err != nil {
		return err
	}

	for _, rate := range gs.ExchangeRates {
		if err := sdk.ValidateDenom(rate.Denom); err != nil {
			return err
		}

		if rate.ExchangeRate.LT(sdk.OneDec()) {
			return sdkerrors.Wrap(ErrInvalidExchangeRate, rate.String())
		}
	}

	for _, badDebt := range gs.BadDebts {
		if _, err := sdk.AccAddressFromBech32(badDebt.Address); err != nil {
			return err
		}

		if err := sdk.ValidateDenom(badDebt.Denom); err != nil {
			return err
		}
	}

	for _, rate := range gs.BorrowRates {
		if err := sdk.ValidateDenom(rate.Denom); err != nil {
			return err
		}

		if rate.APY.IsNegative() {
			return sdkerrors.Wrap(ErrNegativeAPY, rate.String())
		}
	}

	for _, rate := range gs.LendRates {
		if err := sdk.ValidateDenom(rate.Denom); err != nil {
			return err
		}

		if rate.APY.IsNegative() {
			return sdkerrors.Wrap(ErrNegativeAPY, rate.String())
		}
	}

	return nil
}

// GetGenesisStateFromAppState returns x/leverage GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}

// NewBorrow creates the Borrow struct used in GenesisState
func NewBorrow(addr string, amount sdk.Coin) Borrow {
	return Borrow{
		Address: addr,
		Amount:  amount,
	}
}

// NewCollateral creates the Collateral struct used in GenesisState
func NewCollateral(addr string, amount sdk.Coin) Collateral {
	return Collateral{
		Address: addr,
		Amount:  amount,
	}
}

// NewCollateralSetting creates the CollateralSetting struct used in GenesisState
func NewCollateralSetting(addr, denom string) CollateralSetting {
	return CollateralSetting{
		Address: addr,
		Denom:   denom,
	}
}

// NewBadDebt creates the BadDebt struct used in GenesisState
func NewBadDebt(addr, denom string) BadDebt {
	return BadDebt{
		Address: addr,
		Denom:   denom,
	}
}

// NewExchangeRate creates the APY struct used in GenesisState
func NewExchangeRate(denom string, rate sdk.Dec) ExchangeRate {
	return ExchangeRate{
		Denom:        denom,
		ExchangeRate: rate,
	}
}

// NewAPY creates the APY struct used in GenesisState
func NewAPY(denom string, rate sdk.Dec) APY {
	return APY{
		Denom: denom,
		APY:   rate,
	}
}
