package types_test

import (
	"testing"

	"gotest.tools/v3/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/util/coin"
	"github.com/umee-network/umee/v6/x/leverage/types"
)

func TestMaxBorrowScenarioA(t *testing.T) {
	// This borrow position reproduces the initial table of "MaxBorrow Scenario A" from x/leverage/EXAMPLES.md
	initialPosition, err := types.NewAccountPosition(
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
			coin.Dec("BBBB", "20"),
			coin.Dec("CCCC", "20"),
			coin.Dec("DDDD", "5"),
		),
		false,
		noMinimumBorrowFactor,
	)
	assert.NilError(t, err)
	assert.Equal(t,
		"special:\n"+
			"  {0.5, 40 AAAA, 20 BBBB}\n"+
			"  {0.5, 0 BBBB, 0 AAAA}\n"+
			"  {0.4, 50 AAAA, 20 CCCC}\n"+
			"  {0.4, 0 CCCC, 0 AAAA}\n"+
			"normal:\n"+
			"  [10 AAAA (0.4), 1 DDDD (0.1)]\n"+
			"  [40 DDDD (0.1), 4 DDDD (0.1)]\n"+
			"  [260 DDDD (0.1), -]\n",
		initialPosition.String(),
	)
	borrowLimit := initialPosition.Limit()
	assert.DeepEqual(t, sdk.MustNewDecFromStr("71.00"), borrowLimit) // $45 borrowed + $26 generic max borrow

	// the current naive implementation of maxBorrow produces a result below the optimal 35.00
	maxBorrow, err := initialPosition.MaxBorrow("BBBB")
	assert.NilError(t, err)
	assert.DeepEqual(t, sdk.MustNewDecFromStr("30.00"), maxBorrow)
	// TODO: perfect the behavior of MaxBorrow and test that it matches finalPosition below.
	// It would need to move the collateral A in the special pair with C to the more efficient
	// pair with B, demoting the C borrow to ordinary assets.
	assert.Equal(t,
		"special:\n"+
			"  {0.5, 50 AAAA, 25 BBBB}\n"+
			"  {0.5, 0 BBBB, 0 AAAA}\n"+
			"  {0.4, 50 AAAA, 20 CCCC}\n"+
			"  {0.4, 0 CCCC, 0 AAAA}\n"+
			"normal:\n"+
			"  [250 DDDD (0.1), 25 BBBB (0.3)]\n"+
			"  [50 DDDD (0.1), 5 DDDD (0.1)]\n",
		initialPosition.String(),
	)

	// This borrow position reproduces the final table of "MaxBorrow Scenario A" from x/leverage/EXAMPLES.md
	finalPosition, err := types.NewAccountPosition(
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
		noMinimumBorrowFactor,
	)
	assert.NilError(t, err)
	assert.Equal(t,
		"special:\n"+
			"  {0.5, 100 AAAA, 50 BBBB}\n"+
			"  {0.5, 0 BBBB, 0 AAAA}\n"+
			"  {0.4, 0 AAAA, 0 CCCC}\n"+
			"  {0.4, 0 CCCC, 0 AAAA}\n"+
			"normal:\n"+
			"  [50 DDDD (0.1), 5 BBBB (0.3)]\n"+
			"  [200 DDDD (0.1), 20 CCCC (0.2)]\n"+
			"  [50 DDDD (0.1), 5 DDDD (0.1)]\n",
		finalPosition.String(),
	)
}
