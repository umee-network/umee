package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const ModuleName = "price-feeder"

// Price feeder errors
var (
	ErrProviderConnection  = sdkerrors.Register(ModuleName, 1, "provider connection")
	ErrMissingExchangeRate = sdkerrors.Register(ModuleName, 2, "missing exchange rate for %s")
	ErrTickerNotFound      = sdkerrors.Register(ModuleName, 3, "%s failed to get ticker price for %s")
	ErrCandleNotFound      = sdkerrors.Register(ModuleName, 4, "%s failed to get candle price for %s")
)
