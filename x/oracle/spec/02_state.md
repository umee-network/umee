<!--
order: 2
-->

# State

## ExchangeRate

An `sdk.Dec` that stores an exchange rate against USD, which is used by the [Leverage](../../leverage/spec/README.md) module.

- ExchangeRate: `0x01<denom_Bytes> -> amino(sdk.Dec)`

## FeederDelegation

An `sdk.AccAddress` (`umee-` account) address of `operator`'s delegated price feeder.

- FeederDelegation: `0x02<valAddress_Bytes> -> amino(sdk.AccAddress)`

## MissCounter

An `int64` representing the number of `VotePeriods` that validator `operator` missed during the current `SlashWindow`.

- MissCounter: `0x03<valAddress_Bytes> -> amino(int64)`

## AggregateExchangeRatePrevote

`AggregateExchangeRatePrevote` containing a validator's aggregated prevote for all denoms for the current `VotePeriod`.

- AggregateExchangeRatePrevote: `0x04<valAddress_Bytes> -> amino(AggregateExchangeRatePrevote)`

```go
// AggregateVoteHash is a hash value to hide vote exchange rates
// which is formatted as hex string in SHA256("{salt}:{exchange rate}{denom},...,{exchange rate}{denom}:{voter}")
type AggregateVoteHash []byte

type AggregateExchangeRatePrevote struct {
	Hash        AggregateVoteHash // Vote hex hash to keep validators from free-riding
	Voter       sdk.ValAddress    // Voter val address
	SubmitBlock int64
}
```

## AggregateExchangeRateVote

`AggregateExchangeRateVote` containing a validator's aggregate vote for all denoms for the current `VotePeriod`.

- AggregateExchangeRateVote: `0x05<valAddress_Bytes> -> amino(AggregateExchangeRateVote)`

```go
type ExchangeRateTuple struct {
	Denom        string  `json:"denom"`
	ExchangeRate sdk.Dec `json:"exchange_rate"`
}

type ExchangeRateTuples []ExchangeRateTuple

type AggregateExchangeRateVote struct {
	ExchangeRateTuples ExchangeRateTuples // ExchangeRates against USD
	Voter              sdk.ValAddress	  // voter val address of validator
}
```
