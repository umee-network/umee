package uics20_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ics20types "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	"github.com/golang/mock/gomock"

	"github.com/umee-network/umee/v6/tests/tsdk"
	lfixtures "github.com/umee-network/umee/v6/x/leverage/fixtures"
	ltypes "github.com/umee-network/umee/v6/x/leverage/types"
	"github.com/umee-network/umee/v6/x/oracle/types"
	ugovmocks "github.com/umee-network/umee/v6/x/ugov/mocks"
	"github.com/umee-network/umee/v6/x/uibc"
	"github.com/umee-network/umee/v6/x/uibc/mocks"
	"github.com/umee-network/umee/v6/x/uibc/quota"
	"github.com/umee-network/umee/v6/x/uibc/uics20"
)

type MockICS4Wrapper struct {
	porttypes.ICS4Wrapper
}

func (m MockICS4Wrapper) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	return 1, nil
}

func TestSendPacket(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	leverageMock := mocks.NewMockLeverage(ctrl)
	oracleMock := mocks.NewMockOracle(ctrl)
	eg := ugovmocks.NewSimpleEmergencyGroupBuilder()
	mock := MockICS4Wrapper{}

	storeKey := storetypes.NewMemoryStoreKey("quota")
	ctx, _ := tsdk.NewCtxOneStore(t, storeKey)
	kb := quota.NewKeeperBuilder(codec.NewProtoCodec(nil), storeKey, leverageMock, oracleMock, eg)
	dp := uibc.DefaultParams()
	keeper := kb.Keeper(&ctx)
	keeper.SetParams(dp)

	leverageMock.EXPECT().GetTokenSettings(ctx, "test").Return(lfixtures.Token("test", "TEST", 6), nil).AnyTimes()
	leverageMock.EXPECT().GetTokenSettings(ctx, "umee").Return(ltypes.Token{}, ltypes.ErrNotRegisteredToken).AnyTimes()
	oracleMock.EXPECT().Price(ctx, "TEST").Return(sdk.Dec{}, types.ErrMalformedLatestAvgPrice)

	ics4 := uics20.NewICS4(mock, kb)

	// error test cases
	_, err := ics4.SendPacket(ctx, nil, "", "", clienttypes.NewHeight(1, 1), 1, nil)
	assert.ErrorContains(t, err, "bad packet in rate limit's SendPacket")

	tests := []struct {
		name         string
		data         []byte
		preRun       func()
		expectedErr  error
		expectedResp uint64
	}{
		{
			name: "transfers paused",
			preRun: func() {
				dp.IbcStatus = uibc.IBCTransferStatus_IBC_TRANSFER_STATUS_TRANSFERS_PAUSED
				keeper.SetParams(dp)
			},
			data:         nil,
			expectedErr:  ics20types.ErrSendDisabled,
			expectedResp: 0,
		},
		{
			name: "malformed price",
			preRun: func() {
				dp.IbcStatus = uibc.IBCTransferStatus_IBC_TRANSFER_STATUS_QUOTA_ENABLED
				keeper.SetParams(dp)
			},
			data:         ibctransfertypes.NewFungibleTokenPacketData("test", "1", "a3", "a4", "memo").GetBytes(),
			expectedErr:  types.ErrMalformedLatestAvgPrice,
			expectedResp: 0,
		},
		{
			name:         "success",
			preRun:       func() {},
			data:         ibctransfertypes.NewFungibleTokenPacketData("umee", "1", "a3", "a4", "memo").GetBytes(),
			expectedErr:  nil,
			expectedResp: uint64(1),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.preRun()
			resp, err := ics4.SendPacket(ctx, nil, "", "", clienttypes.NewHeight(1, 1), 1, tc.data)
			if tc.expectedResp != 1 {
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				assert.NilError(t, err)
				assert.Equal(t, tc.expectedResp, resp)
			}
		})
	}
}
