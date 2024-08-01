package ibc

import (
	"encoding/json"
	"strings"

	sdkmath "cosmossdk.io/math"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcerrors "github.com/cosmos/ibc-go/v8/modules/core/errors"
)

// GetFundsFromPacket returns transfer amount and denom
func GetFundsFromPacket(data []byte) (sdkmath.Int, string, error) {
	var packetData transfertypes.FungibleTokenPacketData
	err := json.Unmarshal(data, &packetData)
	if err != nil {
		return sdkmath.Int{}, "", err
	}

	amount, ok := sdkmath.NewIntFromString(packetData.Amount)
	if !ok {
		return sdkmath.Int{}, "", ibcerrors.ErrInvalidRequest.Wrapf("invalid transfer amount %s", packetData.Amount)
	}

	return amount, GetLocalDenom(packetData.Denom), nil
}

// GetLocalDenom returns ibc denom
// Expected denoms in the following cases:
//
// send non-native: transfer/channel-0/denom -> ibc/xxx
// send native: denom -> denom
// recv (B)non-native: denom
// recv (B)native: transfer/channel-0/denom
func GetLocalDenom(denom string) string {
	if strings.HasPrefix(denom, "transfer/") {
		denomTrace := transfertypes.ParseDenomTrace(denom)
		return denomTrace.IBCDenom()
	}

	return denom
}
