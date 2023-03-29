package ics20

import (
	"strconv"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	"github.com/golang/mock/gomock"
	"github.com/tendermint/tendermint/crypto"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"gotest.tools/v3/assert"

	lfixtures "github.com/umee-network/umee/v4/x/leverage/fixtures"
	ltypes "github.com/umee-network/umee/v4/x/leverage/types"
	mocks "github.com/umee-network/umee/v4/x/uibc/mocks"
)

func TestCheckIBCInflow(t *testing.T) {
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, nil)
	notRegisteredDenom := strings.Join([]string{
		"transfer",
		"dest_chain",
		"quark",
	}, "/")
	amount := strconv.Itoa(1)

	// gomock initializations
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mlk := mocks.NewMockLeverageKeeper(ctrl)
	// registered token, expected token
	mlk.EXPECT().GetTokenSettings(ctx, "UMEE").Return(lfixtures.Token("uumee", "UMEE", 6), nil).AnyTimes()
	// Not registered token, error is expected
	mlk.EXPECT().GetTokenSettings(ctx, "ibc/C5417054F532756246BD4694713BE30A10D53C878EBF5E2E9EC49BE86B7D89BB").
		Return(lfixtures.Token("uumee", "UMEE", 6), ltypes.ErrNotRegisteredToken).AnyTimes()

	// registered token
	data := ibctransfertypes.NewFungibleTokenPacketData(
		"UMEE",
		amount,
		AddressFromString("a3"),
		AddressFromString("a4"),
	)

	packet := channeltypes.NewPacket(
		data.GetBytes(),
		uint64(1),
		"transfer",
		"channel-0",
		"transfer",
		"channel-0",
		clienttypes.NewHeight(0, 100),
		0,
	)

	ackErr := CheckIBCInflow(ctx, packet, mlk, data.Denom, false)
	assert.Equal(t, nil, ackErr)

	// Not registered token
	data.Denom = notRegisteredDenom
	packet.Data = data.GetBytes()

	ackErr = CheckIBCInflow(ctx, packet, mlk, data.Denom, true)
	assert.Equal(t, false, ackErr.Success())
}

func AddressFromString(address string) string {
	return sdk.AccAddress(crypto.AddressHash([]byte(address))).String()
}
