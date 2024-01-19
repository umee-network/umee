package uics20

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/umee-network/umee/v6/tests/accs"
	"github.com/umee-network/umee/v6/util/coin"
	ltypes "github.com/umee-network/umee/v6/x/leverage/types"
)

func TestMemoSignerCheck(t *testing.T) {
	assert := assert.New(t)
	sender := accs.Alice
	wrongSignerErr := "signer doesn't match the sender"
	asset := coin.New("atom", 10)
	im := ICS20Module{}
	tcs := []struct {
		msg    sdk.Msg
		errstr string
	}{
		{ltypes.NewMsgSupply(accs.Bob, asset), wrongSignerErr},
		{ltypes.NewMsgSupplyCollateral(accs.Bob, asset), wrongSignerErr},
		{ltypes.NewMsgBorrow(accs.Bob, asset), wrongSignerErr},
	}

	for _, tc := range tcs {
		err := im.handleMemoMsg(nil, sender, tc.msg)
		if tc.errstr != "" {
			assert.ErrorContains(err, tc.errstr)
		} else {
			assert.NoError(err)
		}
	}
}
