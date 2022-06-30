package query

import (
	"encoding/json"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	lvkeeper "github.com/umee-network/umee/v2/x/leverage/keeper"
	ockeeper "github.com/umee-network/umee/v2/x/oracle/keeper"
	ocpes "github.com/umee-network/umee/v2/x/oracle/types"
)

// Plugin wraps the query plugin with queriers.
type Plugin struct {
	leverageQuerier lvkeeper.Querier
	oracleQuerier   ocpes.QueryServer
}

// NewQueryPlugin creates a plugin to query native modules.
func NewQueryPlugin(
	leverageKeeper lvkeeper.Keeper,
	oracleKeeper ockeeper.Keeper,
) *Plugin {
	return &Plugin{
		leverageQuerier: lvkeeper.NewQuerier(leverageKeeper),
		oracleQuerier:   ockeeper.NewQuerier(oracleKeeper),
	}
}

// CustomQuerier implements custom querier for wasm smartcontracts acess umee native modules.
func (plugin *Plugin) CustomQuerier() func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		var smartcontractQuery UmeeQuery
		if err := json.Unmarshal(request, &smartcontractQuery); err != nil {
			return nil, sdkerrors.Wrap(err, "umee query")
		}

		switch smartcontractQuery.AssignedQuery {
		case AssignedQueryBorrowed:
			return smartcontractQuery.HandleBorrowed(ctx, plugin.leverageQuerier)
		case AssignedQueryExchangeRates:
			return smartcontractQuery.HandleExchangeRates(ctx, plugin.oracleQuerier)
		case AssignedQueryRegisteredTokens:
			return smartcontractQuery.HandleRegisteredTokens(ctx, plugin.leverageQuerier)
		case AssignedQueryLeverageParams:
			return smartcontractQuery.HandleLeverageParams(ctx, plugin.leverageQuerier)
		case AssignedQueryBorrowedValue:
			return smartcontractQuery.HandleBorrowedValue(ctx, plugin.leverageQuerier)
		case AssignedQueryLoaned:
			return smartcontractQuery.HandleLoaned(ctx, plugin.leverageQuerier)
		case AssignedQueryLoanedValue:
			return smartcontractQuery.HandleLoanedValue(ctx, plugin.leverageQuerier)
		case AssignedQueryAvailableBorrow:
			return smartcontractQuery.HandleAvailableBorrow(ctx, plugin.leverageQuerier)
		case AssignedQueryBorrowAPY:
			return smartcontractQuery.HandleBorrowAPY(ctx, plugin.leverageQuerier)
		case AssignedQueryLendAPY:
			return smartcontractQuery.HandleLendAPY(ctx, plugin.leverageQuerier)
		case AssignedQueryMarketSize:
			return smartcontractQuery.HandleMarketSize(ctx, plugin.leverageQuerier)
		case AssignedQueryTokenMarketSize:
			return smartcontractQuery.HandleTokenMarketSize(ctx, plugin.leverageQuerier)
		case AssignedQueryReserveAmount:
			return smartcontractQuery.HandleReserveAmount(ctx, plugin.leverageQuerier)
			// collateral stuffs can go here
		case AssignedQueryExchangeRate:
			return smartcontractQuery.HandleExchangeRate(ctx, plugin.leverageQuerier)
		case AssignedQueryBorrowLimit:
			return smartcontractQuery.HandleBorrowLimit(ctx, plugin.leverageQuerier)
		case AssignedQueryLiquidationThreshold:
			return smartcontractQuery.HandleLiquidationThreshold(ctx, plugin.leverageQuerier)
		case AssignedQueryLiquidationTargets:
			return smartcontractQuery.HandleLiquidationTargets(ctx, plugin.leverageQuerier)
		case AssignedQueryMarketSummary:
			return smartcontractQuery.HandleMarketSummary(ctx, plugin.leverageQuerier)
		case AssignedQueryActiveExchangeRates:
			return smartcontractQuery.HandleActiveExchangeRates(ctx, plugin.oracleQuerier)
		case AssignedQueryActiveFeederDelegation:
			return smartcontractQuery.HandleFeederDelegation(ctx, plugin.oracleQuerier)
		case AssignedQueryMissCounter:
			return smartcontractQuery.HandleMissCounter(ctx, plugin.oracleQuerier)
		case AssignedQueryAggregatePrevote:
			return smartcontractQuery.HandleAggregatePrevote(ctx, plugin.oracleQuerier)
		case AssignedQueryAggregatePrevotes:
			return smartcontractQuery.HandleAggregatePrevotes(ctx, plugin.oracleQuerier)
		case AssignedQueryAggregateVote:
			return smartcontractQuery.HandleAggregateVote(ctx, plugin.oracleQuerier)
		case AssignedQueryAggregateVotes:
			return smartcontractQuery.HandleAggregateVotes(ctx, plugin.oracleQuerier)
		case AssignedQueryOracleParams:
			return smartcontractQuery.HandleOracleParams(ctx, plugin.oracleQuerier)

		default:
			return nil, wasmvmtypes.UnsupportedRequest{Kind: "invalid assigned umee query"}
		}
	}
}
