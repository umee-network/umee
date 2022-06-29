package wasm

import (
	"encoding/json"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v2/app/wasm/query"
	leveragekeeper "github.com/umee-network/umee/v2/x/leverage/keeper"
	"github.com/umee-network/umee/v2/x/leverage/types"
	oraclekeeper "github.com/umee-network/umee/v2/x/oracle/keeper"
)

// QueryPlugin wraps the query plugin with keepers
type QueryPlugin struct {
	leverageKeeper  leveragekeeper.Keeper
	oracleKeeper    oraclekeeper.Keeper
	leverageQuerier leveragekeeper.Querier
}

// NewQueryPlugin basic constructor
func NewQueryPlugin(
	leverageKeeper leveragekeeper.Keeper,
	oracleKeeper oraclekeeper.Keeper,
) *QueryPlugin {
	return &QueryPlugin{
		leverageKeeper:  leverageKeeper,
		oracleKeeper:    oracleKeeper,
		leverageQuerier: leveragekeeper.NewQuerier(leverageKeeper),
	}
}

// GetBorrow wraps leverage GetBorrow.
func (qp *QueryPlugin) GetBorrow(ctx sdk.Context, borrowerAddr sdk.AccAddress, denom string) sdk.Coin {
	return qp.leverageKeeper.GetBorrow(ctx, borrowerAddr, denom)
}

// GetExchangeRateBase wraps oracle GetExchangeRateBase.
func (qp *QueryPlugin) GetExchangeRateBase(ctx sdk.Context, denom string) (sdk.Dec, error) {
	return qp.oracleKeeper.GetExchangeRateBase(ctx, denom)
}

// GetAllRegisteredTokens wraps oracle GetAllRegisteredTokens.
func (qp *QueryPlugin) GetAllRegisteredTokens(ctx sdk.Context) []types.Token {
	return qp.leverageKeeper.GetAllRegisteredTokens(ctx)
}

// CustomQuerier implements custom querier for wasm smartcontracts acess umee native modules
func CustomQuerier(queryPlugin *QueryPlugin) func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		var smartcontractQuery query.UmeeQuery
		if err := json.Unmarshal(request, &smartcontractQuery); err != nil {
			return nil, sdkerrors.Wrap(err, "umee query")
		}

		switch smartcontractQuery.AssignedQuery {
		case query.AssignedQueryBorrowed:
			return smartcontractQuery.HandleBorrowed(ctx, queryPlugin.leverageQuerier)
		case query.AssignedQueryGetExchangeRateBase:
			return smartcontractQuery.HandleGetExchangeRateBase(ctx, queryPlugin)
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

		default:
			return nil, wasmvmtypes.UnsupportedRequest{Kind: "invalid assigned umee query"}
		}
	}
}
