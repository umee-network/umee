package query

import (
	"encoding/json"
	"fmt"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	inctypes "github.com/umee-network/umee/v5/x/incentive"
	lvtypes "github.com/umee-network/umee/v5/x/leverage/types"
	octypes "github.com/umee-network/umee/v5/x/oracle/types"
)

// UmeeQuery wraps all the queries availables for cosmwasm smartcontracts.
type UmeeQuery struct {
	// Leverage queries
	// Used to query the x/leverage module's parameters.
	LeverageParameters *lvtypes.QueryParams `json:"leverage_parameters,omitempty"`
	// Used to query all the registered tokens.
	RegisteredTokens *lvtypes.QueryRegisteredTokens `json:"registered_tokens,omitempty"`
	// Used to get the summary data of an denom.
	MarketSummary *lvtypes.QueryMarketSummary `json:"market_summary,omitempty"`
	// Used to get an account's current supply, collateral, and borrow positions.
	AccountBalances *lvtypes.QueryAccountBalances `json:"account_balances,omitempty"`
	// Used to queries USD values representing an account's total positions and borrowing limits.
	AccountSummary *lvtypes.QueryAccountSummary `json:"account_summary,omitempty"`
	// request to return a list of borrower addresses eligible for liquidation.
	LiquidationTargets *lvtypes.QueryLiquidationTargets `json:"liquidation_targets,omitempty"`
	// requet to returns list of bad debts
	BadDebts *lvtypes.QueryBadDebts `json:"bad_debts_params,omitempty"`
	// request to returns max withdraw
	MaxWithdraw *lvtypes.QueryMaxWithdraw `json:"max_withdraw_params,omitempty"`
	// request to get max borrows
	MaxBorrow *lvtypes.QueryMaxBorrow `json:"max_borrow_params,omitempty"`

	//  oracle queries
	// Used to get all feeder delegation of a validator.
	FeederDelegation *octypes.QueryFeederDelegation `json:"feeder_delegation,omitempty"`
	// Used to get all the oracle miss counter of a validator.
	MissCounter *octypes.QueryMissCounter `json:"miss_counter,omitempty"`
	// Used to get information of an slash window.
	SlashWindow *octypes.QuerySlashWindow `json:"slash_window,omitempty"`
	// Used to get an aggregate prevote of a validator.
	AggregatePrevote *octypes.QueryAggregatePrevote `json:"aggregate_prevote,omitempty"`
	// Used to get an aggregate prevote of all validators.
	AggregatePrevotes *octypes.QueryAggregatePrevotes `json:"aggregate_prevotes,omitempty"`
	// Used to get an aggregate vote of a validator.
	AggregateVote *octypes.QueryAggregateVote `json:"aggregate_vote,omitempty"`
	// Used to get an aggregate vote of all validators.
	AggregateVotes *octypes.QueryAggregateVotes `json:"aggregate_votes,omitempty"`
	// Used to query the x/oracle module's parameters.
	OracleParams *octypes.QueryParams `json:"oracle_params,omitempty"`
	// Used to get the exchange rates of all denoms.
	ExchangeRates *octypes.QueryExchangeRates `json:"exchange_rates,omitempty"`
	// Used to get all active denoms.
	ActiveExchangeRates *octypes.QueryActiveExchangeRates `json:"active_exchange_rates,omitempty"`
	// Used to get all medians.
	Medians *octypes.QueryMedians `json:"medians,omitempty"`
	// Used to get all median deviations.
	MedianDeviations *octypes.QueryMedianDeviations `json:"median_deviations,omitempty"`

	// incentive queries
	// Incentive module params .
	IncentiveParams *inctypes.QueryParams `json:"incentive_params,omitempty"`
	// TotalBonded queries the sum of all bonded collateral uTokens.
	TotalBonded *inctypes.QueryTotalBonded `json:"total_bonded,omitempty"`
	// TotalUnbonding queries the sum of all unbonding collateral uTokens.
	TotalUnbonding *inctypes.QueryTotalUnbonding `json:"total_unbonding,omitempty"`
	// AccountBonds queries all bonded collateral and unbondings associated with an account.
	AccountBonds *inctypes.QueryAccountBonds `json:"account_bonds,omitempty"`
	// PendingRewards queries unclaimed incentive rewards associated with an account.
	PendingRewards *inctypes.QueryPendingRewards `json:"pending_rewards,omitempty"`
	// CompletedIncentivePrograms queries for all incentives programs that have been passed
	// by governance,
	CompletedIncentivePrograms *inctypes.QueryCompletedIncentivePrograms `json:"completed,omitempty"`
	// OngoingIncentivePrograms queries for all incentives programs that have been passed
	// by governance, funded, and started but not yet completed.
	OngoingIncentivePrograms *inctypes.QueryOngoingIncentivePrograms `json:"ongoing,omitempty"`
	// UpcomingIncentivePrograms queries for all incentives programs that have been passed
	// by governance, but not yet started. They may or may not have been funded.
	UpcomingIncentivePrograms *inctypes.QueryUpcomingIncentivePrograms `json:"upcoming,omitempty"`
	// IncentiveProgram queries a single incentive program by ID.
	IncentiveProgram *inctypes.QueryIncentiveProgram `json:"program,omitempty"`
	// CurrentRates queries the hypothetical return of a bonded uToken denomination
	// if current incentive rewards continued for one year. The response is an sdk.Coins
	// of base token rewards, per reference amount (usually 10^exponent of the uToken.)
	CurrentRates *inctypes.QueryCurrentRates `json:"current_rates,omitempty"`
	// ActualRates queries the hypothetical return of a bonded uToken denomination
	// if current incentive rewards continued for one year. The response is an sdk.Dec
	// representing an oracle-adjusted APY.
	ActualRates *inctypes.QueryActualRates `json:"actual_rates,omitempty"`
	// LastRewardTime queries the last block time at which incentive rewards were calculated.
	LastRewardTime *inctypes.QueryLastRewardTime `json:"last_reward_time,omitempty"`
}

// MarshalResponse marshals any response.
func MarshalResponse(resp interface{}) ([]byte, error) {
	bz, err := json.Marshal(resp)
	if err != nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v umee query response error on marshal", err)}
	}
	return bz, err
}
