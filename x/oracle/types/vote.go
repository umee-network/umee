package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"gopkg.in/yaml.v3"

	"github.com/umee-network/umee/v6/x/leverage/types"
)

func NewAggregateExchangeRatePrevote(
	hash AggregateVoteHash,
	voter sdk.ValAddress,
	submitBlock uint64,
) AggregateExchangeRatePrevote {
	return AggregateExchangeRatePrevote{
		Hash:        hash.String(),  // we could store bytes here!
		Voter:       voter.String(), // we could store bytes here!
		SubmitBlock: submitBlock,
	}
}

// String implement stringify
func (v AggregateExchangeRatePrevote) String() string {
	out, _ := yaml.Marshal(v)
	return string(out)
}

func NewAggregateExchangeRateVote(
	exchangeRateTuples ExchangeRateTuples,
	voter sdk.ValAddress,
) AggregateExchangeRateVote {
	return AggregateExchangeRateVote{
		ExchangeRateTuples: exchangeRateTuples,
		Voter:              voter.String(), // we could use bytes here!
	}
}

// String implement stringify
func (v AggregateExchangeRateVote) String() string {
	out, _ := yaml.Marshal(v)
	return string(out)
}

// NewExchangeRateTuple creates a ExchangeRateTuple instance
func NewExchangeRateTuple(denom string, exchangeRate sdk.Dec) ExchangeRateTuple {
	return ExchangeRateTuple{
		denom,
		exchangeRate,
	}
}

// ExchangeRateTuples - array of ExchangeRateTuple
type ExchangeRateTuples []ExchangeRateTuple

// String implements fmt.Stringer interface
func (tuples ExchangeRateTuples) String() string {
	out, _ := yaml.Marshal(tuples)
	return string(out)
}

// ParseExchangeRateTuples ExchangeRateTuple parser
func ParseExchangeRateTuples(tuplesStr string) (ExchangeRateTuples, error) {
	if len(tuplesStr) == 0 {
		return nil, nil
	}

	tupleStrs := strings.Split(tuplesStr, ",")
	tuples := make(ExchangeRateTuples, len(tupleStrs))

	duplicateCheckMap := make(map[string]struct{})
	for i, tupleStr := range tupleStrs {
		denomAmountStr := strings.Split(tupleStr, ":")
		if len(denomAmountStr) != 2 {
			return nil, fmt.Errorf("invalid exchange rate %s", tupleStr)
		}

		decCoin, err := sdk.NewDecFromStr(denomAmountStr[1])
		if err != nil {
			return nil, err
		}
		if !decCoin.IsPositive() {
			return nil, types.ErrInvalidOraclePrice
		}

		denom := strings.ToUpper(denomAmountStr[0])
		tuples[i] = ExchangeRateTuple{
			Denom:        denom,
			ExchangeRate: decCoin,
		}

		if _, ok := duplicateCheckMap[denom]; ok {
			return nil, sdkerrors.ErrInvalidCoins.Wrapf("duplicated denom %s", denom)
		}
		duplicateCheckMap[denom] = struct{}{}
	}

	return tuples, nil
}
