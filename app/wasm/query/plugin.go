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
	leverageQueryServer lvtypes.QueryServer
	oracleQueryServer   ocpes.QueryServer
}

// NewQueryPlugin creates a plugin to query native modules.
func NewQueryPlugin(
	leverageKeeper lvkeeper.Keeper,
	oracleKeeper ockeeper.Keeper,
) *Plugin {
	return &Plugin{
		leverageQueryServer: lvkeeper.NewQuerier(leverageKeeper),
		oracleQueryServer:   ockeeper.NewQuerier(oracleKeeper),
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
			return smartcontractQuery.HandleBorrowed(ctx, plugin.leverageQueryServer)
		case AssignedQueryExchangeRates:
			return smartcontractQuery.HandleExchangeRates(ctx, plugin.oracleQueryServer)
		case AssignedQueryRegisteredTokens:
			return smartcontractQuery.HandleRegisteredTokens(ctx, plugin.leverageQueryServer)
		case AssignedQueryLeverageParams:
			return smartcontractQuery.HandleLeverageParams(ctx, plugin.leverageQueryServer)
		case AssignedQueryBorrowedValue:
			return smartcontractQuery.HandleBorrowedValue(ctx, plugin.leverageQueryServer)
		case AssignedQueryLoaned:
			return smartcontractQuery.HandleLoaned(ctx, plugin.leverageQueryServer)
		case AssignedQueryLoanedValue:
			return smartcontractQuery.HandleLoanedValue(ctx, plugin.leverageQueryServer)
		case AssignedQueryAvailableBorrow:
			return smartcontractQuery.HandleAvailableBorrow(ctx, plugin.leverageQueryServer)
		case AssignedQueryBorrowAPY:
			return smartcontractQuery.HandleBorrowAPY(ctx, plugin.leverageQueryServer)
		case AssignedQueryLendAPY:
			return smartcontractQuery.HandleLendAPY(ctx, plugin.leverageQueryServer)
		case AssignedQueryMarketSize:
			return smartcontractQuery.HandleMarketSize(ctx, plugin.leverageQueryServer)
		case AssignedQueryTokenMarketSize:
			return smartcontractQuery.HandleTokenMarketSize(ctx, plugin.leverageQueryServer)
		case AssignedQueryReserveAmount:
			return smartcontractQuery.HandleReserveAmount(ctx, plugin.leverageQueryServer)
		case AssignedQueryCollateral:
			return smartcontractQuery.HandleCollateral(ctx, plugin.leverageQueryServer)
		case AssignedQueryCollateralValue:
			return smartcontractQuery.HandleCollateralValue(ctx, plugin.leverageQueryServer)
		case AssignedQueryExchangeRate:
			return smartcontractQuery.HandleExchangeRate(ctx, plugin.leverageQueryServer)
		case AssignedQueryBorrowLimit:
			return smartcontractQuery.HandleBorrowLimit(ctx, plugin.leverageQueryServer)
		case AssignedQueryLiquidationThreshold:
			return smartcontractQuery.HandleLiquidationThreshold(ctx, plugin.leverageQueryServer)
		case AssignedQueryLiquidationTargets:
			return smartcontractQuery.HandleLiquidationTargets(ctx, plugin.leverageQueryServer)
		case AssignedQueryMarketSummary:
			return smartcontractQuery.HandleMarketSummary(ctx, plugin.leverageQueryServer)
		case AssignedQueryActiveExchangeRates:
			return smartcontractQuery.HandleActiveExchangeRates(ctx, plugin.oracleQueryServer)
		case AssignedQueryActiveFeederDelegation:
			return smartcontractQuery.HandleFeederDelegation(ctx, plugin.oracleQueryServer)
		case AssignedQueryMissCounter:
			return smartcontractQuery.HandleMissCounter(ctx, plugin.oracleQueryServer)
		case AssignedQueryAggregatePrevote:
			return smartcontractQuery.HandleAggregatePrevote(ctx, plugin.oracleQueryServer)
		case AssignedQueryAggregatePrevotes:
			return smartcontractQuery.HandleAggregatePrevotes(ctx, plugin.oracleQueryServer)
		case AssignedQueryAggregateVote:
			return smartcontractQuery.HandleAggregateVote(ctx, plugin.oracleQueryServer)
		case AssignedQueryAggregateVotes:
			return smartcontractQuery.HandleAggregateVotes(ctx, plugin.oracleQueryServer)
		case AssignedQueryOracleParams:
			return smartcontractQuery.HandleOracleParams(ctx, plugin.oracleQueryServer)

		default:
			return nil, wasmvmtypes.UnsupportedRequest{Kind: "invalid assigned umee query"}
		}
	}
}
