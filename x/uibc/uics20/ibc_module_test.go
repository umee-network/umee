package uics20

import (
	"encoding/json"
	"testing"

	sdkmath "cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	ics20types "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/golang/mock/gomock"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v6/tests/tsdk"
	ltypes "github.com/umee-network/umee/v6/x/leverage/types"
	ugovmocks "github.com/umee-network/umee/v6/x/ugov/mocks"
	"github.com/umee-network/umee/v6/x/uibc"
	"github.com/umee-network/umee/v6/x/uibc/mocks"
	"github.com/umee-network/umee/v6/x/uibc/quota"
)

var (
	tokenAmount = sdkmath.NewInt(100_000000)
	// sender sending from their native token to umee
	// here umee is receiver
	atomCoin     = sdk.NewCoin("uatom", tokenAmount)
	senderAddr   = "umee1mjk79fjjgpplak5wq838w0yd982gzkyf3qjpef"
	recvAddr     = "umee1y6xz2ggfc0pcsmyjlekh0j9pxh6hk87ymc9due"
	fallbackAddr = "umee10h9stc5v6ntgeygf5xf945njqq5h32r5r2argu"
	relAddr      = sdk.MustAccAddressFromBech32(senderAddr)
	ftData       = ics20types.FungibleTokenPacketData{
		Denom:    atomCoin.Denom,
		Amount:   atomCoin.Amount.String(),
		Sender:   senderAddr,
		Receiver: recvAddr,
	}

	// ftData.
	packet = channeltypes.Packet{
		Sequence:           10,
		SourcePort:         "transfer",
		DestinationPort:    "transfer",
		SourceChannel:      "channel-10",
		DestinationChannel: "channel-1",
	}

	atomIBCDenom = uibc.ExtractDenomFromPacketOnRecv(packet, ftData.Denom)
	atomIBC      = sdk.NewCoin(atomIBCDenom, tokenAmount)
)

func TestIBCOnRecvPacket(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	leverageMock := mocks.NewMockLeverage(ctrl)
	oracleMock := mocks.NewMockOracle(ctrl)
	mockLeverageMsgServer := NewMockLeverageMsgServer()
	mockIBCModule := NewMockIBCModule()

	cdc := tsdk.NewCodec(uibc.RegisterInterfaces, ltypes.RegisterInterfaces)
	storeKey := storetypes.NewMemoryStoreKey("quota")
	ctx, _ := tsdk.NewCtxOneStore(t, storeKey)
	eg := ugovmocks.NewSimpleEmergencyGroupBuilder()
	kb := quota.NewKeeperBuilder(cdc, storeKey, leverageMock, oracleMock, eg, bkeeper.BaseKeeper{})
	ics20Module := NewICS20Module(mockIBCModule, cdc, kb, mockLeverageMsgServer)

	validMemoMsgs := func(noOfMsgs int, fallbackAddr string) string {
		msgs := make([]*codectypes.Any, 0)
		msg, err := codectypes.NewAnyWithValue(ltypes.NewMsgSupply(relAddr, atomIBC))
		assert.NilError(t, err)
		for i := 0; i < noOfMsgs; i++ {
			msgs = append(msgs, msg)
		}
		validMemo := uibc.ICS20Memo{
			Messages: msgs,
		}
		if fallbackAddr != "" {
			validMemo.FallbackAddr = fallbackAddr
		}
		return string(cdc.MustMarshalJSON(&validMemo))
	}

	tests := []struct {
		name string
		memo string
	}{
		{
			name: "fungible token packet data without memo",
			memo: "",
		},
		{
			name: "fungible token packet data with invalid memo message",
			memo: "invalid_memo_message",
		},
		{
			name: "valid memo without fallback_addr",
			memo: validMemoMsgs(1, ""),
		},
		{
			name: "valid memo with valid fallback_addr",
			memo: validMemoMsgs(1, fallbackAddr),
		},
		{
			name: "valid memo (more than one message) with valid fallback_addr",
			memo: validMemoMsgs(3, fallbackAddr),
		},
	}

	for _, tt := range tests {
		ftData.Memo = tt.memo
		mar, err := json.Marshal(ftData)
		assert.NilError(t, err)
		packet.Data = mar

		t.Run(tt.name, func(t *testing.T) {
			acc := ics20Module.OnRecvPacket(ctx, packet, relAddr)
			assert.Equal(t, true, acc.Success())
		})
	}
}

func TestDeserializeFTData(t *testing.T) {
	cdc := tsdk.NewCodec(uibc.RegisterInterfaces, ltypes.RegisterInterfaces)
	serialize := func() []byte {
		d, err := json.Marshal(ftData)
		assert.NilError(t, err)
		return d
	}
	tests := []struct {
		name       string
		packetData func() []byte
		errMsg     string
	}{
		{
			name: "invalid json",
			packetData: func() []byte {
				return []byte("invalid packet data")
			},
			errMsg: "invalid character",
		},
		{
			name:       "valid packet data ",
			packetData: serialize,
			errMsg:     "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			packet.Data = tc.packetData()
			recvFtData, err := deserializeFTData(cdc, packet)
			if tc.errMsg != "" {
				assert.ErrorContains(t, err, tc.errMsg)
			} else {
				assert.NilError(t, err)
				assert.DeepEqual(t, recvFtData, ftData)
			}
		})
	}
}
