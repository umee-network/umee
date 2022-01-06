## Abstract

The Oracle module provides the Umee blockchain with an up-to-date and accurate price feed of exchange rates of multiple currencies against the USD for the [Leverage](../../leverage/spec/README.md) module.

As price information is extrinsic to the blockchain, the Umee network relies on validators to periodically vote on current exchange rates, with the protocol tallying up the results once per `VotePeriod` and updating the on-chain exchange rates as the weighted median of the ballot.

> Since the Oracle service is powered by validators, you may find it interesting to look at the [Staking](https://github.com/cosmos/cosmos-sdk/tree/master/x/staking/spec/README.md) module, which covers the logic for staking and validators.

> This module is a version of [Terra's Oracle](https://github.com/terra-money/core/tree/main/x/oracle), with a few important modifications.

## Contents

1. **[Concepts](01_concepts.md)**
    - [Voting Procedure](01_concepts.md#Voting-Procedure)
    - [Reward Band](01_concepts.md#Reward-Band)
    - [Slashing](01_concepts.md#Slashing)
    - [Abstaining from Voting](01_concepts.md#Abstaining-from-Voting)
2. **[State](02_state.md)**
    - [ExchangeRate](02_state.md#ExchangeRate)
    - [FeederDelegation](02_state.md#FeederDelegation)
    - [MissCounter](02_state.md#MissCounter)
    - [AggregateExchangeRatePrevote](02_state.md#AggregateExchangeRatePrevote)
    - [AggregateExchangeRateVote](02_state.md#AggregateExchangeRateVote)
3. **[EndBlock](03_end_block.md)**
    - [Tally Exchange Rate Votes](03_end_block.md#Tally-Exchange-Rate-Votes)
4. **[Messages](04_messages.md)**
    - [MsgDelegateFeedConsent](04_messages.md#MsgDelegateFeedConsent)
    - [MsgAggregateExchangeRatePrevote](04_messages.md#MsgAggregateExchangeRatePrevote)
    - [MsgAggregateExchangeRateVote](04_messages.md#MsgAggregateExchangeRateVote)
5. **[Events](05_events.md)**
    - [EndBlocker](05_events.md#EndBlocker)
    - [Handlers](05_events.md#Handlers)
6. **[Parameters](06_params.md)**
