package types

import (
	"bytes"
	"sort"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// VoteForTally is a convenience wrapper to reduce redundant lookup cost.
type VoteForTally struct {
	Denom        string
	ExchangeRate sdkmath.LegacyDec
	Voter        sdk.ValAddress
	Power        int64
}

// NewVoteForTally returns a new VoteForTally instance.
func NewVoteForTally(rate sdkmath.LegacyDec, denom string, voter sdk.ValAddress, power int64) VoteForTally {
	return VoteForTally{
		ExchangeRate: rate,
		Denom:        denom,
		Voter:        voter,
		Power:        power,
	}
}

// ExchangeRateBallot is a convenience wrapper around a ExchangeRateVote slice.
type ExchangeRateBallot []VoteForTally

// ToMap return organized exchange rate map by validator.
func (pb ExchangeRateBallot) ToMap() map[string]sdkmath.LegacyDec {
	exchangeRateMap := make(map[string]sdkmath.LegacyDec)
	for _, vote := range pb {
		if vote.ExchangeRate.IsPositive() {
			exchangeRateMap[vote.Voter.String()] = vote.ExchangeRate
		}
	}

	return exchangeRateMap
}

// Power returns the total amount of voting power in the ballot.
func (pb ExchangeRateBallot) Power() int64 {
	var totalPower int64
	for _, vote := range pb {
		totalPower += vote.Power
	}

	return totalPower
}

// WeightedMedian returns the median weighted by the power of the ExchangeRateVote.
// CONTRACT: The ballot must be sorted.
func (pb ExchangeRateBallot) WeightedMedian() (sdkmath.LegacyDec, error) {
	if !sort.IsSorted(pb) {
		return sdkmath.LegacyZeroDec(), ErrBallotNotSorted
	}

	if pb.Len() > 0 {
		totalPower := pb.Power()
		var pivot int64
		for _, v := range pb {
			pivot += v.Power
			if pivot >= (totalPower / 2) {
				return v.ExchangeRate, nil
			}
		}
	}

	return sdkmath.LegacyZeroDec(), nil
}

// StandardDeviation returns the standard deviation by the power of the ExchangeRateVote.
func (pb ExchangeRateBallot) StandardDeviation() (sdkmath.LegacyDec, error) {
	if len(pb) == 0 {
		return sdkmath.LegacyZeroDec(), nil
	}

	median, err := pb.WeightedMedian()
	if err != nil {
		return sdkmath.LegacyZeroDec(), err
	}

	sum := sdkmath.LegacyZeroDec()
	ballotLength := int64(len(pb))
	for _, v := range pb {
		func() {
			defer func() {
				if e := recover(); e != nil {
					ballotLength--
				}
			}()
			deviation := v.ExchangeRate.Sub(median)
			sum = sum.Add(deviation.Mul(deviation))
		}()
	}

	variance := sum.QuoInt64(ballotLength)

	standardDeviation, err := variance.ApproxSqrt()
	if err != nil {
		return sdkmath.LegacyZeroDec(), err
	}

	return standardDeviation, nil
}

// Len implements sort.Interface
func (pb ExchangeRateBallot) Len() int {
	return len(pb)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (pb ExchangeRateBallot) Less(i, j int) bool {
	if pb[i].ExchangeRate.LT(pb[j].ExchangeRate) {
		return true
	}
	if pb[i].ExchangeRate.Equal(pb[j].ExchangeRate) {
		return bytes.Compare(pb[i].Voter, pb[j].Voter) < 0
	}
	return false
}

// Swap implements sort.Interface.
func (pb ExchangeRateBallot) Swap(i, j int) {
	pb[i], pb[j] = pb[j], pb[i]
}

// BallotDenom is a convenience wrapper for setting rates deterministically.
type BallotDenom struct {
	Ballot ExchangeRateBallot
	Denom  string
}

// BallotMapToSlice returns an array of sorted exchange rate ballots.
func BallotMapToSlice(votes map[string]ExchangeRateBallot) []BallotDenom {
	b := make([]BallotDenom, len(votes))
	i := 0
	for denom, ballot := range votes {
		b[i] = BallotDenom{
			Denom:  denom,
			Ballot: ballot,
		}
		i++
	}
	sort.Slice(b, func(i, j int) bool {
		return b[i].Denom < b[j].Denom
	})
	return b
}

// Claim is an interface that directs its rewards to an attached bank account.
type Claim struct {
	Power       int64
	Weight      int64
	TokensVoted int64
	Validator   sdk.ValAddress
}

// NewClaim generates a Claim instance.
func NewClaim(power, weight, winCount int64, v sdk.ValAddress) Claim {
	return Claim{
		Power:       power,
		Weight:      weight,
		TokensVoted: winCount,
		Validator:   v,
	}
}

// ClaimMapToSlice returns an array of sorted exchange rate ballots.
func ClaimMapToSlice(claims map[string]Claim) []Claim {
	c := make([]Claim, len(claims))
	i := 0
	for _, claim := range claims {
		c[i] = claim // makes copy
		i++
	}
	sort.Slice(c, func(i, j int) bool {
		return bytes.Compare(c[i].Validator, c[j].Validator) < 0
	})
	return c
}
