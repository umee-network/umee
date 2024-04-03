package gmp

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PaseMemo will parse the incoming gmp memo
func ParseMemo(ctx sdk.Context, coinReceived sdk.Coin, memo string, receiver sdk.AccAddress) (GMPMemo, error) {
	if len(memo) == 0 {
		return GMPMemo{}, nil
	}

	var msg GMPMemo
	if err := json.Unmarshal([]byte(memo), &msg); err != nil {
		return msg, err
	}

	switch msg.Type {
	case TypeGeneralMessage:
		return msg, fmt.Errorf("only msg.type=%d (TypeGeneralMessageWithToken) is supported",
			TypeGeneralMessageWithToken)
	case TypeGeneralMessageWithToken:
		return msg, nil
	default:
		return msg, fmt.Errorf("unrecognized gmp message type: %d", msg.Type)
	}
}
