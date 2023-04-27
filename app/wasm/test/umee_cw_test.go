package wasm_test

import (
	"encoding/json"
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	appparams "github.com/umee-network/umee/v4/app/params"
	wm "github.com/umee-network/umee/v4/app/wasm/msg"
	wq "github.com/umee-network/umee/v4/app/wasm/query"
	lvtypes "github.com/umee-network/umee/v4/x/leverage/types"
	"github.com/umee-network/umee/v4/x/oracle/types"
	"gotest.tools/v3/assert"
)

func (s *IntegrationTestSuite) TestLeverageQueries() {
	tests := []struct {
		Name          string
		CQ            []byte
		ResponseCheck func(data []byte)
	}{
		{
			Name: "leverage query params",
			CQ: s.genCustomQuery(wq.UmeeQuery{
				AssignedQuery: wq.AssignedQueryLeverageParams,
			}),
			ResponseCheck: func(data []byte) {
				var rr lvtypes.QueryParamsResponse
				err := json.Unmarshal(data, &rr)
				assert.NilError(s.T, err)
				assert.DeepEqual(s.T, rr.Params, lvtypes.DefaultParams())
			},
		},
		{
			Name: "query all registered tokens",
			CQ: s.genCustomQuery(wq.UmeeQuery{
				AssignedQuery: wq.AssignedQueryRegisteredTokens,
			}),
			ResponseCheck: func(data []byte) {
				var rr lvtypes.QueryRegisteredTokensResponse
				err := json.Unmarshal(data, &rr)
				assert.NilError(s.T, err)
				assert.Equal(s.T, true, len(rr.Registry) > 0)
			},
		},
		{
			Name: "query registered token",
			CQ: s.genCustomQuery(wq.UmeeQuery{
				AssignedQuery: wq.AssignedQueryRegisteredTokens,
				RegisteredTokens: &lvtypes.QueryRegisteredTokens{
					BaseDenom: "uumee",
				},
			}),
			ResponseCheck: func(data []byte) {
				var rr lvtypes.QueryRegisteredTokensResponse
				err := json.Unmarshal(data, &rr)
				assert.NilError(s.T, err)
				assert.Equal(s.T, true, len(rr.Registry) > 0)
				assert.Equal(s.T, "uumee", rr.Registry[0].BaseDenom)

			},
		},
		{
			Name: "market summary",
			CQ: s.genCustomQuery(wq.UmeeQuery{
				AssignedQuery: wq.AssignedQueryMarketSummary,
				MarketSummary: &lvtypes.QueryMarketSummary{
					Denom: "uumee",
				},
			}),
			ResponseCheck: func(data []byte) {
				var rr lvtypes.QueryMarketSummaryResponse
				err := json.Unmarshal(data, &rr)
				assert.NilError(s.T, err)
				assert.Equal(s.T, "UMEE", rr.SymbolDenom)
			},
		},
	}

	for _, tc := range tests {
		s.T.Run(tc.Name, func(t *testing.T) {
			resp, err := s.wasmQueryClient.SmartContractState(sdk.WrapSDKContext(s.ctx), &wasmtypes.QuerySmartContractStateRequest{
				Address: s.contractAddr, QueryData: tc.CQ,
			})
			assert.NilError(s.T, err)
			tc.ResponseCheck(resp.Data)
		})
	}
}

func (s *IntegrationTestSuite) TestOracleQueries() {
	tests := []struct {
		Name          string
		CQ            []byte
		ResponseCheck func(data []byte)
	}{
		{
			Name: "oracle query params",
			CQ: s.genCustomQuery(wq.UmeeQuery{
				AssignedQuery: wq.AssignedQueryOracleParams,
				OracleParams:  &types.QueryParams{},
			}),
			ResponseCheck: func(data []byte) {
				var rr types.QueryParamsResponse
				err := json.Unmarshal(data, &rr)
				assert.NilError(s.T, err)
				assert.DeepEqual(s.T, rr.Params, types.DefaultParams())
			},
		},
		{
			Name: "oracle slash window",
			CQ: s.genCustomQuery(wq.UmeeQuery{
				AssignedQuery: wq.AssignedQuerySlashWindow,
				SlashWindow:   &types.QuerySlashWindow{},
			}),
			ResponseCheck: func(data []byte) {
				var rr types.QuerySlashWindowResponse
				err := json.Unmarshal(data, &rr)
				assert.NilError(s.T, err)
			},
		},
	}

	for _, tc := range tests {
		s.T.Run(tc.Name, func(t *testing.T) {
			resp := s.queryContract(tc.CQ)
			tc.ResponseCheck(resp.Data)
		})
	}
}

func (s *IntegrationTestSuite) TestLeverageTxs() {
	msg := s.genCustomTx(wm.UmeeMsg{
		AssignedMsg: wm.AssignedMsgSupply,
		Supply: &lvtypes.MsgSupply{
			Supplier: addr2.String(),
			Asset:    sdk.NewCoin(appparams.BondDenom, sdk.NewInt(100000)),
		},
	})
	s.execContract(addr2, msg)

	query := s.genCustomQuery(wq.UmeeQuery{
		AssignedQuery: wq.AssignedQueryAccountSummary,
		AccountSummary: &lvtypes.QueryAccountSummary{
			Address: addr2.String(),
		},
	})

	resp := s.queryContract(query)
	var rr lvtypes.QueryAccountSummaryResponse
	err := json.Unmarshal(resp.Data, &rr)
	assert.NilError(s.T, err)
}
