package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v4/x/uibc"
	"gotest.tools/v3/assert"
)

func TestGRPCQueryParams(t *testing.T) {
	s := initKeeperTestSuite(t)
	ctx, client := s.ctx, s.queryClient
	tests := []struct {
		name        string
		req         uibc.QueryParams
		errExpected bool
	}{
		{
			name:        "valid",
			req:         uibc.QueryParams{},
			errExpected: false,
		},
	}

	for _, tc := range tests {
		_, err := client.Params(ctx, &tc.req)
		if tc.errExpected {
			assert.Error(t, err, "")
		} else {
			assert.NilError(t, err)
		}
	}
}

func TestGRPCQueryOutflows(t *testing.T) {
	suite := initKeeperTestSuite(t)
	ctx, client := suite.ctx, suite.queryClient
	tests := []struct {
		name          string
		req           uibc.QueryOutflows
		outflowAmount int64
	}{
		{
			name:          "valid: total outflows",
			req:           uibc.QueryOutflows{},
			outflowAmount: 0,
		}, {
			name:          "valid req: OutflowSum zero because ibc-transfer not hapeen",
			req:           uibc.QueryOutflows{Denom: "umee"},
			outflowAmount: 0,
		}, {
			name:          "non existing denom",
			req:           uibc.QueryOutflows{Denom: "doesntexists"},
			outflowAmount: 0,
		}, {
			name:          "existing denom",
			req:           uibc.QueryOutflows{Denom: "utest"},
			outflowAmount: 1111,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// query outflows never returns error if there is no denom present it will return 0
			resp, err := client.Outflows(ctx, &tc.req)

			assert.NilError(t, err)
			assert.DeepEqual(t, sdk.NewDec(tc.outflowAmount), resp.Amount)
		})
	}

	t.Run("all-outflows", func(t *testing.T) {
		resp, err := client.AllOutflows(ctx, &uibc.QueryAllOutflows{})
		assert.NilError(t, err)
		assert.Equal(t, 1, len(resp.Outflows))
	})
}
