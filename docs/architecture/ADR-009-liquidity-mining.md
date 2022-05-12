# ADR 009: Liquidity Mining

## Changelog

- May 1 2022: Initial Draft (@toteki)

## Status

Proposed

## Abstract

Umee wishes to add support for liquidity mining incentives; i.e. additional rewards on top of the normal `x/leverage` lending APY for supplying base assets.

For example, a user might "lock" for 14 days some of their `u/ATOM` collateral held in the leverage module, earning an additional 12% APY of the collateral's value, received as `UMEE` tokens.

Locked tokens will be unavailable for `x/leverage` withdrawal until unbonded, but still able to be liquidated. There will be 3 locking tiers, differing in unbonding duration.

Incentive programs will be created by governance proposals, get funded with tokens, then (from `StartDate` to `EndDate`) distribute those tokens to lenders of based on the lenders' locked value and lock tier. APY will vary as fixed reward amounts are divided amongst all participating lenders.

## Context

- TODO

### Backwards Compatibility

This will introduce a new module as well as new behavior in the `x/leverage` module. The changes will take place during the mainnet `calypso` upgrade.

## Decision

The general approach will be to create an `x/incentive` module with a small surface of interaction with `x/leverage`.

The incentive module will support message types which allow users to lock and begin unbonding uTokens, similar to staking. 

Locked funds must be collateral-enabled uTokens. Locked funds will not leave their original place in the leverage module account. Locking will prevent withdrawal and collateral-disabling (but not liquidation) of the locked uTokens until they are successfully unbonded.

There will be three tiers of locking, differing in their unbonding durations, which may receive differing incentives. The tiers will be of fixed durations and will exist independent of active incentive programs.

The intended structure of what is described as a single incentive program is as follows:
- A fixed sum of a single token denomination
- To be distributed evenly over time between a start date and an end date
- To all addresses which have locked a specified uToken denomination
- Proportional to their total value locked but not currently unbonding
- Then weighted by locking tier
- As calculated at the moment of distribution (each block)

All parameters mentioned above (dates, amounts, and denominations) must be set using a governance proposal, which creates the incentive program unless impossible under our chosen implementation.

No message types to alter or halt incentive programs once voted on are planned, and any number of incentive programs should be capable of being active simultaneously regardless of parameters, including overlapping dates and denominations.

Incentives funding will be stored in the `x/liquidity` module account.

Implementation details will determine the exact method of funding incentive programs and distributing rewards, with a priority being the avoidance of iteration over lenders, especially passively.

## Detailed Design

The `x/incentive` module will have a fixed number (3) of lock tiers, which will be the same for all asset classes at any given time.  Typical tiers might be 1,7, and 14 days.

The module will govern the lock durations of the tiers (in seconds) using parameters:
```
    LockDurationShort uint64
    LockDurationMedium uint64
    LockDurationLong uint64
```

### Locking and Unlocking

Users must be able to lock `uTokens` they have enabled as collateral in `x/leverage` by sumbitting new `x/incentive` message types.

Locking of funds is independent of active incentive programs, and can even be done in their absence.

```go
type MsgLockAssets struct {
  Lender sdk.AccAddress
  Amount sdk.Coin
  Tier   uint32
}

type MsgUnlockAssets struct {
  Lender sdk.AccAddress
  Amount sdk.Coin
  Tier   uint32
}
```

Amounts are `uToken` balances, exact integers which will not experience rounding errors.

On receiving a `MsgLock`, the module must perform the following steps:

- Validate tier and uToken amount
- Verify lender has sufficient unlocked uTokens
- Distribute the lender's current `x/incentive` rewards for the selected denom and tier, if any
- Record the new locked utoken amount for the selected denom and tier

See later sections for reward mechanics - it is mathematically necessary to claim rewards when updating amount.

On receiving a `MsgUnlock`, the module must perform the following steps:

- Validate tier and uToken amount
- Verify lender has sufficient locked uTokens of the selected tier that are not currently unbonding
- Distribute the lender's current `x/incentive` rewards for the selected denom and tier, if any
- Start an unbonding for the lender in question

### Incentive Programs

The `x/incentive` module exists to incentivize users to lock collateral assets at the various tiers.

The basic unit of incentivization is the `IncentiveProgram`, which, continuously between a start time and end time, distributes a fixed amount of a given asset type across all holders of a given denom of locked `uToken` collateral, weighted by amount held and locking tier.

```go
type IncentiveProgram struct {
  LockedDenom      string
  RewardDenom      string
  TotalRewards     sdk.Int
  StartTime        uint64
  Duration         uint64
  MiddleTierWeight sdk.Dec
  ShortTierWeight  sdk.Dec
}
```

Start time is unix time, and duration is measured in seconds.

TotalIncentive is the total amount of rewards to be distributed as a result of the incentive program. It can contain any reward denomination, allowing for external incentive programs.

Valid tier weights range from 0 to 1, and `LongTierWeight` is defined as always `1.0`.

For example, a `MiddleTierWeight` of `0.8` means that collateral locked at the middle duration tier accrues 80% of the rewards that the same amount would accrue at the longest tier.

### Reward Math

Our main requirement is to continuously distribute rewards without iterating over addresses.

The following approach is proposed:

- For every user, each nonzero `Locked(address,LockedDenom,tier) = Amount` is stored in state.
- Additionally, each nonzero `TotalLocked(LockedDenom,tier) = Amount` is kept up to date in state.
- For each `(LockedDenom,tier)` that has ever been incentivized, any nonzero `HistoricalReward(LockedDenom,tier)` is stored in state. This `sdk.DecCoins` represents the total rewards a single locked `uToken` would have accumulated if locked into a given tier at genesis.
- At any given `EndBlock`, each active incentive program performs some computations:
  - Calculates the total `RewardDenom` rewards that will be given by the program in the current block `X = program.TotalRewards * (secondsElapsed / program.Duration)`
  - Each lock tier receives a `weightedValue(program,LockedDenom,tier) = TotalLocked(LockedDenom,tier) * tierWeight(program,tier)`.
  - The amount `X` for each program is then split between the tiers by `weightedValue` into three `X(tier)` values (X1,X2,X3)
  - For each tier, `HistoricalReward(LockedDenom,tier)` is increased by `X(tier) / TotalLocked(LockedDenom,tier)` and stored in state
- For every nonzero `Locked(address,LockedDenom,tier) = Amount` stored in state, each nonzero `RewardBasis(address,LockedDenom,tier)` is stored in state as `sdk.Coins`. When a user's `Locked(address,LockedDenom,tier) = Amount` is updated, they automatically claim any current rewards accumulated on that tier. 
- When a user claims rewards, they receive `Y = Locked(address,LockedDenom,tier) * ( HistoricalReward(LockedDenom,tier) - RewardBasis(address,LockedDenom,tier) )` then set `RewardBasis(address,LockedDenom,tier)` to `HistoricalReward(LockedDenom,tier)` (or clear it if locked amount is being set to zero).

The algorithm above uses an approach similar to [F1 Fee Distribution](https://drops.dagstuhl.de/opus/volltexte/2020/11974/) in that it uses an exchange rate (in our case, HistoricalReward) to track each denom's hypothetical rewards since genesis, and determines actual reward amounts by recording the previous exchange rate (RewardBasis) at which each user made their previous claim.

This math only works if users are forced to claim rewards every time their locked amount increases or decreases (thus, locked amount is known to have stayed constant between any two claims). Our implementation is less complex than F1 because there is no equivalent to slashing  in `x/incentive`.

There is no iteration over accounts ever, but storage per locked `(address,LockedDenom,tier)` is a bit large, with an `sdk.Int` for the amount itself and one `sdk.DecCoin` for each `RewardDenom` that has ever been associated with the `LockedDenom` in question.

### Claiming Rewards

Aside from the automatic reward claiming covered above, which can happen at `MsgLock`, `MsgUnlock`, or `x/leverage/MsgLiquidate` (if locked collateral is lost to liquidation), there will also be a message type which manually claims rewards of a given denom.

```go
type MsgClaim struct {
  Lender      sdk.AccAddress
}
```

This message type would claim rewards for all locked tiers and denominations, in all reward denominations.

### Funding Programs

Incentive programs are passed via governance, but the funding itself mustbe placed in the `x/incentive` module account so it can be distributed.

One way to fund incentive programs would be to first require them to pass governance (setting denoms, dates, and tier weights) then assign them an ID, allowing permissionless funding. 

```go
type MsgFundProgram struct {
  Source    sdk.AccAddress
  ProgramID uint32
  Amount    sdk.Coin
}
```

This would allow external incentive programs, once passed, to be funded permissionlessly by the contributing bodies.

However, any programs created this way could not have their total reward amount known at the time of governance - the `TotalRewards` field would still be zero when voted on. It could be omitted from the `IncentiveProgram` struct in that case.

This also avoids the edge case of refunding rejected incentive proposals, which would be possible if programs needed to be funded before voting.

### TODO

- TODO: Governance
- TODO: Full proto definitions?
- TODO: Unbonding mechanics including liquidation interruption
- TODO: leverage interactions
- TODO: Unbonding struct, and queues too.

## Alternative Approaches

- TODO: Funding before or after gov proposal, permissioned or permissionless
- TODO: Do we need unboding / lock durations at all, now that epoch is gone? Still valuable against bank runs.

## Consequences

The proposed algorithm avoids iteration, at the cost of a moderate amount of complexity and the requirement of a claim transaction.

It allows for external and overlapping incentive programs, but does not provide a means to alter or cancel programs already in progress.

### Positive
- No iteration
- Allows external (non-`uumee`-denominated) incentive programs

### Negative
- Changing lock tier durations after launch would blindside existing users
- Rewards must be claimed via claim transaction (or automatically on other action)
- Fairly complex

### Neutral
- Lock tiers are not adjustable per program
- Rewards are liquid (not locked)
- Allows overlapping incentive programs in the same denom

## References

- [F1 Fee Distribution](https://drops.dagstuhl.de/opus/volltexte/2020/11974/)
