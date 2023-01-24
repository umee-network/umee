//go:build experimental
// +build experimental

package keeper_test

import "github.com/umee-network/umee/v4/x/uibc"

func (suite *KeeperTestSuite) TestGRPCQueryParams() {
	ctx, client := suite.ctx, suite.queryClient
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
		paramsResp, err := client.Params(ctx, &tc.req)
		if tc.errExpected {
			suite.Require().Error(err)
		} else {
			suite.Require().NoError(err)
			suite.Require().NotNil(paramsResp.Params)
		}
	}
}

func (suite *KeeperTestSuite) TestGRPCGetQuota() {
	ctx, client := suite.ctx, suite.queryClient
	tests := []struct {
		name        string
		req         uibc.QueryQuota
		errExpected bool
	}{
		{
			name:        "valid",
			req:         uibc.QueryQuota{},
			errExpected: false,
		},
		{
			name:        "valid req <error expected due to ibc-transfer not hapeen>",
			req:         uibc.QueryQuota{Denom: "umee"},
			errExpected: true,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			_, err := client.Quota(ctx, &tc.req)
			if tc.errExpected {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}
