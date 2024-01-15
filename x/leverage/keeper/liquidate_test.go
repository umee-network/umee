package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v6/x/leverage/keeper"
)

func TestComputeLiquidation(t *testing.T) {
	type testCase struct {
		availableRepay       sdkmath.Int
		availableCollateral  sdkmath.Int
		availableReward      sdkmath.Int
		repayTokenPrice      sdkmath.LegacyDec
		rewardTokenPrice     sdkmath.LegacyDec
		uTokenExchangeRate   sdkmath.LegacyDec
		liquidationIncentive sdkmath.LegacyDec
		leveragedLiquidate   bool
	}

	baseCase := func() testCase {
		return testCase{
			sdkmath.NewInt(1000),                   // 1000 Token A to repay
			sdkmath.NewInt(5000),                   // 5000 uToken B collateral
			sdkmath.NewInt(5000),                   // 5000 Token B liquidity
			sdkmath.LegacyOneDec(),                 // price(A) = $1
			sdkmath.LegacyOneDec(),                 // price(B) = $1
			sdkmath.LegacyOneDec(),                 // utoken exchange rate 1 u/B => 1 B
			sdkmath.LegacyMustNewDecFromStr("0.1"), // reward value is 110% repay value
			false,
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
			tc.leveragedLiquidate,
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
	healthyCase.availableRepay = sdkmath.ZeroInt()
	runTestCase(healthyCase, 0, 0, 0, "healthy borrower")

	// limiting factor is available repay
	repayLimited := baseCase()
	repayLimited.availableRepay = sdkmath.NewInt(100)
	runTestCase(repayLimited, 100, 110, 110, "repay limited")

	// limiting factor is available collateral
	collateralLimited := baseCase()
	collateralLimited.availableCollateral = sdkmath.NewInt(220)
	runTestCase(collateralLimited, 200, 220, 220, "collateral limited")

	// limiting factor is available reward
	rewardLimited := baseCase()
	rewardLimited.availableReward = sdkmath.NewInt(330)
	runTestCase(rewardLimited, 300, 330, 330, "reward limited")

	// limiting factor would be available reward, but leveraged liquidation is not limited by base tokens
	rewardNotLimited := baseCase()
	rewardNotLimited.availableReward = sdkmath.NewInt(330)
	rewardNotLimited.leveragedLiquidate = true
	runTestCase(rewardNotLimited, 1000, 1100, 1100, "reward not limited")

	// repay token is worth more
	expensiveRepay := baseCase()
	expensiveRepay.repayTokenPrice = sdkmath.LegacyMustNewDecFromStr("2")
	runTestCase(expensiveRepay, 1000, 2200, 2200, "expensive repay")

	// reward token is worth more
	expensiveReward := baseCase()
	expensiveReward.rewardTokenPrice = sdkmath.LegacyMustNewDecFromStr("2")
	runTestCase(expensiveReward, 1000, 550, 550, "expensive reward")

	// high collateral uToken exchange rate
	exchangeRate := baseCase()
	exchangeRate.uTokenExchangeRate = sdkmath.LegacyMustNewDecFromStr("2")
	runTestCase(exchangeRate, 1000, 550, 1100, "high uToken exchange rate")

	// high liquidation incentive
	highIncentive := baseCase()
	highIncentive.liquidationIncentive = sdkmath.LegacyMustNewDecFromStr("1.5")
	runTestCase(highIncentive, 1000, 2500, 2500, "high liquidation incentive")

	// no liquidation incentive
	noIncentive := baseCase()
	noIncentive.liquidationIncentive = sdkmath.LegacyZeroDec()
	runTestCase(noIncentive, 1000, 1000, 1000, "no liquidation incentive")

	// complex case, limited by available repay, with various nontrivial values
	complexCase := baseCase()
	complexCase.availableRepay = sdkmath.NewInt(300)
	complexCase.uTokenExchangeRate = sdkmath.LegacyMustNewDecFromStr("2.5")
	complexCase.liquidationIncentive = sdkmath.LegacyMustNewDecFromStr("0.5")
	complexCase.repayTokenPrice = sdkmath.LegacyMustNewDecFromStr("6")
	complexCase.rewardTokenPrice = sdkmath.LegacyMustNewDecFromStr("12")
	// repay = 300 (limiting factor)
	// collateral = 300 * 1.5 * (6/12) / 2.5 = 0.3 * 300 = 90
	// reward = 300 * 1.5 * (6/12) = 225
	runTestCase(complexCase, 300, 90, 225, "complex case")

	// borrow dust case, with high borrowed token value and no rounding
	expensiveBorrowDust := baseCase()
	expensiveBorrowDust.availableRepay = sdkmath.NewInt(1)
	expensiveBorrowDust.repayTokenPrice = sdkmath.LegacyMustNewDecFromStr("40")
	expensiveBorrowDust.rewardTokenPrice = sdkmath.LegacyMustNewDecFromStr("2")
	expensiveBorrowDust.liquidationIncentive = sdkmath.LegacyMustNewDecFromStr("0")
	runTestCase(expensiveBorrowDust, 1, 20, 20, "expensive borrow dust")

	// borrow dust case, with high borrowed token value rounds reward down
	expensiveBorrowDustDown := baseCase()
	expensiveBorrowDustDown.availableRepay = sdkmath.NewInt(1)
	expensiveBorrowDustDown.repayTokenPrice = sdkmath.LegacyMustNewDecFromStr("39.9")
	expensiveBorrowDustDown.rewardTokenPrice = sdkmath.LegacyMustNewDecFromStr("2")
	expensiveBorrowDustDown.liquidationIncentive = sdkmath.LegacyMustNewDecFromStr("0")
	runTestCase(expensiveBorrowDustDown, 1, 20, 20, "expensive borrow dust with price down")

	// borrow dust case, with high borrowed token value rounds collateral burn up
	expensiveBorrowDustUp := baseCase()
	expensiveBorrowDustUp.availableRepay = sdkmath.NewInt(1)
	expensiveBorrowDustUp.repayTokenPrice = sdkmath.LegacyMustNewDecFromStr("40.1")
	expensiveBorrowDustUp.rewardTokenPrice = sdkmath.LegacyMustNewDecFromStr("2")
	expensiveBorrowDustUp.liquidationIncentive = sdkmath.LegacyMustNewDecFromStr("0")
	runTestCase(expensiveBorrowDustUp, 1, 21, 21, "expensive borrow dust with price up")

	// borrow dust case, with low borrowed token value rounds collateral burn and reward to zero
	cheapBorrowDust := baseCase()
	cheapBorrowDust.availableRepay = sdkmath.NewInt(1)
	cheapBorrowDust.repayTokenPrice = sdkmath.LegacyMustNewDecFromStr("2")
	cheapBorrowDust.rewardTokenPrice = sdkmath.LegacyMustNewDecFromStr("40")
	cheapBorrowDust.liquidationIncentive = sdkmath.LegacyMustNewDecFromStr("0")
	runTestCase(cheapBorrowDust, 1, 1, 1, "cheap borrow dust")

	// collateral dust case, with high collateral token value and no rounding
	expensiveCollateralDust := baseCase()
	expensiveCollateralDust.availableCollateral = sdkmath.NewInt(1)
	expensiveCollateralDust.repayTokenPrice = sdkmath.LegacyMustNewDecFromStr("2")
	expensiveCollateralDust.rewardTokenPrice = sdkmath.LegacyMustNewDecFromStr("40")
	expensiveCollateralDust.liquidationIncentive = sdkmath.LegacyMustNewDecFromStr("0")
	runTestCase(expensiveCollateralDust, 20, 1, 1, "expensive collateral dust")

	// collateral dust case, with high collateral token value rounds required repayment up
	expensiveCollateralDustUp := baseCase()
	expensiveCollateralDustUp.availableCollateral = sdkmath.NewInt(1)
	expensiveCollateralDustUp.repayTokenPrice = sdkmath.LegacyMustNewDecFromStr("2")
	expensiveCollateralDustUp.rewardTokenPrice = sdkmath.LegacyMustNewDecFromStr("40.1")
	expensiveCollateralDustUp.liquidationIncentive = sdkmath.LegacyMustNewDecFromStr("0")
	runTestCase(expensiveCollateralDustUp, 21, 1, 1, "expensive collateral dust with price up")

	// collateral dust case, with high collateral token value rounds required repayment up
	expensiveCollateralDustDown := baseCase()
	expensiveCollateralDustDown.availableCollateral = sdkmath.NewInt(1)
	expensiveCollateralDustDown.repayTokenPrice = sdkmath.LegacyMustNewDecFromStr("2")
	expensiveCollateralDustDown.rewardTokenPrice = sdkmath.LegacyMustNewDecFromStr("39.9")
	expensiveCollateralDustDown.liquidationIncentive = sdkmath.LegacyMustNewDecFromStr("0")
	runTestCase(expensiveCollateralDustDown, 20, 1, 1, "expensive collateral dust with price down")

	// collateral dust case, with low collateral token value rounds required repayment up
	cheapCollateralDust := baseCase()
	cheapCollateralDust.availableCollateral = sdkmath.NewInt(1)
	cheapCollateralDust.repayTokenPrice = sdkmath.LegacyMustNewDecFromStr("40")
	cheapCollateralDust.rewardTokenPrice = sdkmath.LegacyMustNewDecFromStr("2")
	cheapCollateralDust.liquidationIncentive = sdkmath.LegacyMustNewDecFromStr("0")
	runTestCase(cheapCollateralDust, 1, 1, 1, "cheap collateral dust")

	// exotic case with cheap collateral base tokens but a very high uToken exchange rate
	// rounds required repayment up and base reward down
	uDust := baseCase()
	uDust.availableCollateral = sdkmath.NewInt(1)
	uDust.repayTokenPrice = sdkmath.LegacyMustNewDecFromStr("40")
	uDust.rewardTokenPrice = sdkmath.LegacyMustNewDecFromStr("2")
	uDust.uTokenExchangeRate = sdkmath.LegacyMustNewDecFromStr("29.5")
	uDust.liquidationIncentive = sdkmath.LegacyMustNewDecFromStr("0")
	runTestCase(uDust, 2, 1, 29, "high exchange rate collateral dust")
}

func TestCloseFactor(t *testing.T) {
	type testCase struct {
		borrowedValue                sdkmath.LegacyDec
		collateralValue              sdkmath.LegacyDec
		liquidationThreshold         sdkmath.LegacyDec
		smallLiquidationSize         sdkmath.LegacyDec
		minimumCloseFactor           sdkmath.LegacyDec
		completeLiquidationThreshold sdkmath.LegacyDec
	}

	baseCase := func() testCase {
		// returns a liquidation scenario where close factor will reach 1 at a borrowed value of 58
		return testCase{
			sdkmath.LegacyMustNewDecFromStr("50"),  // borrowed value 50
			sdkmath.LegacyMustNewDecFromStr("100"), // collateral value 100
			sdkmath.LegacyMustNewDecFromStr("40"),  // liquidation threshold 40
			sdkmath.LegacyMustNewDecFromStr("20"),  // small liquidation size 20
			sdkmath.LegacyMustNewDecFromStr("0.1"), // minimum close factor 10%
			sdkmath.LegacyMustNewDecFromStr("0.3"), // complete liquidation threshold 30%
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

			assert.DeepEqual(t, sdkmath.LegacyMustNewDecFromStr(expectedCloseFactor), closeFactor)
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
	completeLiquidation.borrowedValue = sdkmath.LegacyMustNewDecFromStr("60")
	runTestCase(completeLiquidation, "1", "complete liquidation")

	// If borrowed value is less than small liquidation size, close factor is 1.
	smallLiquidation := baseCase()
	smallLiquidation.smallLiquidationSize = sdkmath.LegacyMustNewDecFromStr("60")
	runTestCase(smallLiquidation, "1", "small liquidation")

	// A liquidation-ineligible target would not have its close factor calculated in
	// practice, but the function should return zero if it were.
	notEligible := baseCase()
	notEligible.borrowedValue = sdkmath.LegacyMustNewDecFromStr("30")
	runTestCase(notEligible, "0", "liquidation ineligible")

	// A liquidation-ineligible target which is below the small liquidation size
	// should still return a close factor of zero.
	smallNotEligible := baseCase()
	smallNotEligible.borrowedValue = sdkmath.LegacyMustNewDecFromStr("10")
	runTestCase(smallNotEligible, "0", "liquidation ineligible (small)")

	// A borrower which is exactly on their liquidation threshold will have a close factor
	// equal to minimumCloseFactor.
	exactThreshold := baseCase()
	exactThreshold.borrowedValue = sdkmath.LegacyMustNewDecFromStr("40")
	runTestCase(exactThreshold, "0.1", "exact threshold")

	// If collateral weights are all 1 (CV = LT), close factor will be MinimumCloseFactor.
	// This situation will not occur in practice as collateral weights are less than one.
	highCollateralWeight := baseCase()
	highCollateralWeight.collateralValue = sdkmath.LegacyMustNewDecFromStr("40")
	runTestCase(highCollateralWeight, "0.1", "high collateral weights")
}
