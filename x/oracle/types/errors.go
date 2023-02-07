package types

import (
	"fmt"

	"cosmossdk.io/errors"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

// Oracle sentinel errors
var (
<<<<<<< HEAD
	ErrInvalidExchangeRate   = sdkerrors.Register(ModuleName, 1, "invalid exchange rate")
	ErrNoPrevote             = sdkerrors.Register(ModuleName, 2, "no prevote")
	ErrNoVote                = sdkerrors.Register(ModuleName, 3, "no vote")
	ErrNoVotingPermission    = sdkerrors.Register(ModuleName, 4, "unauthorized voter")
	ErrInvalidHash           = sdkerrors.Register(ModuleName, 5, "invalid hash")
	ErrInvalidHashLength     = sdkerrors.Register(ModuleName, 6, fmt.Sprintf("invalid hash length; should equal %d", tmhash.TruncatedSize)) //nolint: lll
	ErrVerificationFailed    = sdkerrors.Register(ModuleName, 7, "hash verification failed")
	ErrRevealPeriodMissMatch = sdkerrors.Register(ModuleName, 8, "reveal period of submitted vote does not match with registered prevote") //nolint: lll
	ErrInvalidSaltLength     = sdkerrors.Register(ModuleName, 9, "invalid salt length; must be 64")
	ErrInvalidSaltFormat     = sdkerrors.Register(ModuleName, 10, "invalid salt format")
	ErrNoAggregatePrevote    = sdkerrors.Register(ModuleName, 11, "no aggregate prevote")
	ErrNoAggregateVote       = sdkerrors.Register(ModuleName, 12, "no aggregate vote")
	ErrUnknownDenom          = sdkerrors.Register(ModuleName, 13, "unknown denom")
	ErrNegativeOrZeroRate    = sdkerrors.Register(ModuleName, 14, "invalid exchange rate; should be positive")
	ErrExistingPrevote       = sdkerrors.Register(ModuleName, 15, "prevote already submitted for this voting period")
	ErrBallotNotSorted       = sdkerrors.Register(ModuleName, 16, "ballot must be sorted before this operation")
	ErrNotImplemented        = sdkerrors.Register(ModuleName, 17, "functon not implemented")
	ErrNoHistoricPrice       = sdkerrors.Register(ModuleName, 18, "no historic price for this denom at this block")
	ErrNoMedian              = sdkerrors.Register(ModuleName, 19, "no median for this denom at this block")
	ErrNoMedianDeviation     = sdkerrors.Register(ModuleName, 20, "no median deviation for this denom at this block")
=======
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
	ErrNotImplemented          = errors.Register(ModuleName, 17, "function not implemented")
	ErrNoHistoricPrice         = errors.Register(ModuleName, 18, "no historic price for this denom at this block")
	ErrNoMedian                = errors.Register(ModuleName, 19, "no median for this denom at this block")
	ErrNoMedianDeviation       = errors.Register(ModuleName, 20, "no median deviation for this denom at this block")
	ErrMalformedLatestAvgPrice = errors.Register(ModuleName, 21, "malformed latest avg price, expecting one byte")
	ErrNoLatestAvgPrice        = errors.Register(ModuleName, 22, "no latest average price")
>>>>>>> 37909a3 (fix: deprecated use of sdkerrors (#1788))
)
