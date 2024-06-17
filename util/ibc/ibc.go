package ibc

import (
	"encoding/json"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcerrors "github.com/cosmos/ibc-go/v7/modules/core/errors"
)

const (
	MaximumReceiverLength = 2048  // maximum length of the receiver address in bytes (value chosen arbitrarily)
	MaximumMemoLength     = 32768 // maximum length of the memo in bytes (value chosen arbitrarily)
)

func ValidateRecvAddr(receiver string) error {
	if len(receiver) > MaximumReceiverLength {
		return ibcerrors.ErrInvalidAddress.Wrapf("recipient address must not exceed %d bytes", MaximumReceiverLength)
	}
	return nil
}

func ValidateMemo(memo string) error {
	if len(memo) > MaximumMemoLength {
		return fmt.Errorf("memo must not exceed %d bytes", MaximumMemoLength)
	}
	return nil
}

// GetFundsFromPacket returns transfer amount and denom
func GetFundsFromPacket(data []byte) (sdkmath.Int, string, error) {
	var packetData transfertypes.FungibleTokenPacketData
	err := json.Unmarshal(data, &packetData)
	if err != nil {
		return sdkmath.Int{}, "", err
	}

	if err := ValidateRecvAddr(packetData.Receiver); err != nil {
		return sdkmath.Int{}, "", err
	}

	if err := ValidateMemo(packetData.Memo); err != nil {
		return sdkmath.Int{}, "", err
	}

	amount, ok := sdkmath.NewIntFromString(packetData.Amount)
	if !ok {
		return sdkmath.Int{}, "", ibcerrors.ErrInvalidRequest.Wrapf("invalid transfer amount %s", packetData.Amount)
	}

	return amount, GetLocalDenom(packetData.Denom), nil
}

// GetLocalDenom retruns ibc denom
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
