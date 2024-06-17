package checkers

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestCoinPositive(t *testing.T) {
	assert := assert.New(t)

	c0 := sdk.NewInt64Coin("abc", 0)
	cNeg := sdk.NewInt64Coin("abc", 0)
	cPos := sdk.NewInt64Coin("abc", 1)

	assert.Empty(PositiveCoins(""))
	assert.Empty(PositiveCoins("", cPos))
	assert.Empty(PositiveCoins("", cPos, cPos))

	assert.Contains(errsToStr(PositiveCoins("", c0)), "coin[0]")
	assert.Contains(errsToStr(PositiveCoins("", cNeg)), "coin[0]")
	assert.Contains(errsToStr(PositiveCoins("", cPos, c0)), "coin[1]")
	assert.Contains(errsToStr(PositiveCoins("", cPos, cNeg)), "coin[1]")
	assert.NotContains(errsToStr(PositiveCoins("", cPos, cNeg)), "coin[0]")

	assert.Contains(errsToStr(PositiveCoins("", cPos, c0, cNeg)), "coin[1]")
	assert.Contains(errsToStr(PositiveCoins("", cPos, c0, cNeg)), "coin[2]")
	assert.NotContains(errsToStr(PositiveCoins("", cPos, c0, cNeg)), "coin[0]")
}
