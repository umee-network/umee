package wasm_test

import (
	"encoding/json"
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"

	appparams "github.com/umee-network/umee/v6/app/params"
	wq "github.com/umee-network/umee/v6/app/wasm/query"
	"github.com/umee-network/umee/v6/x/incentive"
	lvtypes "github.com/umee-network/umee/v6/x/leverage/types"
	"github.com/umee-network/umee/v6/x/metoken"
)

func (s *IntegrationTestSuite) TestStargateQueries() {
	tests := []struct {
		name string
		sq   func() StargateQuery
		resp func(resp wasmtypes.QuerySmartContractStateResponse)
	}{
		{
			name: "stargate: leverage params ",
			sq: func() StargateQuery {
				data := lvtypes.QueryParams{}
				d, err := data.Marshal()
				assert.NilError(s.T, err)
				sq := StargateQuery{}
				sq.Chain.Stargate = wasmvmtypes.StargateQuery{
					Path: "/umee.leverage.v1.Query/Params",
					Data: d,
				}
				return sq
			},
			resp: func(resp wasmtypes.QuerySmartContractStateResponse) {
				var rr lvtypes.QueryParamsResponse
				err := s.encfg.Codec.UnmarshalJSON(resp.Data, &rr)
				assert.NilError(s.T, err)
				assert.DeepEqual(s.T, lvtypes.DefaultParams(), rr.Params)
			},
		},
		{
			name: "stargate: metoken queries ",
			sq: func() StargateQuery {
				data := metoken.QueryParams{}
				d, err := data.Marshal()
				assert.NilError(s.T, err)
				sq := StargateQuery{}
				sq.Chain.Stargate = wasmvmtypes.StargateQuery{
					Path: "/umee.metoken.v1.Query/Params",
					Data: d,
				}
				return sq
			},
			resp: func(resp wasmtypes.QuerySmartContractStateResponse) {
				var rr metoken.QueryParamsResponse
				err := s.encfg.Codec.UnmarshalJSON(resp.Data, &rr)
				assert.NilError(s.T, err)
				assert.DeepEqual(s.T, metoken.DefaultParams(), rr.Params)
			},
		},
		{
			name: "stargate: leverage market summary",
			sq: func() StargateQuery {
				data := lvtypes.QueryMarketSummary{
					Denom: appparams.BondDenom,
				}
				d, err := data.Marshal()
				assert.NilError(s.T, err)
				sq := StargateQuery{}
				sq.Chain.Stargate = wasmvmtypes.StargateQuery{
					Path: "/umee.leverage.v1.Query/MarketSummary",
					Data: d,
				}
				return sq
			},
			resp: func(resp wasmtypes.QuerySmartContractStateResponse) {
				var rr lvtypes.QueryMarketSummaryResponse
				err := s.encfg.Codec.UnmarshalJSON(resp.Data, &rr)
				assert.NilError(s.T, err)
				assert.Equal(s.T, "UMEE", rr.SymbolDenom)
			},
		},
	}

	for _, test := range tests {
		s.T.Run(test.name, func(t *testing.T) {
			cq, err := json.Marshal(test.sq())
			assert.NilError(s.T, err)
			resp, err := s.wasmQueryClient.SmartContractState(sdk.WrapSDKContext(s.ctx), &wasmtypes.QuerySmartContractStateRequest{
				Address: s.contractAddr, QueryData: cq,
			})
			assert.NilError(s.T, err)
			test.resp(*resp)
		})
	}
}

func (s *IntegrationTestSuite) TestIncentiveQueries() {
	tests := []struct {
		Name          string
		CQ            []byte
		ResponseCheck func(data []byte)
	}{
		{
			Name: "incentive query params",
			CQ: s.genCustomQuery(wq.UmeeQuery{
				IncentiveParameters: &incentive.QueryParams{},
			}),
			ResponseCheck: func(data []byte) {
				var rr incentive.QueryParamsResponse
				err := json.Unmarshal(data, &rr)
				assert.NilError(s.T, err)
				assert.DeepEqual(s.T, rr.Params, incentive.DefaultParams())
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

func (s *IntegrationTestSuite) TestMetokenQueries() {
	tests := []struct {
		Name          string
		CQ            []byte
		ResponseCheck func(data []byte)
	}{
		{
			Name: "metoken query params",
			CQ: s.genCustomQuery(wq.UmeeQuery{
				MeTokenParameters: &metoken.QueryParams{},
			}),
			ResponseCheck: func(data []byte) {
				var rr metoken.QueryParamsResponse
				err := json.Unmarshal(data, &rr)
				assert.NilError(s.T, err)
				assert.DeepEqual(s.T, rr.Params, metoken.DefaultParams())
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
