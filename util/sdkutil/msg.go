package sdkutil

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// StartMsg unpacks sdk.Context and validates msg.
func StartMsg(ctx context.Context, msg sdk.Msg) (sdk.Context, error) {
	if err := msg.ValidateBasic(); err != nil {
		return sdk.Context{}, err
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	var err error
	if fvmsg, ok := msg.(fullValidate); ok {
		err = fvmsg.Validate(&sdkCtx)
	}
	return sdkCtx, err
}

type fullValidate interface {
	Validate(*sdk.Context) error
}
