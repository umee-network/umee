package types_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v6/x/leverage/types"
)

func TestPriceModes(t *testing.T) {
	tcs := []struct {
		mode            types.PriceMode
		withoutHistoric types.PriceMode
		allowsExpired   bool
	}{
		{
			types.PriceModeSpot,
			types.PriceModeSpot,
			false,
		},
	}

	for _, tc := range tcs {
		assert.Equal(t,
			tc.mode.IgnoreHistoric(),
			tc.withoutHistoric,
		)
		assert.Equal(t,
			tc.allowsExpired,
			tc.mode.AllowsExpired(),
		)
	}
}
