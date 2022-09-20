package v1

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/price-feeder/oracle/provider"
)

// Oracle defines the Oracle interface contract that the v1 router depends on.
type Oracle interface {
	GetLastPriceSyncTimestamp() time.Time
	GetPrices() map[string]sdk.Dec
	GetTvwapPrices() map[provider.Name]map[string]sdk.Dec
	GetVwapPrices() map[provider.Name]map[string]sdk.Dec
}
