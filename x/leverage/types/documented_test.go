package types_test

import (
	"testing"

	"gotest.tools/v3/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/util/coin"
	"github.com/umee-network/umee/v5/x/leverage/types"
)

func TestDocScenario(t *testing.T) {
	borrowPosition := types.NewAccountPosition(
		[]types.Token{
			testToken("AAAA", "0.4", "0.5"),
			testToken("BBBB", "0.3", "0.5"),
			testToken("CCCC", "0.2", "0.5"),
			testToken("DDDD", "0.1", "0.5"),
		},
		[]types.SpecialAssetPair{
			testPair("AAAA", "BBBB", "0.5", "0.5"),
			testPair("BBBB", "AAAA", "0.5", "0.5"),

			testPair("AAAA", "CCCC", "0.4", "0.4"),
			testPair("CCCC", "AAAA", "0.4", "0.4"),
		},
		sdk.NewDecCoins(
			coin.Dec("AAAA", "100"),
			coin.Dec("DDDD", "300"),
		),
		sdk.NewDecCoins(
			coin.Dec("BBBB", "55"),
			coin.Dec("CCCC", "20"),
			coin.Dec("DDDD", "5"),
		),
		false,
	)
	assert.Equal(t,
		sdk.MustNewDecFromStr("80.00").String(),
		borrowPosition.Limit().String(),
	)
}
