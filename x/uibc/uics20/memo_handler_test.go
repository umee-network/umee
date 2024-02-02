package uics20

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/stretchr/testify/assert"

	"github.com/umee-network/umee/v6/tests/accs"
	"github.com/umee-network/umee/v6/tests/tsdk"
	"github.com/umee-network/umee/v6/util/coin"
	ltypes "github.com/umee-network/umee/v6/x/leverage/types"
	"github.com/umee-network/umee/v6/x/uibc"
	"github.com/umee-network/umee/v6/x/uibc/mocks"
)

func TestValidateMemoMsg(t *testing.T) {
	assert := assert.New(t)
	receiver := accs.Alice
	wrongSignerErr := "signer doesn't match the receiver"
	asset := coin.New("atom", 10)
	sent := coin.New("atom", 10)
	goodMsgSupplyColl := ltypes.NewMsgSupplyCollateral(receiver, asset)
	goodMsgBorrow := ltypes.NewMsgBorrow(receiver, asset)
	mh := MemoHandler{leverage: mocks.NewLvgNoopMsgSrv()}
	tcs := []struct {
		msgs   []sdk.Msg
		errstr string
	}{
		{[]sdk.Msg{ltypes.NewMsgSupply(accs.Bob, asset)}, wrongSignerErr},
		{[]sdk.Msg{ltypes.NewMsgSupplyCollateral(accs.Bob, asset)}, wrongSignerErr},

		{[]sdk.Msg{ltypes.NewMsgSupply(receiver, asset)}, ""},
		{[]sdk.Msg{goodMsgSupplyColl}, ""},

		{[]sdk.Msg{goodMsgSupplyColl,
			ltypes.NewMsgBorrow(accs.Bob, asset)}, wrongSignerErr},
		{[]sdk.Msg{goodMsgSupplyColl, goodMsgSupplyColl}, "only MsgBorrow is supported as "},
		{[]sdk.Msg{goodMsgSupplyColl, goodMsgBorrow}, ""},

		{[]sdk.Msg{goodMsgSupplyColl, goodMsgBorrow, goodMsgBorrow},
			"memo with more than 2 messages are not supported"},

		{[]sdk.Msg{goodMsgBorrow},
			" are supported as messages[0]"},
		{[]sdk.Msg{ltypes.NewMsgDecollateralize(receiver, asset)},
			" are supported as messages[0]"},
	}

	for i, tc := range tcs {
		err := mh.validateMemoMsg(receiver, sent, tc.msgs)
		if tc.errstr != "" {
			assert.ErrorContains(err, tc.errstr, "test: %d", i)
		} else {
			assert.NoError(err, "test: %d", i)
		}
	}
}

func TestMsgMarshalling(t *testing.T) {
	assert := assert.New(t)
	cdc := tsdk.NewCodec(uibc.RegisterInterfaces, ltypes.RegisterInterfaces)
	msgs := []sdk.Msg{
		&uibc.MsgGovSetIBCStatus{
			Authority: "auth1", Description: "d1",
			IbcStatus: uibc.IBCTransferStatus_IBC_TRANSFER_STATUS_QUOTA_OUT_DISABLED,
		},
		ltypes.NewMsgCollateralize(accs.Alice, sdk.NewCoin("ATOM", sdk.NewInt(1020))),
	}
	anyMsg, err := tx.SetMsgs(msgs)
	assert.NoError(err)
	memo := uibc.ICS20Memo{Messages: anyMsg}

	bz, err := cdc.MarshalJSON(&memo)
	assert.NoError(err)

	msgs2, err := deserializeMemoMsgs(cdc, bz)
	assert.NoError(err)
	for i := range msgs2 {
		assert.Equal(msgs[i], msgs2[i], "idx=%d", i)
	}
}
