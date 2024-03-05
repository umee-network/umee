package uics20

import (
	"encoding/json"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	atomIBC      = sdk.NewCoin("ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9", tokenAmount)
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
	packet = channeltypes.Packet{
		Sequence:           10,
		SourcePort:         "transfer",
		DestinationPort:    "transfer",
		SourceChannel:      "channel-10",
		DestinationChannel: "channel-1",
	}
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
	kb := quota.NewKeeperBuilder(cdc, storeKey, leverageMock, oracleMock, eg)
	ics20Module := NewICS20Module(mockIBCModule, cdc, kb, mockLeverageMsgServer)

	validMemoMsgs := func(noOfMsgs int) []*codectypes.Any {
		msgs := make([]*codectypes.Any, 0)
		msg, err := codectypes.NewAnyWithValue(ltypes.NewMsgSupply(relAddr, atomIBC))
		assert.NilError(t, err)
		for i := 0; i < noOfMsgs; i++ {
			msgs = append(msgs, msg)
		}
		return msgs
	}

	tests := []struct {
		name string
		memo func(cdc codec.Codec) string
	}{
		{
			name: "fungible token packet data without memo",
			memo: func(cdc codec.Codec) string {
				return ""
			},
		},
		{
			name: "fungible token packet data with invalid memo message",
			memo: func(cdc codec.Codec) string {
				return "invalid_memo_message"
			},
		},
		{
			name: "valid memo without fallback_addr",
			memo: func(cdc codec.Codec) string {
				msgs := validMemoMsgs(1)
				validMemo := uibc.ICS20Memo{
					Messages: msgs,
				}
				return string(cdc.MustMarshalJSON(&validMemo))
			},
		},
		{
			name: "valid memo with valid fallback_addr",
			memo: func(cdc codec.Codec) string {
				msgs := validMemoMsgs(1)
				validMemo := uibc.ICS20Memo{
					Messages:     msgs,
					FallbackAddr: fallbackAddr,
				}
				return string(cdc.MustMarshalJSON(&validMemo))
			},
		},
		{
			name: "valid memo (more than one message) with valid fallback_addr",
			memo: func(cdc codec.Codec) string {
				msgs := validMemoMsgs(3)
				validMemo := uibc.ICS20Memo{
					Messages:     msgs,
					FallbackAddr: fallbackAddr,
				}
				return string(cdc.MustMarshalJSON(&validMemo))
			},
		},
	}

	for _, tt := range tests {
		ftData.Memo = tt.memo(cdc)
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

	tests := []struct {
		name       string
		packetData func() []byte
		errMsg     string
	}{
		{
			name: "invalid packet data",
			packetData: func() []byte {
				return []byte("invalid packet data")
			},
			errMsg: "invalid character",
		},
		{
			name: "valid packet data ",
			packetData: func() []byte {
				mar, err := json.Marshal(ftData)
				assert.NilError(t, err)
				return mar
			},
			errMsg: "",
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
