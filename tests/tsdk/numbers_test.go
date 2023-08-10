package tsdk

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestDecF(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(sdk.MustNewDecFromStr("10.20"), DecF(10.2))
	assert.Equal(sdk.MustNewDecFromStr("0.002"), DecF(0.002))
	assert.Equal(sdk.MustNewDecFromStr("0"), DecF(0))
	assert.Equal(sdk.MustNewDecFromStr("-1.9"), DecF(-1.9))
}
