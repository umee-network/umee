package types_test

import (
	"testing"

	"gotest.tools/v3/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/util/coin"
	"github.com/umee-network/umee/v6/x/leverage/types"
)

func TestMaxBorrowScenarioA(t *testing.T) {
	// TODO: update the table to reflect normal asset groupings
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
			"  {0.5, 40 AAAA, 20 BBBB}\n"+ // $20 instead of $16 borrowed
			"  {0.4, 50 AAAA, 20 CCCC}\n"+ // no effect
			"collateral:\n"+
			"  100.000000000000000000AAAA\n"+ // +$40 ordinary borrow limit
			"  300.000000000000000000DDDD\n"+ // +$30 ordinary borrow limit
			"borrowed:\n"+
			"  20.000000000000000000BBBB\n"+
			"  20.000000000000000000CCCC\n"+
			"  5.000000000000000000DDDD",
		initialPosition.String(),
	)
	borrowLimit := initialPosition.Limit()
	assert.DeepEqual(t, sdk.MustNewDecFromStr("74.00"), borrowLimit) // 40 + 30 + (20 - 16) borrow limit

	// maxBorrow is more efficient than borrow limit predicts due to special pairs
	maxBorrow := initialPosition.MaxBorrow("BBBB")
	assert.DeepEqual(t, sdk.MustNewDecFromStr("29.00"), maxBorrow) // > $24

	// TODO: perfect the behavior of MaxBorrow and test that it matches finalPosition below (+$35)
	// It would need to move the collateral A in the special pair with C to the more efficient
	// pair with B, demoting the C borrow to ordinary assets.

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
			"collateral:\n"+
			"  100.000000000000000000AAAA\n"+
			"  300.000000000000000000DDDD\n"+
			"borrowed:\n"+
			"  55.000000000000000000BBBB\n"+
			"  20.000000000000000000CCCC\n"+
			"  5.000000000000000000DDDD",
		finalPosition.String(),
	)
}
