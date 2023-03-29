package ibc

import (
	"strconv"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	"github.com/tendermint/tendermint/crypto"
	"gotest.tools/v3/assert"
)

func TestGetFundsFromPacket(t *testing.T) {
	denom := strings.Join([]string{
		"transfer",
		"dest_chain",
		"quark",
	}, "/")

	amount := strconv.Itoa(1)
	data := ibctransfertypes.NewFungibleTokenPacketData(
		denom,
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

	famount, fdenom, err := GetFundsFromPacket(packet)

	assert.NilError(t, err)
	assert.Equal(t, denom, fdenom)
	assert.Equal(t, famount.String(), amount)
}

func TestGetLocalDenom(t *testing.T) {
	denom := strings.Join([]string{
		"transfer",
		"dest_chain",
		"quark",
	}, "/")

	rdenom := GetLocalDenom(denom)
	assert.Equal(t, rdenom, denom)
}

func AddressFromString(address string) string {
	return sdk.AccAddress(crypto.AddressHash([]byte(address))).String()
}
