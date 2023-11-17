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
