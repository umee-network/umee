package fixtures

import (
	sdkmath "cosmossdk.io/math"

	"github.com/umee-network/umee/v6/x/leverage/types"
)

// Params returns leverage params used in testing.
func Params() types.Params {
	return types.Params{
		CompleteLiquidationThreshold: sdkmath.LegacyMustNewDecFromStr("0.1"),
		MinimumCloseFactor:           sdkmath.LegacyMustNewDecFromStr("0.01"),
		OracleRewardFactor:           sdkmath.LegacyMustNewDecFromStr("0.01"),
		RewardsAuctionFee:            sdkmath.LegacyMustNewDecFromStr("0.02"),
		SmallLiquidationSize:         sdkmath.LegacyMustNewDecFromStr("100.00"),
		DirectLiquidationFee:         sdkmath.LegacyMustNewDecFromStr("0.1"),
	}
}
