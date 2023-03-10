package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v4/x/leverage/keeper"
)

func TestComputeLiquidation(t *testing.T) {
	type testCase struct {
		availableRepay       sdkmath.Int
		availableCollateral  sdkmath.Int
		availableReward      sdkmath.Int
		repayTokenPrice      sdk.Dec
		rewardTokenPrice     sdk.Dec
		uTokenExchangeRate   sdk.Dec
		liquidationIncentive sdk.Dec
	}

	baseCase := func() testCase {
		return testCase{
			sdkmath.NewInt(1000),         // 1000 Token A to repay
			sdkmath.NewInt(5000),         // 5000 uToken B collateral
			sdkmath.NewInt(5000),         // 5000 Token B liquidity
			sdk.OneDec(),                 // price(A) = $1
			sdk.OneDec(),                 // price(B) = $1
			sdk.OneDec(),                 // utoken exchange rate 1 u/B => 1 B
			sdk.MustNewDecFromStr("0.1"), // reward value is 110% repay value
		}
	}

	runTestCase := func(tc testCase, expectedRepay, expectedCollateral, expectedReward int64, msg string) {
		priceRatio := tc.repayTokenPrice.Quo(tc.rewardTokenPrice)
		repay, collateral, reward := keeper.ComputeLiquidation(
			tc.availableRepay,
			tc.availableCollateral,
			tc.availableReward,
			priceRatio,
			tc.uTokenExchangeRate,
			tc.liquidationIncentive,
		)

		assert.Equal(t, true, sdkmath.NewInt(expectedRepay).Equal(repay),
			msg+" (repay); expected: %d, got: %s", expectedRepay, repay)
		assert.Equal(t, true, sdkmath.NewInt(expectedCollateral).Equal(collateral),
			msg+" (collateral); expected: %d, got: %s", expectedCollateral, collateral)
		assert.Equal(t, true, sdkmath.NewInt(expectedReward).Equal(reward), msg+" (reward); got: %d, expected: %s", expectedReward, reward)
	}

	// basic liquidation of 1000 borrowed tokens with plenty of available rewards and collateral
	runTestCase(baseCase(), 1000, 1100, 1100, "base case")

	// borrower is healthy (zero max repay would result from close factor of zero) so liquidation cannot occur
	healthyCase := baseCase()
	healthyCase.availableRepay = sdk.ZeroInt()
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

	// complex case, limited by available repay, with various nontrivial values
	complexCase := baseCase()
	complexCase.availableRepay = sdkmath.NewInt(300)
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
	expensiveBorrowDust.availableRepay = sdkmath.NewInt(1)
	expensiveBorrowDust.repayTokenPrice = sdk.MustNewDecFromStr("40")
	expensiveBorrowDust.rewardTokenPrice = sdk.MustNewDecFromStr("2")
	expensiveBorrowDust.liquidationIncentive = sdk.MustNewDecFromStr("0")
	runTestCase(expensiveBorrowDust, 1, 20, 20, "expensive borrow dust")

	// borrow dust case, with high borrowed token value rounds reward down
	expensiveBorrowDustDown := baseCase()
	expensiveBorrowDustDown.availableRepay = sdkmath.NewInt(1)
	expensiveBorrowDustDown.repayTokenPrice = sdk.MustNewDecFromStr("39.9")
	expensiveBorrowDustDown.rewardTokenPrice = sdk.MustNewDecFromStr("2")
	expensiveBorrowDustDown.liquidationIncentive = sdk.MustNewDecFromStr("0")
	runTestCase(expensiveBorrowDustDown, 1, 19, 19, "expensive borrow dust with price down")

	// borrow dust case, with high borrowed token value rounds collateral burn up
	expensiveBorrowDustUp := baseCase()
	expensiveBorrowDustUp.availableRepay = sdkmath.NewInt(1)
	expensiveBorrowDustUp.repayTokenPrice = sdk.MustNewDecFromStr("40.1")
	expensiveBorrowDustUp.rewardTokenPrice = sdk.MustNewDecFromStr("2")
	expensiveBorrowDustUp.liquidationIncentive = sdk.MustNewDecFromStr("0")
	runTestCase(expensiveBorrowDustUp, 1, 20, 20, "expensive borrow dust with price up")

	// borrow dust case, with low borrowed token value rounds collateral burn and reward to zero
	cheapBorrowDust := baseCase()
	cheapBorrowDust.availableRepay = sdkmath.NewInt(1)
	cheapBorrowDust.repayTokenPrice = sdk.MustNewDecFromStr("2")
	cheapBorrowDust.rewardTokenPrice = sdk.MustNewDecFromStr("40")
	cheapBorrowDust.liquidationIncentive = sdk.MustNewDecFromStr("0")
	runTestCase(cheapBorrowDust, 1, 0, 0, "cheap borrow dust")

	// collateral dust case, with high collateral token value and no rounding
	expensiveCollateralDust := baseCase()
	expensiveCollateralDust.availableCollateral = sdkmath.NewInt(1)
	expensiveCollateralDust.repayTokenPrice = sdk.MustNewDecFromStr("2")
	expensiveCollateralDust.rewardTokenPrice = sdk.MustNewDecFromStr("40")
	expensiveCollateralDust.liquidationIncentive = sdk.MustNewDecFromStr("0")
	runTestCase(expensiveCollateralDust, 20, 1, 1, "expensive collateral dust")

	// collateral dust case, with high collateral token value rounds required repayment up
	expensiveCollateralDustUp := baseCase()
	expensiveCollateralDustUp.availableCollateral = sdkmath.NewInt(1)
	expensiveCollateralDustUp.repayTokenPrice = sdk.MustNewDecFromStr("2")
	expensiveCollateralDustUp.rewardTokenPrice = sdk.MustNewDecFromStr("40.1")
	expensiveCollateralDustUp.liquidationIncentive = sdk.MustNewDecFromStr("0")
	runTestCase(expensiveCollateralDustUp, 21, 0, 0, "expensive collateral dust with price up")

	// collateral dust case, with high collateral token value rounds required repayment up
	expensiveCollateralDustDown := baseCase()
	expensiveCollateralDustDown.availableCollateral = sdkmath.NewInt(1)
	expensiveCollateralDustDown.repayTokenPrice = sdk.MustNewDecFromStr("2")
	expensiveCollateralDustDown.rewardTokenPrice = sdk.MustNewDecFromStr("39.9")
	expensiveCollateralDustDown.liquidationIncentive = sdk.MustNewDecFromStr("0")
	runTestCase(expensiveCollateralDustDown, 20, 0, 0, "expensive collateral dust with price down")

	// collateral dust case, with low collateral token value rounds required repayment up
	cheapCollateralDust := baseCase()
	cheapCollateralDust.availableCollateral = sdkmath.NewInt(1)
	cheapCollateralDust.repayTokenPrice = sdk.MustNewDecFromStr("40")
	cheapCollateralDust.rewardTokenPrice = sdk.MustNewDecFromStr("2")
	cheapCollateralDust.liquidationIncentive = sdk.MustNewDecFromStr("0")
	runTestCase(cheapCollateralDust, 1, 1, 1, "cheap collateral dust")

	// exotic case with cheap collateral base tokens but a very high uToken exchange rate
	// rounds required repayment up and base reward down
	uDust := baseCase()
	uDust.availableCollateral = sdk.NewInt(1)
	uDust.repayTokenPrice = sdk.MustNewDecFromStr("40")
	uDust.rewardTokenPrice = sdk.MustNewDecFromStr("2")
	uDust.uTokenExchangeRate = sdk.MustNewDecFromStr("29.5")
	uDust.liquidationIncentive = sdk.MustNewDecFromStr("0")
	runTestCase(uDust, 2, 1, 29, "high exchange rate collateral dust")
}

func TestCloseFactor(t *testing.T) {
	type testCase struct {
		borrowedValue                sdk.Dec
		collateralValue              sdk.Dec
		liquidationThreshold         sdk.Dec
		smallLiquidationSize         sdk.Dec
		minimumCloseFactor           sdk.Dec
		completeLiquidationThreshold sdk.Dec
	}

	baseCase := func() testCase {
		// returns a liquidation scenario where close factor will reach 1 at a borrowed value of 58
		return testCase{
			sdk.MustNewDecFromStr("50"),  // borrowed value 50
			sdk.MustNewDecFromStr("100"), // collateral value 100
			sdk.MustNewDecFromStr("40"),  // liquidation threshold 40
			sdk.MustNewDecFromStr("20"),  // small liquidation size 20
			sdk.MustNewDecFromStr("0.1"), // minimum close factor 10%
			sdk.MustNewDecFromStr("0.3"), // complete liquidation threshold 30%
		}
	}

	runTestCase := func(tc testCase, expectedCloseFactor string, msg string) {
		t.Run(msg, func(t *testing.T) {
			closeFactor := keeper.ComputeCloseFactor(
				tc.borrowedValue,
				tc.collateralValue,
				tc.liquidationThreshold,
				tc.smallLiquidationSize,
				tc.minimumCloseFactor,
				tc.completeLiquidationThreshold,
			)

			assert.DeepEqual(t, sdk.MustNewDecFromStr(expectedCloseFactor), closeFactor)
		})
	}

	// In the base case, close factor scales from 10% to 100% as borrowed value
	// goes from liquidation threshold ($40) to a critical value, which is defined
	// to be 30% of the way between liquidation threshold and collateral value ($100).
	// Since the borrowed value of $50 is 5/9 the way from liquidation threshold to
	// the base case's critical value of $58, the computed close factor will be
	// 5/9 of the way from 10% to 100% - thus, 60%.
	runTestCase(baseCase(), "0.6", "base case")

	// If borrowed value has passed the critical point, close factor is 1
	completeLiquidation := baseCase()
	completeLiquidation.borrowedValue = sdk.MustNewDecFromStr("60")
	runTestCase(completeLiquidation, "1", "complete liquidation")

	// If borrowed value is less than small liquidation size, close factor is 1.
	smallLiquidation := baseCase()
	smallLiquidation.smallLiquidationSize = sdk.MustNewDecFromStr("60")
	runTestCase(smallLiquidation, "1", "small liquidation")

	// A liquidation-ineligible target would not have its close factor calculated in
	// practice, but the function should return zero if it were.
	notEligible := baseCase()
	notEligible.borrowedValue = sdk.MustNewDecFromStr("30")
	runTestCase(notEligible, "0", "liquidation ineligible")

	// A liquidation-ineligible target which is below the small liquidation size
	// should still return a close factor of zero.
	smallNotEligible := baseCase()
	smallNotEligible.borrowedValue = sdk.MustNewDecFromStr("10")
	runTestCase(smallNotEligible, "0", "liquidation ineligible (small)")

	// A borrower which is exactly on their liquidation threshold will have a close factor
	// equal to minimumCloseFactor.
	exactThreshold := baseCase()
	exactThreshold.borrowedValue = sdk.MustNewDecFromStr("40")
	runTestCase(exactThreshold, "0.1", "exact threshold")

	// If collateral weights are all 1 (CV = LT), close factor will be MinimumCloseFactor.
	// This situation will not occur in practice as collateral weights are less than one.
	highCollateralWeight := baseCase()
	highCollateralWeight.collateralValue = sdk.MustNewDecFromStr("40")
	runTestCase(highCollateralWeight, "0.1", "high collateral weights")
}
