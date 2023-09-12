package v2_test

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v6/x/ugov"
	"github.com/umee-network/umee/v6/x/ugov/keeper/intest"
	v2 "github.com/umee-network/umee/v6/x/ugov/migrations/v2"
)

func TestMigrateStore(t *testing.T) {
	_, k := intest.MkKeeper(t)

	// before migration
	p := k.InflationParams()
	assert.Equal(t, ugov.InflationParams{}, p)
	cycleEnd := k.InflationCycleEnd()
	assert.Equal(t, time.UnixMilli(0), cycleEnd)

	// after migration
	err := v2.MigrateStore(k)
	assert.NilError(t, err)
	p = k.InflationParams()
	assert.DeepEqual(t, ugov.DefaultInflationParams(), p)

	cycleEnd = k.InflationCycleEnd()
	assert.DeepEqual(t, time.Date(2023, time.October, 15, 0, 0, 0, 0, time.UTC), cycleEnd)
}
