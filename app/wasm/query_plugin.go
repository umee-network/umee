package wasm

import (
	"encoding/json"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v2/app/wasm/query"
	lvkeeper "github.com/umee-network/umee/v2/x/leverage/keeper"
	ockeeper "github.com/umee-network/umee/v2/x/oracle/keeper"
	ocpes "github.com/umee-network/umee/v2/x/oracle/types"
)

// QueryPlugin wraps the query plugin with queriers.
type QueryPlugin struct {
	leverageQuerier lvkeeper.Querier
	oracleQuerier   ocpes.QueryServer
}

// NewQueryPlugin creates a plugin to query native modules.
func NewQueryPlugin(
	leverageKeeper lvkeeper.Keeper,
	oracleKeeper ockeeper.Keeper,
) *QueryPlugin {
	return &QueryPlugin{
		leverageQuerier: lvkeeper.NewQuerier(leverageKeeper),
		oracleQuerier:   ockeeper.NewQuerier(oracleKeeper),
	}
}

// CustomQuerier implements custom querier for wasm smartcontracts acess umee native modules.
func CustomQuerier(queryPlugin *QueryPlugin) func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		var smartcontractQuery query.UmeeQuery
		if err := json.Unmarshal(request, &smartcontractQuery); err != nil {
			return nil, sdkerrors.Wrap(err, "umee query")
		}

		switch smartcontractQuery.AssignedQuery {
		case query.AssignedQueryBorrowed:
			return smartcontractQuery.HandleBorrowed(ctx, queryPlugin.leverageQuerier)
		case query.AssignedQueryExchangeRates:
			return smartcontractQuery.HandleExchangeRates(ctx, queryPlugin.oracleQuerier)
		case query.AssignedQueryRegisteredTokens:
			return smartcontractQuery.HandleRegisteredTokens(ctx, queryPlugin.leverageQuerier)
		case query.AssignedQueryLeverageParams:
			return smartcontractQuery.HandleLeverageParams(ctx, queryPlugin.leverageQuerier)
		case query.AssignedQueryBorrowedValue:
			return smartcontractQuery.HandleBorrowedValue(ctx, queryPlugin.leverageQuerier)
		case query.AssignedQueryLoaned:
			return smartcontractQuery.HandleLoaned(ctx, queryPlugin.leverageQuerier)
		case query.AssignedQueryLoanedValue:
			return smartcontractQuery.HandleLoanedValue(ctx, queryPlugin.leverageQuerier)
		case query.AssignedQueryAvailableBorrow:
			return smartcontractQuery.HandleAvailableBorrow(ctx, queryPlugin.leverageQuerier)
		case query.AssignedQueryBorrowAPY:
			return smartcontractQuery.HandleBorrowAPY(ctx, queryPlugin.leverageQuerier)
		case query.AssignedQueryLendAPY:
			return smartcontractQuery.HandleLendAPY(ctx, queryPlugin.leverageQuerier)
		case query.AssignedQueryMarketSize:
			return smartcontractQuery.HandleMarketSize(ctx, queryPlugin.leverageQuerier)
		case query.AssignedQueryTokenMarketSize:
			return smartcontractQuery.HandleTokenMarketSize(ctx, queryPlugin.leverageQuerier)
		case query.AssignedQueryReserveAmount:
			return smartcontractQuery.HandleReserveAmount(ctx, queryPlugin.leverageQuerier)
			// collateral stuffs can go here
		case query.AssignedQueryExchangeRate:
			return smartcontractQuery.HandleExchangeRate(ctx, queryPlugin.leverageQuerier)
		case query.AssignedQueryBorrowLimit:
			return smartcontractQuery.HandleBorrowLimit(ctx, queryPlugin.leverageQuerier)
		case query.AssignedQueryLiquidationThreshold:
			return smartcontractQuery.HandleLiquidationThreshold(ctx, queryPlugin.leverageQuerier)
		case query.AssignedQueryLiquidationTargets:
			return smartcontractQuery.HandleLiquidationTargets(ctx, queryPlugin.leverageQuerier)
		case query.AssignedQueryMarketSummary:
			return smartcontractQuery.HandleMarketSummary(ctx, queryPlugin.leverageQuerier)
		case query.AssignedQueryActiveExchangeRates:
			return smartcontractQuery.HandleActiveExchangeRates(ctx, queryPlugin.oracleQuerier)
		case query.AssignedQueryActiveFeederDelegation:
			return smartcontractQuery.HandleFeederDelegation(ctx, queryPlugin.oracleQuerier)
		case query.AssignedQueryMissCounter:
			return smartcontractQuery.HandleMissCounter(ctx, queryPlugin.oracleQuerier)
		case query.AssignedQueryAggregatePrevote:
			return smartcontractQuery.HandleAggregatePrevote(ctx, queryPlugin.oracleQuerier)
		case query.AssignedQueryAggregatePrevotes:
			return smartcontractQuery.HandleAggregatePrevotes(ctx, queryPlugin.oracleQuerier)

		default:
			return nil, wasmvmtypes.UnsupportedRequest{Kind: "invalid assigned umee query"}
		}
	}
}
