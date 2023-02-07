package uibc

import (
	"cosmossdk.io/errors"
)

var (
	ErrQuotaExceeded      = errors.Register(ModuleName, 1, "quota transfer exceeded")
	ErrNoQuotaForIBCDenom = errors.Register(ModuleName, 2, "no quota for ibc denom")
)
