package params

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestDenoms(t *testing.T) {
	assert := assert.New(t)

	assert.NoError(sdk.ValidateDenom(BondDenom))
	assert.NoError(sdk.ValidateDenom(BaseExtraDenom))
	assert.NoError(sdk.ValidateDenom(DisplayDenom))
	assert.NoError(sdk.ValidateDenom(LegacyDisplayDenom))
}

func TestUXMetadata(t *testing.T) {
	um := UmeeTokenMetadata()
	assert := assert.New(t)
	assert.Equal("uumee", um.Base, "Umee base denom must not change")
	assert.Equal("UX", um.Name)
	assert.Equal("UX", um.DenomUnits[1].Denom)
}
