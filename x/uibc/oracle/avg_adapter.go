package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/x/uibc"
)

type UmeeAvgPriceOracle interface {
	HistoricAvgPrice(ctx sdk.Context, denom string) (sdk.Dec, error)
}

func FromUmeeAvgPriceOracle(o UmeeAvgPriceOracle) uibc.Oracle {
	return umeeAvgPriceOracle{o}
}

type umeeAvgPriceOracle struct {
	o UmeeAvgPriceOracle
}

func (o umeeAvgPriceOracle) Price(ctx sdk.Context, denom string) (sdk.Dec, error) {
	return o.o.HistoricAvgPrice(ctx, denom)
}
