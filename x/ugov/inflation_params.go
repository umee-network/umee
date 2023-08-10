package ugov

import (
	time "time"

	"cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	appparams "github.com/umee-network/umee/v6/app/params"
	"github.com/umee-network/umee/v6/util/bpmath"
	"github.com/umee-network/umee/v6/util/coin"
)

func DefaultInflationParams() InflationParams {
	return InflationParams{
		MaxSupply:              coin.New(appparams.BondDenom, 21_000000000_000000), // 21 Billion Maximum
		InflationCycle:         time.Hour * 24 * 365 * 2,                           // 2 years for default inflation cycle
		InflationReductionRate: bpmath.FixedBP(2500),                               // 25% reduction rate for inflation cyle
	}
}

var zeroInt = math.NewInt(0)

func (ip InflationParams) Validate() error {
	if ip.MaxSupply.Amount.LT(zeroInt) {
		return sdkerrors.ErrInvalidRequest.Wrap("max_supply must be positive")
	}

	if ip.InflationReductionRate > bpmath.One || ip.InflationReductionRate < 100 {
		return sdkerrors.ErrInvalidRequest.Wrap("inflation reduction must be between 100bp to 10'000bp")
	}

	if ip.InflationCycle.Seconds() <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("inflation cycle must be positive")
	}

	return nil
}
