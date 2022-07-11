package query

import (
	"encoding/json"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	lvkeeper "github.com/umee-network/umee/v2/x/leverage/keeper"
	lvtypes "github.com/umee-network/umee/v2/x/leverage/types"
	ockeeper "github.com/umee-network/umee/v2/x/oracle/keeper"
	ocpes "github.com/umee-network/umee/v2/x/oracle/types"
)

// Plugin wraps the query plugin with queriers.
type Plugin struct {
	lvQueryServer lvtypes.QueryServer
	ocQueryServer ocpes.QueryServer
}

// NewQueryPlugin creates a plugin to query native modules.
func NewQueryPlugin(
	leverageKeeper lvkeeper.Keeper,
	oracleKeeper ockeeper.Keeper,
) *Plugin {
	return &Plugin{
		lvQueryServer: lvkeeper.NewQuerier(leverageKeeper),
		ocQueryServer: ockeeper.NewQuerier(oracleKeeper),
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
			return smartcontractQuery.HandleBorrowed(ctx, plugin.lvQueryServer)
		case AssignedQueryExchangeRates:
			return smartcontractQuery.HandleExchangeRates(ctx, plugin.ocQueryServer)
		case AssignedQueryRegisteredTokens:
			return smartcontractQuery.HandleRegisteredTokens(ctx, plugin.lvQueryServer)
		case AssignedQueryLeverageParams:
			return smartcontractQuery.HandleLeverageParams(ctx, plugin.lvQueryServer)
		case AssignedQueryBorrowedValue:
			return smartcontractQuery.HandleBorrowedValue(ctx, plugin.lvQueryServer)
		case AssignedQuerySupplied:
			return smartcontractQuery.HandleSupplied(ctx, plugin.lvQueryServer)
		case AssignedQuerySuppliedValue:
			return smartcontractQuery.HandleSuppliedValue(ctx, plugin.lvQueryServer)
		case AssignedQueryAvailableBorrow:
			return smartcontractQuery.HandleAvailableBorrow(ctx, plugin.lvQueryServer)
		case AssignedQueryBorrowAPY:
			return smartcontractQuery.HandleBorrowAPY(ctx, plugin.lvQueryServer)
		case AssignedQuerySupplyAPY:
			return smartcontractQuery.HandleSupplyAPY(ctx, plugin.lvQueryServer)
		case AssignedQueryMarketSize:
			return smartcontractQuery.HandleMarketSize(ctx, plugin.lvQueryServer)
		case AssignedQueryTokenMarketSize:
			return smartcontractQuery.HandleTokenMarketSize(ctx, plugin.lvQueryServer)
		case AssignedQueryReserveAmount:
			return smartcontractQuery.HandleReserveAmount(ctx, plugin.lvQueryServer)
		case AssignedQueryCollateral:
			return smartcontractQuery.HandleCollateral(ctx, plugin.lvQueryServer)
		case AssignedQueryCollateralValue:
			return smartcontractQuery.HandleCollateralValue(ctx, plugin.lvQueryServer)
		case AssignedQueryExchangeRate:
			return smartcontractQuery.HandleExchangeRate(ctx, plugin.lvQueryServer)
		case AssignedQueryBorrowLimit:
			return smartcontractQuery.HandleBorrowLimit(ctx, plugin.lvQueryServer)
		case AssignedQueryLiquidationThreshold:
			return smartcontractQuery.HandleLiquidationThreshold(ctx, plugin.lvQueryServer)
		case AssignedQueryLiquidationTargets:
			return smartcontractQuery.HandleLiquidationTargets(ctx, plugin.lvQueryServer)
		case AssignedQueryMarketSummary:
			return smartcontractQuery.HandleMarketSummary(ctx, plugin.lvQueryServer)
		case AssignedQueryTotalCollateral:
			return smartcontractQuery.HandleTotalCollateral(ctx, plugin.lvQueryServer)
		case AssignedQueryTotalBorrowed:
			return smartcontractQuery.HandleTotalBorrowed(ctx, plugin.lvQueryServer)
		case AssignedQueryActiveExchangeRates:
			return smartcontractQuery.HandleActiveExchangeRates(ctx, plugin.ocQueryServer)
		case AssignedQueryActiveFeederDelegation:
			return smartcontractQuery.HandleFeederDelegation(ctx, plugin.ocQueryServer)
		case AssignedQueryMissCounter:
			return smartcontractQuery.HandleMissCounter(ctx, plugin.ocQueryServer)
		case AssignedQueryAggregatePrevote:
			return smartcontractQuery.HandleAggregatePrevote(ctx, plugin.ocQueryServer)
		case AssignedQueryAggregatePrevotes:
			return smartcontractQuery.HandleAggregatePrevotes(ctx, plugin.ocQueryServer)
		case AssignedQueryAggregateVote:
			return smartcontractQuery.HandleAggregateVote(ctx, plugin.ocQueryServer)
		case AssignedQueryAggregateVotes:
			return smartcontractQuery.HandleAggregateVotes(ctx, plugin.ocQueryServer)
		case AssignedQueryOracleParams:
			return smartcontractQuery.HandleOracleParams(ctx, plugin.ocQueryServer)

		default:
			return nil, wasmvmtypes.UnsupportedRequest{Kind: "invalid assigned umee query"}
		}
	}
}
