package gmp

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ParseMemo will parse the incoming gmp memo
func ParseMemo(ctx sdk.Context, coinReceived sdk.Coin, memo string, receiver sdk.AccAddress) (Memo, error) {
	if len(memo) == 0 {
		return Memo{}, nil
	}

	var msg Memo
	if err := json.Unmarshal([]byte(memo), &msg); err != nil {
		return msg, err
	}

	switch msg.Type {
	case TypeGeneralMessage:
		return msg, fmt.Errorf("msg.type=%d (TypeGeneralMessage) is not supported. Supported types include: %d",
			TypeGeneralMessage, TypeGeneralMessageWithToken)
	case TypeGeneralMessageWithToken:
		return msg, nil
	default:
		return msg, fmt.Errorf("unrecognized gmp message type: %d", msg.Type)
	}
}
