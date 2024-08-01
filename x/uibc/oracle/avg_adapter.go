package oracle

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/x/uibc"
)

type UmeeAvgPriceOracle interface {
	HistoricAvgPrice(ctx sdk.Context, denom string) (sdkmath.LegacyDec, error)
}

func FromUmeeAvgPriceOracle(o UmeeAvgPriceOracle) uibc.Oracle {
	return umeeAvgPriceOracle{o}
}

type umeeAvgPriceOracle struct {
	o UmeeAvgPriceOracle
}

func (o umeeAvgPriceOracle) Price(ctx sdk.Context, denom string) (sdkmath.LegacyDec, error) {
	return o.o.HistoricAvgPrice(ctx, denom)
}
