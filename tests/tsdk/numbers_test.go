package tsdk

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/assert"
)

func TestDecF(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(sdkmath.LegacyMustNewDecFromStr("10.20"), DecF(10.2))
	assert.Equal(sdkmath.LegacyMustNewDecFromStr("0.002"), DecF(0.002))
	assert.Equal(sdkmath.LegacyMustNewDecFromStr("0"), DecF(0))
	assert.Equal(sdkmath.LegacyMustNewDecFromStr("-1.9"), DecF(-1.9))
}
