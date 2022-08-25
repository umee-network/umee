package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const ModuleName = "price-feeder"

// Price feeder errors
var (
	ErrProviderConnection = sdkerrors.Register(ModuleName, 1, "provider connection")
)
