package ibc

import (
	"strconv"
	"strings"
	"testing"

	"github.com/cometbft/cometbft/crypto"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
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
