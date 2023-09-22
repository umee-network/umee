package types

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
)

// NewGenesisState creates a new GenesisState object
func NewGenesisState(
	params Params,
	rates []DenomExchangeRate,
	feederDelegations []FeederDelegation,
	missCounters []MissCounter,
	aggregateExchangeRatePrevotes []AggregateExchangeRatePrevote,
	aggregateExchangeRateVotes []AggregateExchangeRateVote,
	historicPrices []Price,
	medianPrices []Price,
	medianDeviationPrices []Price,
	acp AvgCounterParams,
) *GenesisState {
	return &GenesisState{
		Params:                        params,
		ExchangeRates:                 rates,
		FeederDelegations:             feederDelegations,
		MissCounters:                  missCounters,
		AggregateExchangeRatePrevotes: aggregateExchangeRatePrevotes,
		AggregateExchangeRateVotes:    aggregateExchangeRateVotes,
		HistoricPrices:                historicPrices,
		Medians:                       medianPrices,
		MedianDeviations:              medianDeviationPrices,
		AvgCounterParams:              acp,
	}
}

// DefaultGenesisState returns the default genesis state for the x/oracle
// module.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:                        DefaultParams(),
		ExchangeRates:                 []DenomExchangeRate{},
		FeederDelegations:             []FeederDelegation{},
		MissCounters:                  []MissCounter{},
		AggregateExchangeRatePrevotes: []AggregateExchangeRatePrevote{},
		AggregateExchangeRateVotes:    []AggregateExchangeRateVote{},
		HistoricPrices:                []Price{},
		Medians:                       []Price{},
		MedianDeviations:              []Price{},
		AvgCounterParams:              DefaultAvgCounterParams(),
	}
}

// ValidateGenesis validates the oracle genesis state.
func ValidateGenesis(data *GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}

	return data.AvgCounterParams.Validate()
}

// GetGenesisStateFromAppState returns x/oracle GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}

func DefaultAvgCounterParams() AvgCounterParams {
	return AvgCounterParams{
		AvgPeriod: DefaultAvgPeriod, // 16 hours
		AvgShift:  DefaultAvgShift,  // 12 hours
	}
}

func (acp AvgCounterParams) Equal(other *AvgCounterParams) bool {
	return acp.AvgPeriod == other.AvgPeriod && acp.AvgShift == other.AvgShift
}

func (acp AvgCounterParams) Validate() error {
	if acp.AvgPeriod.Seconds() <= 0 {
		return fmt.Errorf("avg period must be positive: %d", acp.AvgPeriod)
	}

	if acp.AvgShift.Seconds() <= 0 {
		return fmt.Errorf("avg shift must be positive: %d", acp.AvgShift)
	}

	return nil
}
