package main

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestIBCDenom(t *testing.T) {

	tests := []struct {
		name           string
		baseDenom      string
		channelID      string
		errExpected    string
		execptedResult string
	}{
		{
			name:           "invalid base denom",
			baseDenom:      "",
			channelID:      "channel-12",
			errExpected:    "base denomination cannot be blank",
			execptedResult: "",
		},
		{
			name:           "invalid channel-id",
			baseDenom:      "uakt",
			channelID:      "channel",
			errExpected:    "invalid channel ID",
			execptedResult: "",
		},
		{
			name:           "success",
			baseDenom:      "uakt",
			channelID:      "channel-12",
			errExpected:    "",
			execptedResult: "ibc/8DF58541612917752DA1CCACC8441FCFE367F9960E51151968A75CE22671D717",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ibcDenom, err := ibcDenom(test.baseDenom, test.channelID)
			if len(test.errExpected) != 0 {
				assert.ErrorContains(t, err, test.errExpected)
			} else {
				assert.NilError(t, err)
				assert.Equal(t, ibcDenom, test.execptedResult)
			}
		})
	}
}
