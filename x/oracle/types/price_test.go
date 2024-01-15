package types

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestPrices(t *testing.T) {
	zero := sdkmath.LegacyZeroDec()
	p1 := NewPrice(zero, "atom", 4)
	p2 := NewPrice(zero, "atom", 3)
	prices := Prices{p1, p2}
	prices.Sort()
	assert.DeepEqual(t, Prices{p2, p1}, prices)
}
