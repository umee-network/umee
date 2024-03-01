package uics20

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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
	asset := coin.New("atom", 10)
	assetH := coin.New("atom", 5)
	asset11 := coin.New("atom", 11)
	sent := coin.New("atom", 10)
	goodMsgSupply := ltypes.NewMsgSupply(receiver, asset)
	goodMsgSupply11 := ltypes.NewMsgSupply(receiver, asset11)
	goodMsgSupplyColl := ltypes.NewMsgSupplyCollateral(receiver, asset)
	// goodMsgSupplyCollH := ltypes.NewMsgSupplyCollateral(receiver, assetH)
	goodMsgSupplyColl11 := ltypes.NewMsgSupplyCollateral(receiver, asset11)
	goodMsgBorrow := ltypes.NewMsgBorrow(receiver, asset)
	goodMsgBorrowH := ltypes.NewMsgBorrow(receiver, assetH)
	goodMsgLiquidate := ltypes.NewMsgLiquidate(receiver, accs.Bob, assetH, "uumee")
	goodMsgLiquidate11 := ltypes.NewMsgLiquidate(receiver, accs.Bob, asset11, "uumee")
	msgSend := &banktypes.MsgSend{FromAddress: receiver.String()}

	errManyMsgs := "memo with more than 1 message is not supported"
	errNoSubCoins := errNoSubCoins.Error()
	errMsg0type := errMsg0Type.Error()
	// errWrongSigner := "signer doesn't match the receiver"

	mh := MemoHandler{leverage: mocks.NewLvgNoopMsgSrv()}
	tcs := []struct {
		msgs   []sdk.Msg
		errstr string
	}{
		/** we don't check signers in handlers v1
		{[]sdk.Msg{ltypes.NewMsgSupply(accs.Bob, asset)}, errWrongSigner},
		{[]sdk.Msg{ltypes.NewMsgSupplyCollateral(accs.Bob, asset)}, errWrongSigner},
		{[]sdk.Msg{goodMsgSupplyColl,
			ltypes.NewMsgBorrow(accs.Bob, asset)}, errWrongSigner},
		*/

		// good messages[0]
		{[]sdk.Msg{goodMsgSupply}, ""},
		{[]sdk.Msg{goodMsgSupplyColl}, ""},
		{[]sdk.Msg{goodMsgLiquidate}, ""}, // in handlers v2 this will be a good message

		// messages[0] use more assets than the transfer
		{[]sdk.Msg{goodMsgSupply11}, errNoSubCoins},
		{[]sdk.Msg{goodMsgSupplyColl11}, errNoSubCoins},
		{[]sdk.Msg{goodMsgSupplyColl11}, errNoSubCoins},
		{[]sdk.Msg{goodMsgLiquidate11}, errNoSubCoins},

		// wrong message types
		{[]sdk.Msg{goodMsgBorrow}, errMsg0type},
		{[]sdk.Msg{msgSend}, errMsg0type}, // bank msg
		{[]sdk.Msg{ltypes.NewMsgDecollateralize(receiver, asset)}, errMsg0type},
		{[]sdk.Msg{&ltypes.MsgLeveragedLiquidate{Liquidator: receiver.String()}}, errMsg0type},

		{[]sdk.Msg{goodMsgBorrow, goodMsgBorrow}, errManyMsgs},
		{[]sdk.Msg{goodMsgBorrow, goodMsgSupplyColl}, errManyMsgs},
		{[]sdk.Msg{goodMsgSupplyColl, goodMsgSupplyColl}, errManyMsgs},
		{[]sdk.Msg{goodMsgSupplyColl, goodMsgLiquidate}, errManyMsgs},
		{[]sdk.Msg{goodMsgSupplyColl, msgSend}, errManyMsgs},
		{[]sdk.Msg{goodMsgSupplyColl, goodMsgBorrow}, errManyMsgs},

		/** uncomment when msg borrow enabled
		// check msg borrow is after supply collateral
		{[]sdk.Msg{goodMsgSupply, goodMsgBorrow}, "MsgBorrow must use MsgSupplyCollateral"},
		{[]sdk.Msg{goodMsgSupply, goodMsgBorrowH}, "MsgBorrow must use MsgSupplyCollateral"},
		{[]sdk.Msg{goodMsgLiquidate, goodMsgBorrowH}, "MsgBorrow must use MsgSupplyCollateral"},
		{[]sdk.Msg{goodMsgSupplyCollH, goodMsgBorrow}, "MsgBorrow must use MsgSupplyCollateral"},
		{[]sdk.Msg{msgSend, goodMsgBorrow}, msg0typeErr},
		{[]sdk.Msg{goodMsgSupplyCollH, goodMsgBorrowH}, ""},
		{[]sdk.Msg{goodMsgSupplyColl, goodMsgBorrow}, ""},
		*/

		// more than 2 messages
		{[]sdk.Msg{goodMsgSupplyColl, goodMsgBorrowH, goodMsgBorrowH}, errManyMsgs},
		{[]sdk.Msg{goodMsgSupplyColl, goodMsgLiquidate, goodMsgBorrow}, errManyMsgs},
		{[]sdk.Msg{goodMsgSupplyColl, goodMsgBorrowH, goodMsgBorrowH, goodMsgBorrowH}, errManyMsgs},
	}

	for i, tc := range tcs {
		mh.receiver = receiver
		mh.received = sent
		mh.msgs = tc.msgs
		err := mh.validateMemoMsg()
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

	memo2, err := deserializeMemo(cdc, bz)
	assert.NoError(err)
	assert.Equal(memo2.FallbackAddr, "")
	msgs2, err := memo2.GetMsgs()
	for i := range msgs2 {
		assert.Equal(msgs[i], msgs2[i], "idx=%d", i)
	}

	bz = []byte("{}")
	memo2, err = deserializeMemo(cdc, bz)
	assert.NoError(err)
	assert.Equal(memo2, uibc.ICS20Memo{})

	// we expect to fail deserialization if Any is not properly formatted
	bz = []byte(`{"messages": ["any message"]}`)
	memo2, err = deserializeMemo(cdc, bz)
	assert.Error(err)
}
