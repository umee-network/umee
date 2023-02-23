# Oracle Module

## Abstract

The Oracle module provides the Umee blockchain with an up-to-date and accurate price feed of exchange rates of multiple currencies against the USD for the [Leverage](../leverage/README.md) module.

As price information is extrinsic to the blockchain, the Umee network relies on validators to periodically vote on current exchange rates, with the protocol tallying up the results once per `VotePeriod` and updating the on-chain exchange rates as the weighted median of the ballot.

> Since the Oracle service is powered by validators, you may find it interesting to look at the [Staking](https://github.com/cosmos/cosmos-sdk/blob/79b74ff1216b8a07c5c9decedbe09bbd951f6a54/x/staking/README.md) module, which covers the logic for staking and validators.

## Contents

1. **[Concepts](#concepts)**
   - [Voting Procedure](#voting-procedure)
   - [Reward Band](#reward-band)
   - [Slashing](#slashing)
   - [Abstaining from Voting](#abstaining-from-voting)
2. **[State](#state)**
   - [ExchangeRate](#exchangerate)
   - [FeederDelegation](#feederdelegation)
   - [MissCounter](#misscounter)
   - [AggregateExchangeRatePrevote](#aggregateexchangerateprevote)
   - [AggregateExchangeRateVote](#aggregateexchangeratevote)
3. **[End Block](#end-block)**
   - [Tally Exchange Rate Votes](#tally-exchange-rate-votes)
4. **[Messages](#messages)**
5. **[Events](#events)**
6. **[Parameters](#params)**

## Concepts

### Voting Procedure

During each `VotePeriod`, the Oracle module obtains consensus on the exchange rate of multiple denominations against USD specified in `AcceptList` by requiring all members of the validator set to submit a vote for exchange rates before the end of the interval.

Validators must first pre-commit to a set of exchange rates, then in the subsequent `VotePeriod` submit and reveal their exchange rates alongside a proof that they had pre-commited at those prices. This scheme forces the voter to commit to a submission before knowing the votes of others and thereby reduces centralization and free-rider risk in the Oracle.

- Prevote and Vote

  Let `P_t` be the current time interval of duration defined by `VotePeriod` (currently set to 30 seconds, or 5 blocks) during which validators must submit two messages:

  - A `MsgAggregateExchangeRatePrevote`, containing the SHA256 hash of the exchange rates of multiple denominations. A prevote must be submitted for all different denominations specified in `AcceptList`.
  - A `MsgAggregateExchangeRateVote`, containing the salt used to create the hash for the aggregate prevote submitted in the previous interval `P_t-1`.

- Vote Tally

  At the end of `P_t`, the submitted votes are tallied.

  The submitted salt of each vote is used to verify consistency with the prevote submitted by the validator in `P_t-1`. If the validator has not submitted a prevote, or the SHA256 resulting from the salt does not match the hash from the prevote, the vote is dropped.

  For each exchange rate, if the total voting power of submitted votes exceeds 50%, the weighted median of the votes is recorded on-chain as the effective rate for that denomination against USD for the following `VotePeriod` `P_t+1`.

  Exchange rates receiving fewer than `VoteThreshold` total voting power have their exchange rates deleted from the store.

- Ballot Rewards

  After the votes are tallied, the winners of the ballots are determined with `tally()`.

  Voters that have managed to vote within a narrow band around the weighted median are rewarded with a portion of the collected seigniorage. See `k.RewardBallotWinners()` for more details.

### Reward Band

Let `M` be the weighted median, `𝜎` be the standard deviation of the votes in the ballot, and `R` be the RewardBand parameter. The band around the median is set to be `𝜀 = max(𝜎, R/2)`. All valid (i.e. bonded and non-jailed) validators that submitted an exchange rate vote in the interval `[M - 𝜀, M + 𝜀]` should be included in the set of winners, weighted by their relative vote power.

### Reward Pool

The Oracle module's reward pool is composed of any tokens present in its module account. This pool is funded by the `x/leverage` module as portion of interest accrued on borrowed tokens. If there are no tokens present in the Oracle module reward pool during a reward period, no tokens are distributed for that period.

From the Oracle module's perspective, tokens of varied denominations from the `AcceptList` simply appear in the module account at a regular interval.
The interval is the `x/Leverage` module's `InterestEpoch` parameter, e.g. every 100 blocks, which is generally not equal to the Oracle's `VotePeriod` or other parameters.

The reward pool is not distributed all at once, but instead over a period of time, determined by the param `RewardDistributionWindow`, currently set to `5256000`.

### Slashing

> Be sure to read this section carefully as it concerns potential loss of funds.

A `VotePeriod` during which either of the following events occur is considered a "miss":

- The validator fails to submits a vote for **each and every** exchange rate specified in `AcceptList`.

- The validator fails to vote within the `reward band` around the weighted median for one or more denominations.

A `SlashWindow` is a window of time during which validators can miss votes. At the end of this period, the amount of misses are tallied and the proper reward or punishment is carried out.

During every `SlashWindow` (currently set to 7 days), participating validators must maintain a valid vote rate of at least `MinValidPerWindow` (5%), lest they get their stake slashed (currently set to 0.01%).
The slashed validator is automatically temporarily "jailed" by the protocol (to protect the funds of delegators), and the operator is expected to fix the discrepancy promptly to resume validator participation.
If the validator does not unjail, it will remain outside of the active set and delegates will not receive rewards.

`MinValidPerWindow` is currently set to 5%. This means validator must not miss more than 95% of votes in order to be safe from jailing.

### Abstaining from Voting

In Terra's implementation, validators have the option of abstaining from voting. To quote Terra's documentation :

> A validator may abstain from voting by submitting a non-positive integer for the `ExchangeRate` field in `MsgExchangeRateVote`. Doing so will absolve them of any penalties for missing `VotePeriod`s, but also disqualify them from receiving Oracle seigniorage rewards for faithful reporting.

In order to ensure that we have the most accurate exchange rates, we have removed this feature. Non-positive exchange rates in `MsgAggregateExchangeRateVote` are instead dropped.

The control flow for vote-tallying, exchange rate updates, ballot rewards and slashing happens at the end of every `VotePeriod`, and is found at the [end-block ABCI](#end-block) function rather than inside message handlers.

## State

### ExchangeRate

An `sdk.Dec` that stores an exchange rate against USD, which is used by the [Leverage](../leverage/README.md) module.

- ExchangeRate: `0x01 | byte(denom) -> sdk.Dec`

### FeederDelegation

An `sdk.AccAddress` (`umee-` account) address for `operator` price feeder rewards.

- FeederDelegation: `0x02 | byte(valAddress length) | byte(valAddress) -> sdk.AccAddress`

### MissCounter

An `int64` representing the number of `VotePeriods` that validator `operator` missed during the current `SlashWindow`.

- MissCounter: `0x03 | byte(valAddress length) | byte(valAddress) -> ProtocolBuffer(uint64)`

### AggregateExchangeRatePrevote

`AggregateExchangeRatePrevote` containing a validator's aggregated prevote for all denoms for the current `VotePeriod`.

- AggregateExchangeRatePrevote: `0x04 | byte(valAddress length) | byte(valAddress) -> ProtocolBuffer(AggregateExchangeRatePrevote)`

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

### AggregateExchangeRateVote

`AggregateExchangeRateVote` containing a validator's aggregate vote for all denoms for the current `VotePeriod`.

- AggregateExchangeRateVote: `0x05 | byte(valAddress length) | byte(valAddress) -> ProtocolBuffer(AggregateExchangeRateVote)`

```go
type ExchangeRateTuple struct {
    Denom           string  `json:"denom"`
    ExchangeRate    sdk.Dec `json:"exchange_rate"`
}

type ExchangeRateTuples []ExchangeRateTuple

type AggregateExchangeRateVote struct {
    ExchangeRateTuples  ExchangeRateTuples  // ExchangeRates against USD
    Voter               sdk.ValAddress      // voter val address of validator
}
```

## End Block

### Tally Exchange Rate Votes

At the end of every block, the `Oracle` module checks whether it's the last block of the `VotePeriod`. If it is, it runs the [Voting Procedure](#voting-procedure):

1. All current active exchange rates are purged from the store

2. Received votes are organized into ballots by denomination. Votes by inactive or jailed validators are ignored.

3. Exchange rates not meeting the following requirements will be dropped:

   - Must appear in the permitted denominations in `AcceptList`
   - Ballot for rate must have at least `VoteThreshold` total vote power

4. For each remaining `denom` with a passing ballot:

   - Tally up votes and find the weighted median exchange rate and winners with `tally()`
   - Iterate through winners of the ballot and add their weight to their running total
   - Set the exchange rate on the blockchain for that `denom` with `k.SetExchangeRate()`
   - Emit an `exchange_rate_update` event

5. Count up the validators who [missed](#slashing) the Oracle vote and increase the appropriate miss counters

6. If at the end of a `SlashWindow`, penalize validators who have missed more than the penalty threshold (submitted fewer valid votes than `MinValidPerWindow`)

7. Distribute rewards to ballot winners with `k.RewardBallotWinners()`

8. Clear all prevotes (except ones for the next `VotePeriod`) and votes from the store

## Messages

See [oracle tx proto](https://github.com/umee-network/umee/blob/main/proto/umee/oracle/v1/tx.proto#L11) for list of supported messages.

## Events

See [oracle events proto](https://github.com/umee-network/umee/blob/main/proto/umee/oracle/v1/events.proto) for list of supported events.

## Params

See [oracle events proto](https://github.com/umee-network/umee/blob/main/proto/umee/oracle/v1/oracle.proto#L11) for list of module parameters.
