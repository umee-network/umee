package wasm_test

import (
	"encoding/json"
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"gotest.tools/v3/assert"

	appparams "github.com/umee-network/umee/v5/app/params"
	wm "github.com/umee-network/umee/v5/app/wasm/msg"
	wq "github.com/umee-network/umee/v5/app/wasm/query"
	"github.com/umee-network/umee/v5/x/incentive"
	lvtypes "github.com/umee-network/umee/v5/x/leverage/types"
	"github.com/umee-network/umee/v5/x/oracle/types"
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
				LeverageParameters: &lvtypes.QueryParams{},
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
				RegisteredTokens: &lvtypes.QueryRegisteredTokens{},
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
				RegisteredTokens: &lvtypes.QueryRegisteredTokens{
					BaseDenom: appparams.BondDenom,
				},
			}),
			ResponseCheck: func(data []byte) {
				var rr lvtypes.QueryRegisteredTokensResponse
				err := json.Unmarshal(data, &rr)
				assert.NilError(s.T, err)
				assert.Equal(s.T, true, len(rr.Registry) > 0)
				assert.Equal(s.T, appparams.BondDenom, rr.Registry[0].BaseDenom)
			},
		},
		{
			Name: "market summary",
			CQ: s.genCustomQuery(wq.UmeeQuery{
				MarketSummary: &lvtypes.QueryMarketSummary{
					Denom: appparams.BondDenom,
				},
			}),
			ResponseCheck: func(data []byte) {
				var rr lvtypes.QueryMarketSummaryResponse
				err := json.Unmarshal(data, &rr)
				assert.NilError(s.T, err)
				assert.Equal(s.T, "UMEE", rr.SymbolDenom)
			},
		},
		{
			Name: "query bad debts",
			CQ: s.genCustomQuery(wq.UmeeQuery{
				BadDebts: &lvtypes.QueryBadDebts{},
			}),
			ResponseCheck: func(data []byte) {
				var rr lvtypes.QueryBadDebtsResponse
				err := json.Unmarshal(data, &rr)
				assert.NilError(s.T, err)
				assert.Equal(s.T, true, len(rr.Targets) == 0)
			},
		},
		{
			Name: "query max withdraw (zero)",
			CQ: s.genCustomQuery(wq.UmeeQuery{
				MaxWithdraw: &lvtypes.QueryMaxWithdraw{
					Address: addr.String(),
					Denom:   appparams.BondDenom,
				},
			}),
			ResponseCheck: func(data []byte) {
				var rr lvtypes.QueryMaxWithdrawResponse
				err := json.Unmarshal(data, &rr)
				assert.NilError(s.T, err)
				assert.Equal(s.T, true, len(rr.Tokens) == 0)
				assert.Equal(s.T, true, len(rr.UTokens) == 0)
			},
		},
		{
			Name: "query max borrow (zero)",
			CQ: s.genCustomQuery(wq.UmeeQuery{
				MaxBorrow: &lvtypes.QueryMaxBorrow{
					Address: addr.String(),
					Denom:   appparams.BondDenom,
				},
			}),
			ResponseCheck: func(data []byte) {
				var rr lvtypes.QueryMaxBorrowResponse
				err := json.Unmarshal(data, &rr)
				assert.NilError(s.T, err)
				assert.Equal(s.T, true, len(rr.Tokens) == 0)
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

func (s *IntegrationTestSuite) TestStargateQueries() {
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

	cq, err := json.Marshal(sq)
	assert.NilError(s.T, err)
	resp, err := s.wasmQueryClient.SmartContractState(sdk.WrapSDKContext(s.ctx), &wasmtypes.QuerySmartContractStateRequest{
		Address: s.contractAddr, QueryData: cq,
	})
	assert.NilError(s.T, err)
	var rr lvtypes.QueryMarketSummaryResponse
	err = s.encfg.Codec.UnmarshalJSON(resp.Data, &rr)
	assert.NilError(s.T, err)
	assert.Equal(s.T, "UMEE", rr.SymbolDenom)
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
				OracleParams: &types.QueryParams{},
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
				SlashWindow: &types.QuerySlashWindow{},
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
	accAddr := sdk.MustAccAddressFromBech32(s.contractAddr)
	err := s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, minttypes.ModuleName, accAddr, sdk.NewCoins(sdk.NewCoin(appparams.BondDenom, sdk.NewInt(100000))))
	assert.NilError(s.T, err)
	txTests := []struct {
		Name string
		Msg  []byte
	}{
		{
			Name: "supply",
			Msg: s.genCustomTx(wm.UmeeMsg{
				Supply: &lvtypes.MsgSupply{
					Supplier: s.contractAddr,
					Asset:    sdk.NewCoin(appparams.BondDenom, sdk.NewInt(700)),
				},
			}),
		},
		{
			Name: "add collateral",
			Msg: s.genCustomTx(wm.UmeeMsg{
				Collateralize: &lvtypes.MsgCollateralize{
					Borrower: s.contractAddr,
					Asset:    sdk.NewCoin("u/uumee", sdk.NewInt(700)),
				},
			}),
		},
		// {
		// 	Name: "borrow",
		// 	Msg: s.genCustomTx(wm.UmeeMsg{
		// 		Borrow: &lvtypes.MsgBorrow{
		// 			Borrower: addr2.String(),
		// 			Asset:    sdk.NewCoin(appparams.BondDenom, sdk.NewInt(150)),
		// 		},
		// 	}),
		// },
	}

	for _, tc := range txTests {
		s.T.Run(tc.Name, func(t *testing.T) {
			s.execContract(addr2, tc.Msg)
		})
	}

	query := s.genCustomQuery(wq.UmeeQuery{
		AccountSummary: &lvtypes.QueryAccountSummary{
			Address: addr2.String(),
		},
	})

	resp := s.queryContract(query)
	var rr lvtypes.QueryAccountSummaryResponse
	err = json.Unmarshal(resp.Data, &rr)
	assert.NilError(s.T, err)
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
