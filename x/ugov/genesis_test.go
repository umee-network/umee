package ugov

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v6/util/coin"
)

func TestGenesis(t *testing.T) {
	require := require.New(t)
	gs := DefaultGenesis()

	require.NoError(gs.Validate(), "default genesis must be correct")

	gs.MinGasPrice = coin.UtokenDec(coin.Dollar, "0.1")
	require.NoError(gs.Validate(), "min_gas_price = 0.1dollar is correct")

	gs.MinGasPrice.Amount = sdkmath.LegacyMustNewDecFromStr("-1")
	require.Error(gs.Validate(), "negative min_gas_price is NOT correct")
}
