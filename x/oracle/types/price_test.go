package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"
)

func TestPrices(t *testing.T) {
	zero := sdk.ZeroDec()
	p1 := NewPrice(zero, "atom", 4)
	p2 := NewPrice(zero, "atom", 3)
	prices := Prices{p1, p2}
	prices.Sort()
	assert.DeepEqual(t, Prices{p2, p1}, prices)
}
