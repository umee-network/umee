package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v2/x/leverage/types"
)

func TestComputeLiquidation(t *testing.T) {
	type testCase struct {
		availableRepay       sdk.Int
		availableCollateral  sdk.Int
		availableReward      sdk.Int
		repayTokenPrice      sdk.Dec
		rewardTokenPrice     sdk.Dec
		uTokenExchangeRate   sdk.Dec
		liquidationIncentive sdk.Dec
		closeFactor          sdk.Dec
		borrowedValue        sdk.Dec
	}

	baseCase := func() testCase {
		return testCase{
			sdk.NewInt(1000),               // 1000 Token A to repay
			sdk.NewInt(5000),               // 5000 uToken B collateral
			sdk.NewInt(5000),               // 5000 Token B liquidity
			sdk.OneDec(),                   // price(A) = $1
			sdk.OneDec(),                   // price(B) = $1
			sdk.OneDec(),                   // utoken exchange rate 1 u/B => 1 B
			sdk.MustNewDecFromStr("0.1"),   // reward value is 110% repay value
			sdk.OneDec(),                   // unlimited close factor
			sdk.MustNewDecFromStr("10000"), // $10000 borrowed value
		}
	}

	runTestCase := func(tc testCase, expectedRepay, expectedCollateral, expectedReward int64, msg string) {
		repay, collateral, reward := types.ComputeLiquidation(
			tc.availableRepay,
			tc.availableCollateral,
			tc.availableReward,
			tc.repayTokenPrice,
			tc.rewardTokenPrice,
			tc.uTokenExchangeRate,
			tc.liquidationIncentive,
			tc.closeFactor,
			tc.borrowedValue,
		)

		require.Equal(t, sdk.NewInt(expectedRepay), repay, msg+" (repay)")
		require.Equal(t, sdk.NewInt(expectedCollateral), collateral, msg+" (collateral)")
		require.Equal(t, sdk.NewInt(expectedReward), reward, msg+" (reward)")
	}

	// basic liquidation of 1000 borrowed tokens with plenty of available rewards and collateral
	runTestCase(baseCase(), 1000, 1100, 1100, "base case")

	// borrower is healthy (as implied by a close factor of zero) so liquidation cannot occur
	healthyCase := baseCase()
	healthyCase.closeFactor = sdk.ZeroDec()
	runTestCase(healthyCase, 0, 0, 0, "healthy borrower")

	// limiting factor is available repay
	repayLimited := baseCase()
	repayLimited.availableRepay = sdk.NewInt(100)
	runTestCase(repayLimited, 100, 110, 110, "repay limited")

	// limiting factor is available collateral
	collateralLimited := baseCase()
	collateralLimited.availableCollateral = sdk.NewInt(220)
	runTestCase(collateralLimited, 200, 220, 220, "collateral limited")

	// limiting factor is available reward
	rewardLimited := baseCase()
	rewardLimited.availableReward = sdk.NewInt(330)
	runTestCase(rewardLimited, 300, 330, 330, "reward limited")

	// repay token is worth more
	expensiveRepay := baseCase()
	expensiveRepay.repayTokenPrice = sdk.MustNewDecFromStr("2")
	runTestCase(expensiveRepay, 1000, 2200, 2200, "expensive repay")

	// reward token is worth more
	expensiveReward := baseCase()
	expensiveReward.rewardTokenPrice = sdk.MustNewDecFromStr("2")
	runTestCase(expensiveReward, 1000, 550, 550, "expensive reward")

	// high collateral uToken exchange rate
	exchangeRate := baseCase()
	exchangeRate.uTokenExchangeRate = sdk.MustNewDecFromStr("2")
	runTestCase(exchangeRate, 1000, 550, 1100, "high uToken exchange rate")

	// high liquidation incentive
	highIncentive := baseCase()
	highIncentive.liquidationIncentive = sdk.MustNewDecFromStr("1.5")
	runTestCase(highIncentive, 1000, 2500, 2500, "high liquidation incentive")

	// no liquidation incentive
	noIncentive := baseCase()
	noIncentive.liquidationIncentive = sdk.ZeroDec()
	runTestCase(noIncentive, 1000, 1000, 1000, "no liquidation incentive")

	// partial close factor
	partialClose := baseCase()
	partialClose.closeFactor = sdk.MustNewDecFromStr("0.03")
	runTestCase(partialClose, 300, 330, 330, "close factor")

	// lowered borrowed value
	lowValue := baseCase()
	lowValue.borrowedValue = sdk.MustNewDecFromStr("700")
	runTestCase(lowValue, 700, 770, 770, "lowered borrowed value")

	// complex case, limited by available repay, with various nontrivial values
	complexCase := baseCase()
	complexCase.availableRepay = sdk.NewInt(300)
	complexCase.uTokenExchangeRate = sdk.MustNewDecFromStr("2.5")
	complexCase.liquidationIncentive = sdk.MustNewDecFromStr("0.5")
	complexCase.repayTokenPrice = sdk.MustNewDecFromStr("6")
	complexCase.rewardTokenPrice = sdk.MustNewDecFromStr("12")
	// repay = 300 (limiting factor)
	// collateral = 300 * 1.5 * (6/12) / 2.5 = 0.3 * 300 = 90
	// reward = 300 * 1.5 * (6/12) = 225
	runTestCase(complexCase, 300, 90, 225, "complex case")

	// borrow dust case, with high borrowed token value and no rounding
	expensiveBorrowDust := baseCase()
	expensiveBorrowDust.availableRepay = sdk.NewInt(1)
	expensiveBorrowDust.repayTokenPrice = sdk.MustNewDecFromStr("40")
	expensiveBorrowDust.rewardTokenPrice = sdk.MustNewDecFromStr("2")
	expensiveBorrowDust.liquidationIncentive = sdk.MustNewDecFromStr("0")
	runTestCase(expensiveBorrowDust, 1, 20, 20, "expensive borrow dust")

	// borrow dust case, with high borrowed token value rounds reward down
	expensiveBorrowDustDown := baseCase()
	expensiveBorrowDustDown.availableRepay = sdk.NewInt(1)
	expensiveBorrowDustDown.repayTokenPrice = sdk.MustNewDecFromStr("39.9")
	expensiveBorrowDustDown.rewardTokenPrice = sdk.MustNewDecFromStr("2")
	expensiveBorrowDustDown.liquidationIncentive = sdk.MustNewDecFromStr("0")
	runTestCase(expensiveBorrowDustDown, 1, 19, 19, "expensive borrow dust with price down")

	// borrow dust case, with high borrowed token value rounds collateral burn up
	expensiveBorrowDustUp := baseCase()
	expensiveBorrowDustUp.availableRepay = sdk.NewInt(1)
	expensiveBorrowDustUp.repayTokenPrice = sdk.MustNewDecFromStr("40.1")
	expensiveBorrowDustUp.rewardTokenPrice = sdk.MustNewDecFromStr("2")
	expensiveBorrowDustUp.liquidationIncentive = sdk.MustNewDecFromStr("0")
	runTestCase(expensiveBorrowDustUp, 1, 20, 20, "expensive borrow dust with price up")

	// borrow dust case, with low borrowed token value rounds collateral burn and reward to zero
	cheapBorrowDust := baseCase()
	cheapBorrowDust.availableRepay = sdk.NewInt(1)
	cheapBorrowDust.repayTokenPrice = sdk.MustNewDecFromStr("2")
	cheapBorrowDust.rewardTokenPrice = sdk.MustNewDecFromStr("40")
	cheapBorrowDust.liquidationIncentive = sdk.MustNewDecFromStr("0")
	runTestCase(cheapBorrowDust, 1, 0, 0, "cheap borrow dust")

	// collateral dust case, with high collateral token value and no rounding
	expensiveCollateralDust := baseCase()
	expensiveCollateralDust.availableCollateral = sdk.NewInt(1)
	expensiveCollateralDust.repayTokenPrice = sdk.MustNewDecFromStr("2")
	expensiveCollateralDust.rewardTokenPrice = sdk.MustNewDecFromStr("40")
	expensiveCollateralDust.liquidationIncentive = sdk.MustNewDecFromStr("0")
	runTestCase(expensiveCollateralDust, 20, 1, 1, "expensive collateral dust")

	// collateral dust case, with high collateral token value and no rounding
	expensiveCollateralDustUp := baseCase()
	expensiveCollateralDustUp.availableCollateral = sdk.NewInt(1)
	expensiveCollateralDustUp.repayTokenPrice = sdk.MustNewDecFromStr("2")
	expensiveCollateralDustUp.rewardTokenPrice = sdk.MustNewDecFromStr("40.1")
	expensiveCollateralDustUp.liquidationIncentive = sdk.MustNewDecFromStr("0")
	runTestCase(expensiveCollateralDustUp, 21, 1, 1, "expensive collateral dust with price up")

	// collateral dust case, with high collateral token value and no rounding
	expensiveCollateralDustDown := baseCase()
	expensiveCollateralDustDown.availableCollateral = sdk.NewInt(1)
	expensiveCollateralDustDown.repayTokenPrice = sdk.MustNewDecFromStr("2")
	expensiveCollateralDustDown.rewardTokenPrice = sdk.MustNewDecFromStr("39.9")
	expensiveCollateralDustDown.liquidationIncentive = sdk.MustNewDecFromStr("0")
	runTestCase(expensiveCollateralDustDown, 20, 1, 1, "expensive collateral dust with price down")

	// TODO: more rounding and dust scenarios.
	// borrowed/collat/reward high/low price and maybe high/low utoken
}
