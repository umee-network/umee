package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestParams(t *testing.T) {
	require := require.New(t)
	p := DefaultParams()
	check := func(errMsg, msg string) {
		err := p.Validate()
		if errMsg == "" {
			require.NoError(err, msg)
		} else {
			require.ErrorContains(err, errMsg, msg)
		}
	}

	check("", "default must validate")

	p.MaxShareTrigger = 0
	check("MaxShareTrigger must be", "MaxShareTrigger must not be 0")

	p.MaxShareTrigger = 1
	check("", "MaxShareTrigger works with 1")

	p = DefaultParams()
	p.CompleteLiquidationThreshold = sdk.NewDec(2)
	check("complete_liquidation_threshold must be", "complete_liquidation_threshold too big")

	p.CompleteLiquidationThreshold = sdk.MustNewDecFromStr("1.1")
	check("complete_liquidation_threshold must be", "complete_liquidation_threshold too big")

	p.CompleteLiquidationThreshold = sdk.MustNewDecFromStr("1")
	check("", "complete_liquidation_threshold works with 1")
}
