package types

import (
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/umee-network/umee/v6/util/sdkutil"
	"gopkg.in/yaml.v3"
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

func (v ExchangeRateTuple) String() string {
	return fmt.Sprintf("{\"denom\":%q, \"exchange_rate\":%q}", v.Denom, sdkutil.FormatDec(v.ExchangeRate))
}

func (v ExchangeRateTuple) MarshalJSON() ([]byte, error) {
	return []byte(v.String()), nil
}

// ExchangeRateTuples - array of ExchangeRateTuple
type ExchangeRateTuples []ExchangeRateTuple

// String implements fmt.Stringer interface
func (tuples ExchangeRateTuples) String() string {
	bz, _ := json.Marshal(tuples)
	return string(bz) // fmt.Sprint([]ExchangeRateTuple(tuples))
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
			return nil, fmt.Errorf("exchange rate can't be negative: %s", tupleStr)
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

func (v ExchangeRate) String() string {
	bz, _ := json.Marshal(v)
	return string(bz)
}
