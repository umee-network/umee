package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	ltypes "github.com/umee-network/umee/v4/x/leverage/types"
)

type LeverageKeeper interface {
	GetTokenSettings(ctx sdk.Context, baseDenom string) (ltypes.Token, error)
}
