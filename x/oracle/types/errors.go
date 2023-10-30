package types

import (
	"fmt"

	"cosmossdk.io/errors"
	"github.com/cometbft/cometbft/crypto/tmhash"
)

// Oracle sentinel errors
var (
	ErrInvalidExchangeRate     = errors.Register(ModuleName, 1, "invalid exchange rate")
	ErrNoPrevote               = errors.Register(ModuleName, 2, "no prevote")
	ErrNoVote                  = errors.Register(ModuleName, 3, "no vote")
	ErrNoVotingPermission      = errors.Register(ModuleName, 4, "unauthorized voter")
	ErrInvalidHash             = errors.Register(ModuleName, 5, "invalid hash")
	ErrInvalidHashLength       = errors.Register(ModuleName, 6, fmt.Sprintf("invalid hash length; should equal %d", tmhash.TruncatedSize)) //nolint: lll
	ErrVerificationFailed      = errors.Register(ModuleName, 7, "hash verification failed")
	ErrRevealPeriodMissMatch   = errors.Register(ModuleName, 8, "reveal period of submitted vote does not match with registered prevote") //nolint: lll
	ErrInvalidSaltLength       = errors.Register(ModuleName, 9, "invalid salt length; must be 64")
	ErrInvalidSaltFormat       = errors.Register(ModuleName, 10, "invalid salt format")
	ErrNoAggregatePrevote      = errors.Register(ModuleName, 11, "no aggregate prevote")
	ErrNoAggregateVote         = errors.Register(ModuleName, 12, "no aggregate vote")
	ErrUnknownDenom            = errors.Register(ModuleName, 13, "unknown denom")
	ErrNegativeOrZeroRate      = errors.Register(ModuleName, 14, "invalid exchange rate; should be positive")
	ErrExistingPrevote         = errors.Register(ModuleName, 15, "prevote already submitted for this voting period")
	ErrBallotNotSorted         = errors.Register(ModuleName, 16, "ballot must be sorted before this operation")
	ErrNoHistoricPrice         = errors.Register(ModuleName, 18, "no historic price for this denom at this block")
	ErrNoMedian                = errors.Register(ModuleName, 19, "no median for this denom at this block")
	ErrNoMedianDeviation       = errors.Register(ModuleName, 20, "no median deviation for this denom at this block")
	ErrMalformedLatestAvgPrice = errors.Register(ModuleName, 21, "malformed latest avg price, expecting one byte")
	ErrNoLatestAvgPrice        = errors.Register(ModuleName, 22, "no latest average price")
)
