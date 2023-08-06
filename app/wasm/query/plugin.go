package query

import (
	"encoding/json"
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cosmos/gogoproto/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/x/incentive"
	inckeeper "github.com/umee-network/umee/v5/x/incentive/keeper"
	lvkeeper "github.com/umee-network/umee/v5/x/leverage/keeper"
	lvtypes "github.com/umee-network/umee/v5/x/leverage/types"
	"github.com/umee-network/umee/v5/x/metoken"
	metokenkeeper "github.com/umee-network/umee/v5/x/metoken/keeper"
	ockeeper "github.com/umee-network/umee/v5/x/oracle/keeper"
	ocpes "github.com/umee-network/umee/v5/x/oracle/types"
)

// Plugin wraps the query plugin with queriers.
type Plugin struct {
	lvQueryServer      lvtypes.QueryServer
	ocQueryServer      ocpes.QueryServer
	incQueryServer     incentive.QueryServer
	metokenQueryServer metoken.QueryServer
}

// NewQueryPlugin creates a plugin to query native modules.
func NewQueryPlugin(
	leverageKeeper lvkeeper.Keeper,
	oracleKeeper ockeeper.Keeper,
	incentiveKeeper inckeeper.Keeper,
	metokenBuilder metokenkeeper.Builder,
) *Plugin {
	return &Plugin{
		lvQueryServer:      lvkeeper.NewQuerier(leverageKeeper),
		ocQueryServer:      ockeeper.NewQuerier(oracleKeeper),
		incQueryServer:     inckeeper.NewQuerier(incentiveKeeper),
		metokenQueryServer: metokenkeeper.NewQuerier(metokenBuilder),
	}
}

// CustomQuerier implements custom querier for wasm smartcontracts acess umee native modules.
func (plugin *Plugin) CustomQuerier() func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		var smartcontractQuery UmeeQuery
		if err := json.Unmarshal(request, &smartcontractQuery); err != nil {
			return nil, sdkerrors.Wrap(err, "umee query")
		}

		var resp proto.Message
		var err error

		switch {
		case smartcontractQuery.LeverageParameters != nil:
			resp, err = smartcontractQuery.HandleLeverageParams(ctx, plugin.lvQueryServer)
		case smartcontractQuery.RegisteredTokens != nil:
			resp, err = smartcontractQuery.HandleRegisteredTokens(ctx, plugin.lvQueryServer)
		case smartcontractQuery.MarketSummary != nil:
			resp, err = smartcontractQuery.HandleMarketSummary(ctx, plugin.lvQueryServer)
		case smartcontractQuery.AccountBalances != nil:
			resp, err = smartcontractQuery.HandleAccountBalances(ctx, plugin.lvQueryServer)
		case smartcontractQuery.AccountSummary != nil:
			resp, err = smartcontractQuery.HandleAccountSummary(ctx, plugin.lvQueryServer)
		case smartcontractQuery.LiquidationTargets != nil:
			resp, err = smartcontractQuery.HandleLiquidationTargets(ctx, plugin.lvQueryServer)
		case smartcontractQuery.BadDebts != nil:
			resp, err = smartcontractQuery.HandleBadDebts(ctx, plugin.lvQueryServer)
		case smartcontractQuery.MaxWithdraw != nil:
			resp, err = smartcontractQuery.HandleMaxWithdraw(ctx, plugin.lvQueryServer)
		case smartcontractQuery.MaxBorrow != nil:
			resp, err = smartcontractQuery.HandleMaxBorrow(ctx, plugin.lvQueryServer)

		case smartcontractQuery.FeederDelegation != nil:
			resp, err = smartcontractQuery.HandleFeederDelegation(ctx, plugin.ocQueryServer)
		case smartcontractQuery.MissCounter != nil:
			resp, err = smartcontractQuery.HandleMissCounter(ctx, plugin.ocQueryServer)
		case smartcontractQuery.SlashWindow != nil:
			resp, err = smartcontractQuery.HandleSlashWindow(ctx, plugin.ocQueryServer)
		case smartcontractQuery.AggregatePrevote != nil:
			resp, err = smartcontractQuery.HandleAggregatePrevote(ctx, plugin.ocQueryServer)
		case smartcontractQuery.AggregatePrevotes != nil:
			resp, err = smartcontractQuery.HandleAggregatePrevotes(ctx, plugin.ocQueryServer)
		case smartcontractQuery.AggregateVote != nil:
			resp, err = smartcontractQuery.HandleAggregateVote(ctx, plugin.ocQueryServer)
		case smartcontractQuery.AggregateVotes != nil:
			resp, err = smartcontractQuery.HandleAggregateVotes(ctx, plugin.ocQueryServer)
		case smartcontractQuery.OracleParams != nil:
			resp, err = smartcontractQuery.HandleOracleParams(ctx, plugin.ocQueryServer)
		case smartcontractQuery.ExchangeRates != nil:
			resp, err = smartcontractQuery.HandleExchangeRates(ctx, plugin.ocQueryServer)
		case smartcontractQuery.ActiveExchangeRates != nil:
			resp, err = smartcontractQuery.HandleActiveExchangeRates(ctx, plugin.ocQueryServer)
		case smartcontractQuery.Medians != nil:
			resp, err = smartcontractQuery.HandleMedians(ctx, plugin.ocQueryServer)
		case smartcontractQuery.MedianDeviations != nil:
			resp, err = smartcontractQuery.HandleMedianDeviations(ctx, plugin.ocQueryServer)

		// incentive
		case smartcontractQuery.IncentiveParameters != nil:
			resp, err = smartcontractQuery.HandleIncentiveParams(ctx, plugin.incQueryServer)
		case smartcontractQuery.TotalBonded != nil:
			resp, err = smartcontractQuery.HandleTotalBonded(ctx, plugin.incQueryServer)
		case smartcontractQuery.TotalUnbonding != nil:
			resp, err = smartcontractQuery.HandleTotalUnbonding(ctx, plugin.incQueryServer)
		case smartcontractQuery.AccountBonds != nil:
			resp, err = smartcontractQuery.HandleAccountBonds(ctx, plugin.incQueryServer)
		case smartcontractQuery.PendingRewards != nil:
			resp, err = smartcontractQuery.HandlePendingRewards(ctx, plugin.incQueryServer)
		case smartcontractQuery.CompletedIncentivePrograms != nil:
			resp, err = smartcontractQuery.HandleCompletedIncentivePrograms(ctx, plugin.incQueryServer)
		case smartcontractQuery.UpcomingIncentivePrograms != nil:
			resp, err = smartcontractQuery.HandleUpcomingIncentivePrograms(ctx, plugin.incQueryServer)
		case smartcontractQuery.OngoingIncentivePrograms != nil:
			resp, err = smartcontractQuery.HandleOngoingIncentivePrograms(ctx, plugin.incQueryServer)
		case smartcontractQuery.IncentiveProgram != nil:
			resp, err = smartcontractQuery.HandleIncentiveParams(ctx, plugin.incQueryServer)
		case smartcontractQuery.CurrentRates != nil:
			resp, err = smartcontractQuery.HandleCurrentRates(ctx, plugin.incQueryServer)
		case smartcontractQuery.ActualRates != nil:
			resp, err = smartcontractQuery.HandleActualRates(ctx, plugin.incQueryServer)
		case smartcontractQuery.LastRewardTime != nil:
			resp, err = smartcontractQuery.HandleLastRewardTime(ctx, plugin.incQueryServer)

			// metoken
		case smartcontractQuery.MeTokenParameters != nil:
			resp, err = smartcontractQuery.HandleMeTokenParams(ctx, plugin.metokenQueryServer)
		case smartcontractQuery.Indexes != nil:
			resp, err = smartcontractQuery.HandleMeTokenIndexes(ctx, plugin.metokenQueryServer)
		case smartcontractQuery.SwapFee != nil:
			resp, err = smartcontractQuery.HandleMeTokenSwapFee(ctx, plugin.metokenQueryServer)
		case smartcontractQuery.RedeemFee != nil:
			resp, err = smartcontractQuery.HandleMeTokenRedeemFee(ctx, plugin.metokenQueryServer)
		case smartcontractQuery.IndexBalances != nil:
			resp, err = smartcontractQuery.HandleMeTokenIndexBalances(ctx, plugin.metokenQueryServer)
		case smartcontractQuery.IndexPrice != nil:
			resp, err = smartcontractQuery.HandleMeTokenIndexPrice(ctx, plugin.metokenQueryServer)

		default:
			return nil, wasmvmtypes.UnsupportedRequest{Kind: "invalid umee query"}
		}

		if err != nil {
			return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("error %+v while query the assigned query ", err)}
		}

		return MarshalResponse(resp)
	}
}
