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
	wrongSignerErr := "signer doesn't match the receiver"
	asset := coin.New("atom", 10)
	assetH := coin.New("atom", 5)
	asset11 := coin.New("atom", 11)
	sent := coin.New("atom", 10)
	goodMsgSupply := ltypes.NewMsgSupply(receiver, asset)
	goodMsgSupply11 := ltypes.NewMsgSupply(receiver, asset11)
	goodMsgSupplyColl := ltypes.NewMsgSupplyCollateral(receiver, asset)
	goodMsgSupplyCollH := ltypes.NewMsgSupplyCollateral(receiver, assetH)
	goodMsgSupplyColl11 := ltypes.NewMsgSupplyCollateral(receiver, asset11)
	goodMsgBorrow := ltypes.NewMsgBorrow(receiver, asset)
	goodMsgBorrowH := ltypes.NewMsgBorrow(receiver, assetH)
	goodMsgLiquidate := ltypes.NewMsgLiquidate(receiver, accs.Bob, assetH, "uumee")
	goodMsgLiquidate11 := ltypes.NewMsgLiquidate(receiver, accs.Bob, asset11, "uumee")
	msgSend := &banktypes.MsgSend{FromAddress: receiver.String()}

	mh := MemoHandler{leverage: mocks.NewLvgNoopMsgSrv()}
	tcs := []struct {
		msgs   []sdk.Msg
		errstr string
	}{
		{[]sdk.Msg{ltypes.NewMsgSupply(accs.Bob, asset)}, wrongSignerErr},
		{[]sdk.Msg{ltypes.NewMsgSupplyCollateral(accs.Bob, asset)}, wrongSignerErr},
		{[]sdk.Msg{goodMsgSupplyColl,
			ltypes.NewMsgBorrow(accs.Bob, asset)}, wrongSignerErr},

		// good messages[0]
		{[]sdk.Msg{goodMsgSupply}, ""},
		{[]sdk.Msg{goodMsgSupplyColl}, ""},
		{[]sdk.Msg{goodMsgLiquidate}, ""},

		// messages[0] use more assets than the transfer
		{[]sdk.Msg{goodMsgSupply11}, "message must use only coins sent from the transfer"},
		{[]sdk.Msg{goodMsgSupplyColl11}, "message must use only coins sent from the transfer"},
		{[]sdk.Msg{goodMsgLiquidate11}, "message must use only coins sent from the transfer"},

		// wrong message types
		{[]sdk.Msg{goodMsgBorrow}, msg0typeErr},
		{[]sdk.Msg{msgSend}, msg0typeErr}, // bank msg
		{[]sdk.Msg{ltypes.NewMsgDecollateralize(receiver, asset)}, msg0typeErr},
		{[]sdk.Msg{goodMsgBorrow, goodMsgBorrow}, msg0typeErr},
		{[]sdk.Msg{goodMsgBorrow, goodMsgSupplyColl}, msg0typeErr},
		{[]sdk.Msg{&ltypes.MsgLeveragedLiquidate{Liquidator: receiver.String()}}, msg0typeErr},
		{[]sdk.Msg{goodMsgSupplyColl, goodMsgSupplyColl}, msg1typeErr},
		{[]sdk.Msg{goodMsgSupplyColl, goodMsgLiquidate}, msg1typeErr},
		{[]sdk.Msg{goodMsgSupplyColl, msgSend}, msg1typeErr},

		// check msg borrow is after supply collateral
		{[]sdk.Msg{goodMsgSupply, goodMsgBorrow}, "MsgBorrow must use MsgSupplyCollateral"},
		{[]sdk.Msg{goodMsgSupply, goodMsgBorrowH}, "MsgBorrow must use MsgSupplyCollateral"},
		{[]sdk.Msg{goodMsgLiquidate, goodMsgBorrowH}, "MsgBorrow must use MsgSupplyCollateral"},
		{[]sdk.Msg{goodMsgSupplyCollH, goodMsgBorrow}, "MsgBorrow must use MsgSupplyCollateral"},
		{[]sdk.Msg{msgSend, goodMsgBorrow}, msg0typeErr},
		{[]sdk.Msg{goodMsgSupplyCollH, goodMsgBorrowH}, ""},
		{[]sdk.Msg{goodMsgSupplyColl, goodMsgBorrow}, ""},

		// more than 2 messages
		{[]sdk.Msg{goodMsgSupplyColl, goodMsgBorrowH, goodMsgBorrowH},
			"memo with more than 2 messages are not supported"},
		{[]sdk.Msg{goodMsgSupplyColl, goodMsgLiquidate, goodMsgBorrow},
			"memo with more than 2 messages are not supported"},
		{[]sdk.Msg{goodMsgSupplyColl, goodMsgBorrowH, goodMsgBorrowH, goodMsgBorrowH},
			"memo with more than 2 messages are not supported"},
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
