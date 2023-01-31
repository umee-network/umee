//go:build experimental
// +build experimental

package keeper_test

import (
	"testing"

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
		errMsg      string
	}{
		{
			name:        "valid",
			req:         uibc.QueryParams{},
			errExpected: false,
			errMsg:      "",
		},
	}

	for _, tc := range tests {
		_, err := client.Params(ctx, &tc.req)
		if tc.errExpected {
			assert.Error(t, err, tc.errMsg)
		} else {
			assert.NilError(t, err)
		}
	}
}

func TestGRPCGetQuota(t *testing.T) {
	suite := initKeeperTestSuite(t)
	ctx, client := suite.ctx, suite.queryClient
	tests := []struct {
		name        string
		req         uibc.QueryQuota
		errExpected bool
		errMsg      string
	}{
		{
			name:        "valid",
			req:         uibc.QueryQuota{},
			errExpected: false,
			errMsg:      "",
		},
		{
			name:        "valid req <error expected due to ibc-transfer not hapeen>",
			req:         uibc.QueryQuota{Denom: "umee"},
			errExpected: true,
			errMsg:      "no quota for ibc denom",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.Quota(ctx, &tc.req)
			if tc.errExpected {
				assert.Error(t, err, tc.errMsg)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}
