package v2

import (
	"time"

	"github.com/umee-network/umee/v6/x/ugov"
)

// MigrateStore performs in-place store migrations from 1 to 2
func MigrateStore(k ugov.Keeper) error {
	ip := ugov.DefaultInflationParams()
	if err := k.SetInflationParams(ip); err != nil {
		return err
	}

	cycleEnd := time.Date(2023, time.October, 15, 0, 0, 0, 0, time.UTC)
	return k.SetInflationCycleEnd(cycleEnd)
}
