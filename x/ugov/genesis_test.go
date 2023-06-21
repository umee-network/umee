package ugov

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v5/util/coin"
)

func TestGenesis(t *testing.T) {
	require := require.New(t)
	gs := DefaultGenesis()

	require.NoError(gs.Validate(), "default genesis must be correct")

	gs.MinGasPrice = coin.UtokenDec(coin.Dollar, "0.1")
	require.NoError(gs.Validate(), "min_gas_price = 0.1dollar is correct")

	gs.MinGasPrice.Amount = sdk.MustNewDecFromStr("-1")
	require.Error(gs.Validate(), "negative min_gas_price is NOT correct")
}
