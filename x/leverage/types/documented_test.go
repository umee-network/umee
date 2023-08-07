package types_test

import (
	"testing"

	"gotest.tools/v3/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/util/coin"
	"github.com/umee-network/umee/v5/x/leverage/types"
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
			"  40.000000000000000000AAAA, 20.000000000000000000BBBB, 0.500000000000000000\n"+
			"  0.000000000000000000BBBB, 0.000000000000000000AAAA, 0.500000000000000000\n"+
			"  50.000000000000000000AAAA, 20.000000000000000000CCCC, 0.400000000000000000\n"+
			"  0.000000000000000000CCCC, 0.000000000000000000AAAA, 0.400000000000000000\n"+
			"normal:\n"+
			"  {10.000000000000000000AAAA 0.400000000000000000}, {1.000000000000000000DDDD 0.100000000000000000}\n"+
			"  {40.000000000000000000DDDD 0.100000000000000000}, {4.000000000000000000DDDD 0.100000000000000000}\n"+
			"  {260.000000000000000000DDDD 0.100000000000000000}, -\n",
		initialPosition.String(),
	)
	borrowLimit := initialPosition.Limit()
	assert.DeepEqual(t, sdk.MustNewDecFromStr("71.00"), borrowLimit) // $45 borrowed + $26 generic max borrow

	// the current naive implementatio of maxBorrow produces a result below the optimal 35.00
	maxBorrow, err := initialPosition.MaxBorrow("BBBB")
	assert.NilError(t, err)
	assert.DeepEqual(t, sdk.MustNewDecFromStr("30.00"), maxBorrow)
	// TODO: perfect the behavior of MaxBorrow and test that it matches finalPosition below.
	// It would need to move the collateral A in the special pair with C to the more efficient
	// pair with B, demoting the C borrow to ordinary assets.
	assert.Equal(t,
		"special:\n"+
			"  50.000000000000000000AAAA, 25.000000000000000000BBBB, 0.500000000000000000\n"+
			"  0.000000000000000000BBBB, 0.000000000000000000AAAA, 0.500000000000000000\n"+
			"  50.000000000000000000AAAA, 20.000000000000000000CCCC, 0.400000000000000000\n"+
			"  0.000000000000000000CCCC, 0.000000000000000000AAAA, 0.400000000000000000\n"+
			"normal:\n"+
			"  {250.000000000000000000DDDD 0.100000000000000000}, {25.000000000000000000BBBB 0.300000000000000000}\n"+
			"  {50.000000000000000000DDDD 0.100000000000000000}, {5.000000000000000000DDDD 0.100000000000000000}\n",
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
			"  100.000000000000000000AAAA, 50.000000000000000000BBBB, 0.500000000000000000\n"+
			"  0.000000000000000000BBBB, 0.000000000000000000AAAA, 0.500000000000000000\n"+
			"  0.000000000000000000AAAA, 0.000000000000000000CCCC, 0.400000000000000000\n"+
			"  0.000000000000000000CCCC, 0.000000000000000000AAAA, 0.400000000000000000\n"+
			"normal:\n"+
			"  {50.000000000000000000DDDD 0.100000000000000000}, {5.000000000000000000BBBB 0.300000000000000000}\n"+
			"  {200.000000000000000000DDDD 0.100000000000000000}, {20.000000000000000000CCCC 0.200000000000000000}\n"+
			"  {50.000000000000000000DDDD 0.100000000000000000}, {5.000000000000000000DDDD 0.100000000000000000}\n",
		finalPosition.String(),
	)
}
