package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Denomination constants
const (
	AtomDenom       string = "ATOM"
	UmeeDenom       string = "UMEE"
	USDDenom        string = "USD"
	BlocksPerMinute        = uint64(10)
	BlocksPerHour          = BlocksPerMinute * 60
	BlocksPerDay           = BlocksPerHour * 24
	BlocksPerWeek          = BlocksPerDay * 7
	BlocksPerMonth         = BlocksPerDay * 30
	BlocksPerYear          = BlocksPerDay * 365
	MicroUnit              = int64(1e6)
)

type (
	// ExchangeRatePrevote defines a structure to store a validator's prevote on
	// the rate of USD in the denom asset.
	ExchangeRatePrevote struct {
		Hash        VoteHash       `json:"hash"`  // Vote hex hash to protect centralize data source problem
		Denom       string         `json:"denom"` // Ticker name of target fiat currency
		Voter       sdk.ValAddress `json:"voter"` // Voter val address
		SubmitBlock int64          `json:"submit_block"`
	}

	// ExchangeRateVote defines a structure to store a validator's vote on the
	// rate of USD in the denom asset.
	ExchangeRateVote struct {
		ExchangeRate sdk.Dec        `json:"exchange_rate"` // ExchangeRate of Luna in target fiat currency
		Denom        string         `json:"denom"`         // Ticker name of target fiat currency
		Voter        sdk.ValAddress `json:"voter"`         // voter val address of validator
	}

	// VoteHash defines a hash value to hide vote exchange rate
	// which is formatted as hex string in SHA256("{salt}:{exchange_rate}:{denom}:{voter}")
	VoteHash []byte
)
