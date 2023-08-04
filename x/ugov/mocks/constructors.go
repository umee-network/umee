package mocks

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ugov "github.com/umee-network/umee/v5/x/ugov"
)

// NewUgovParamsBuilder creates a ParamsKeeper builder function
func NewUgovParamsBuilder(pk ugov.ParamsKeeper) ugov.ParamsKeeperBuilder {
	return func(_ *sdk.Context) ugov.ParamsKeeper {
		return pk
	}
}
