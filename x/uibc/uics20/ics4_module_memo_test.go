package uics20

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/umee-network/umee/v6/tests/accs"
	"github.com/umee-network/umee/v6/util/coin"
	ltypes "github.com/umee-network/umee/v6/x/leverage/types"
	"github.com/umee-network/umee/v6/x/uibc/mocks"
)

func TestMemoSignerCheck(t *testing.T) {
	assert := assert.New(t)
	receiver := accs.Alice
	wrongSignerErr := "signer doesn't match the receiver"
	asset := coin.New("atom", 10)
	sent := coin.New("atom", 10)
	im := ICS20Module{leverage: mocks.NewLvgNoopMsgSrv()}
	// sdkCtx := sdk.Context{}
	tcs := []struct {
		msgs   []sdk.Msg
		errstr string
	}{
		{[]sdk.Msg{ltypes.NewMsgSupply(accs.Bob, asset)}, wrongSignerErr},
		{[]sdk.Msg{ltypes.NewMsgSupplyCollateral(accs.Bob, asset)}, wrongSignerErr},

		{[]sdk.Msg{ltypes.NewMsgSupply(receiver, asset)}, ""},
		{[]sdk.Msg{ltypes.NewMsgSupplyCollateral(receiver, asset)}, ""},

		{[]sdk.Msg{ltypes.NewMsgSupplyCollateral(receiver, asset),
			ltypes.NewMsgBorrow(accs.Bob, asset)}, wrongSignerErr},
		{[]sdk.Msg{ltypes.NewMsgSupplyCollateral(receiver, asset),
			ltypes.NewMsgBorrow(receiver, asset)}, ""},

		{
			[]sdk.Msg{ltypes.NewMsgDecollateralize(receiver, asset)},
			"are supported as messages[0]",
		},
	}

	for i, tc := range tcs {
		err := im.validateMemoMsg(receiver, sent, tc.msgs)
		if tc.errstr != "" {
			assert.ErrorContains(err, tc.errstr, "test: %d", i)
		} else {
			assert.NoError(err, "test: %d", i)
		}
	}
}
