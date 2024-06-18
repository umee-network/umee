package ibc

import (
	"strconv"
	"strings"
	"testing"

	"github.com/cometbft/cometbft/crypto"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/umee-network/umee/v6/tests/tsdk"
	"gotest.tools/v3/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
		"memo",
	)

	famount, fdenom, err := GetFundsFromPacket(data.GetBytes())
	assert.NilError(t, err)
	assert.Equal(t, denom, fdenom)
	assert.Equal(t, famount.String(), amount)

	// invalid address
	data.Receiver = tsdk.GenerateString(MaximumReceiverLength + 1)
	_, _, err = GetFundsFromPacket(data.GetBytes())
	assert.ErrorContains(t, err, "recipient address must not exceed")

	// invalid memo
	data.Receiver = AddressFromString("a4")
	data.Memo = tsdk.GenerateString(MaximumMemoLength + 1)
	_, _, err = GetFundsFromPacket(data.GetBytes())
	assert.ErrorContains(t, err, "memo must not exceed")

	// valid address and memo
	data.Receiver = tsdk.GenerateString(MaximumReceiverLength)
	data.Memo = tsdk.GenerateString(MaximumMemoLength)
	_, _, err = GetFundsFromPacket(data.GetBytes())
	assert.NilError(t, err, "Should handle valid inputs without error")
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
