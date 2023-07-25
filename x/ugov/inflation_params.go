package ugov

import (
	fmt "fmt"
	time "time"

	"cosmossdk.io/math"

	appparams "github.com/umee-network/umee/v5/app/params"
	"github.com/umee-network/umee/v5/util/bpmath"
	"github.com/umee-network/umee/v5/util/coin"
)

func DefaultInflationParams() InflationParams {
	return InflationParams{
		MaxSupply:              coin.New(appparams.BondDenom, 21_000000000_000000), // 21 Billion Maximum
		InflationCycle:         time.Hour * 24 * 365,                               // 2 years for default inflation cycle
		InflationReductionRate: bpmath.FixedBP(2500),                               // 25% reduction rate for inflation cyle
	}
}

func (ip InflationParams) Validate() error {
	if ip.MaxSupply.Amount.LT(math.NewInt(0)) {
		return fmt.Errorf("max_supply must be positive")
	}

	if ip.InflationReductionRate > bpmath.One || ip.InflationReductionRate < 100 {
		return fmt.Errorf("inflation reduction must be between 100(0.1) to 10000 (1)")
	}

	if ip.InflationCycle.Seconds() <= 0 {
		return fmt.Errorf("inflation cycle must be positive")
	}

	return nil
}
