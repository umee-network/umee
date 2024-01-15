package uics20_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/stretchr/testify/assert"

	"github.com/umee-network/umee/v6/tests/accs"
	"github.com/umee-network/umee/v6/tests/tsdk"
	ltypes "github.com/umee-network/umee/v6/x/leverage/types"
	"github.com/umee-network/umee/v6/x/uibc"
	"github.com/umee-network/umee/v6/x/uibc/uics20"
)

func TestMsgMarshalling(t *testing.T) {
	assert := assert.New(t)
	cdc := tsdk.NewCodec(uibc.RegisterInterfaces, ltypes.RegisterInterfaces)
	msgs := []sdk.Msg{
		&uibc.MsgGovSetIBCStatus{Authority: "auth1", Description: "d1",
			IbcStatus: uibc.IBCTransferStatus_IBC_TRANSFER_STATUS_QUOTA_OUT_DISABLED},
		ltypes.NewMsgCollateralize(accs.Alice, sdk.NewCoin("ATOM", sdkmath.NewInt(1020))),
	}
	anyMsg, err := tx.SetMsgs(msgs)
	assert.NoError(err)
	var memo = uibc.ICS20Memo{Messages: anyMsg}

	bz, err := cdc.MarshalJSON(&memo)
	assert.NoError(err)

	msgs2, err := uics20.DeserializeMemoMsgs(cdc, bz)
	assert.NoError(err)
	for i := range msgs2 {
		assert.Equal(msgs[i], msgs2[i], "idx=%d", i)
	}
}
