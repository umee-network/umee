<!--
order: 1
-->

# Concepts

## Voting Procedure

During each `VotePeriod`, the Oracle module obtains consensus on the exchange rate of multiple denominations against USD specified in `AcceptList` by requiring all members of the validator set to submit a vote for exchange rates before the end of the interval.

Validators must first pre-commit to a set of exchange rates, then in the subsequent `VotePeriod` submit and reveal their exchange rates alongside a proof that they had pre-commited at those prices. This scheme forces the voter to commit to a submission before knowing the votes of others and thereby reduces centralization and free-rider risk in the Oracle.

* Prevote and Vote

    Let `P_t` be the current time interval of duration defined by `VotePeriod` (currently set to 30 seconds, or 5 blocks) during which validators must submit two messages:

  * A `MsgAggregateExchangeRatePrevote`, containing the SHA256 hash of the exchange rates of multiple denominations. A prevote must be submitted for all different denominations specified in `AcceptList`.
  * A `MsgAggregateExchangeRateVote`, containing the salt used to create the hash for the aggregate prevote submitted in the previous interval `P_t-1`.

* Vote Tally

    At the end of `P_t`, the submitted votes are tallied.

    The submitted salt of each vote is used to verify consistency with the prevote submitted by the validator in `P_t-1`. If the validator has not submitted a prevote, or the SHA256 resulting from the salt does not match the hash from the prevote, the vote is dropped.

    For each exchange rate, if the total voting power of submitted votes exceeds 50%, the weighted median of the votes is recorded on-chain as the effective rate for that denomination against USD for the following `VotePeriod` `P_t+1`.

    Exchange rates receiving fewer than `VoteThreshold` total voting power have their exchange rates deleted from the store.

* Ballot Rewards

    After the votes are tallied, the winners of the ballots are determined with `tally()`.

    Voters that have managed to vote within a narrow band around the weighted median are rewarded with a portion of the collected seigniorage. See `k.RewardBallotWinners()` for more details.

## Reward Band

Let `M` be the weighted median, `𝜎` be the standard deviation of the votes in the ballot, and `R` be the RewardBand parameter. The band around the median is set to be `𝜀 = max(𝜎, R/2)`. All valid (i.e. bonded and non-jailed) validators that submitted an exchange rate vote in the interval `[M - 𝜀, M + 𝜀]` should be included in the set of winners, weighted by their relative vote power.

## Reward Pool

The Oracle module's reward pool is composed of any tokens present in its module account. This pool is funded by the `x/leverage` module as portion of interest accrued on borrowed tokens. If there are no tokens present in the Oracle module reward pool during a reward period, no tokens are distributed.

From the Oracle module's perspective, tokens of varied denominations from the `AcceptList` simply appear in the module account at a regular interval.
The interval is the `x/Leverage` module's `InterestEpoch` parameter, e.g. every 100 blocks, which is generally not equal to the Oracle's `VotePeriod` or other parameters.

The reward pool is not distributed all at once, but instead over a period of time, determined by the param `RewardDistributionWindow`, currently set to `5256000`.
## Slashing

> Be sure to read this section carefully as it concerns potential loss of funds.

A `VotePeriod` during which either of the following events occur is considered a "miss":

* The validator fails to submits a vote for **each and every** exchange rate specified in `AcceptList`.

* The validator fails to vote within the `reward band` around the weighted median for one or more denominations.

During every `SlashWindow`, participating validators must maintain a valid vote rate of at least `MinValidPerWindow` (5%), lest they get their stake slashed (currently set to 0.01%). The slashed validator is automatically temporarily "jailed" by the protocol (to protect the funds of delegators), and the operator is expected to fix the discrepancy promptly to resume validator participation.

## Abstaining from Voting

In Terra's implementation, validators have the option of abstaining from voting. To quote Terra's documentation :

> A validator may abstain from voting by submitting a non-positive integer for the `ExchangeRate` field in `MsgExchangeRateVote`. Doing so will absolve them of any penalties for missing `VotePeriod`s, but also disqualify them from receiving Oracle seigniorage rewards for faithful reporting.

In order to ensure that we have the most accurate exchange rates, we have removed this feature. Non-positive exchange rates in `MsgAggregateExchangeRateVote` are instead dropped.

## Messages

> The control flow for vote-tallying, exchange rate updates, ballot rewards and slashing happens at the end of every `VotePeriod`, and is found at the [end-block ABCI](./03_end_block.md) function rather than inside message handlers.
