//go:build experimental
// +build experimental

package keeper_test

import (
	"testing"

	"github.com/umee-network/umee/v4/x/uibc"
	"gotest.tools/v3/assert"
)

func TestGRPCQueryParams(t *testing.T) {
	s := initIntegrationSuite(t)
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

func TestGRPCGetQuota(t *testing.T) {
	t.Parallel()
	suite := initIntegrationSuite(t)
	ctx, client := suite.ctx, suite.queryClient
	tests := []struct {
		name   string
		req    uibc.QueryQuota
		errMsg string
	}{
		{
			name:   "valid",
			req:    uibc.QueryQuota{},
			errMsg: "",
		}, {
			name:   "valid req <get: uumee>",
			req:    uibc.QueryQuota{Denom: "umee"},
			errMsg: "no quota for ibc denom",
		}, {
			name:   "valid req <error expected due to ibc-transfer not hapeen>",
			req:    uibc.QueryQuota{Denom: "umee"},
			errMsg: "no quota for ibc denom",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.Quota(ctx, &tc.req)
			if tc.errMsg == "" {
				assert.NilError(t, err)
			} else {
				assert.Error(t, err, tc.errMsg)
			}
		})
	}
}
