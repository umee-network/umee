package v1

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/price-feeder/v2/oracle"
)

// Oracle defines the Oracle interface contract that the v1 router depends on.
type Oracle interface {
	GetLastPriceSyncTimestamp() time.Time
	GetPrices() map[string]sdk.Dec
	GetTvwapPrices() oracle.PricesByProvider
	GetVwapPrices() oracle.PricesByProvider
}
