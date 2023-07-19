package ugov

import (
	fmt "fmt"
	time "time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/umee-network/umee/v5/app/params"
	"github.com/umee-network/umee/v5/util/coin"
)

func DefaultInflationParams() InflationParams {
	return InflationParams{
		MaxSupply:              coin.New(appparams.BondDenom, 21_000000000), // 21 Billition Maximum for Staking Bonding Denom
		InflationCycle:         time.Hour * 24 * 365,                        // 2 years for default inflation cycle
		InflationReductionRate: sdk.MustNewDecFromStr("0.25"),               // 25% reduction rate for inflation cyle
	}
}

func (lp InflationParams) Validate() error {
	if lp.MaxSupply.Amount.LT(math.NewInt(0)) {
		return fmt.Errorf("max_supply must be not negative")
	}

	if lp.InflationReductionRate.LT(sdk.ZeroDec()) {
		return fmt.Errorf("inflation reduction rate must be not negative")
	}

	if lp.InflationCycle.Seconds() <= 0 {
		return fmt.Errorf("inflation cycle must be positive")
	}

	return nil
}
