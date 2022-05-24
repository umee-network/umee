# ADR 009: Liquidity Mining

## Changelog

- May 23 2022: Initial Draft (@toteki)

## Status

Proposed

## Abstract

Umee wishes to add support for liquidity mining incentives; i.e. additional rewards on top of the normal `x/leverage` lending APY for supplying base assets.

For example, a user might "lock" for 14 days some of their `u/ATOM` collateral held in the leverage module, earning an additional 12% APY of the collateral's value, received as `UMEE` tokens.

Locked tokens will be unavailable for `x/leverage` withdrawal until unbonded, but still able to be liquidated. There will be 3 locking tiers, differing in unbonding duration.

Incentive programs will be created by governance proposals, get funded with tokens, then (from `StartDate` to `EndDate`) distribute those tokens to lenders of based on the lenders' locked value and lock tier. APY will vary as fixed reward amounts are divided amongst all participating lenders.

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

All parameters mentioned except amount (see _Funding Programs_) must be set using a governance proposal, which creates the incentive program unless impossible under our chosen implementation.

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

Locked uTokens remain in the `x/leverage` module account, and locked amounts cannot be withdrawn or decollateralized as long as they remain locked.

Locking of funds is independent of active incentive programs, and can even be done in their absence.

```go
type MsgLock struct {
  Lender sdk.AccAddress
  Amount sdk.Coin
  Tier   uint32
}

type MsgUnlock struct {
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

See later sections for reward mechanics - it is mathematically necessary to update pending rewards when updating locked amounts.

On receiving a `MsgUnlock`, the module must perform the following steps:

- Validate tier and uToken amount
- Verify lender has sufficient locked uTokens of the selected tier that are not currently unbonding
- Distribute the lender's current `x/incentive` rewards for the selected denom and tier, if any
- Start an unbonding for the lender in question

Unbondings are defined as a struct:
```
type CollateralUnbonding struct {
  Lender sdk.AccAddress
  Amount sdk.Coin
  Tier   uint32
  End    uint64
}

type CollateralUnbondings struct {
    Unbondings []CollateralUnbonding
}
```

Ongoing unbondings are stored together in state. Concurrent unbondings are added to the end of the slice as they are created.
```
UnbondingPrefix | addr => CollateralUnbondings
```

To avoid spam, a parameter in the module will limit the number of concurrent unbondings per address
```
MaxUnbondings uint32
```

When an unbonding period ends, the restrictions on withdrawing and disabling collateral release, but the `uTokens` remain in the `x/leverage` module account as collateral. Since no transfer occurs at the moment funds are released, there is no need for any queues or checks in `EndBlock`.

Completed unbondings for an address are cleared from state the next time withdraw restrictions are calculated during a transaction - that is, during a `MsgLock`, `MsgUnlock`, `MsgWithdraw`, `MsgSetCollateral` sent by the lender, or a `MsgLiquidate` where the lender is being liquidated.

Any collateral can potentially be seized during `MsgLiquidate`, whether it is locked, unbonding, or unlocked. In the event that the target of liquidation has collateral in various such states, it will be liquidated in this order:
1) Unlocked collateral
2) Locked collateral, starting from the least incentivized tier
3) Unbonding collateral, starting from the least incentivized tier and breaking ties within tiers by choosing the unbondings created first. The siezed collateral is subtracted from the `Amount` of any active unbonding affected.

### Incentive Programs

The `x/incentive` module exists to incentivize users to lock collateral assets at the various tiers.

The basic unit of incentivization is the `IncentiveProgram`, which, continuously between a start time and end time, distributes a fixed amount of a given asset type across all holders of a given denom of locked `uToken` collateral, weighted by amount held and locking tier.

```go
type IncentiveProgram struct {
  ID               uint32
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

TotalRewards is the total amount of RewardDenom to be distributed as a result of the incentive program.

Valid tier weights range from 0 to 1, and `LongTierWeight` is defined as always `1.0`.

For example, a `MiddleTierWeight` of `0.8` means that collateral locked at the middle duration tier accrues 80% of the rewards that the same amount would accrue at the longest tier.

Incentive programs, both active and historical, will be stored in state using protobuf marshaling and a simple uint32 ID once their initial governance proposal passes.

```
IncentiveProgramPrefix | ID => IncentiveProgram
```

### Governance

A single governance proposal type is needed to create new incentive programs.

```go
type IncentiveProgramProposal struct {
	Title       string
	Description string
	Program     IncentiveProgram
}
```

See _Funding Proposals_.

### Reward Math

Our main requirement is to continuously distribute rewards without iterating over addresses.

The following approach is proposed:

- For every user, each nonzero `Locked(address,denom,tier) = Amount` is stored.
- Additionally, each nonzero `TotalLocked(denom,tier) = Amount` is kept up to date in state. These track the sum of uTokens locked across all addresses.
- For each `(denom,tier)` that has ever been incentivized, any nonzero `RewardAccumulator(denom,tier)` is stored in state. This `sdk.DecCoins` represents the total rewards a single locked `uToken` would have accumulated if locked into a given tier at genesis.
- For each user, any nonzero `PendingReward` is stored as `sdk.DecCoins`.
- For each nonzero `Locked(address,denom,tier)`, a value `RewardBasis(address,denom,tier)` is stored which tracks the value of `RewardAccumulator(denom,tier)` at the most recent time rewards were added to `PendingReward` by the given (address,denom,tier).
- At any given `EndBlock`, each active incentive program performs some computations:
  - Calculates the total `RewardDenom` rewards that will be given by the program in the current block `X = program.TotalRewards * (secondsElapsed / program.Duration)`
  - Each lock tier receives a `weightedValue(program,denom,tier) = TotalLocked(denom,tier) * tierWeight(program,tier)`.
  - The amount `X` for each program is then split between the tiers by `weightedValue` into three `X(tier)` values (X1,X2,X3)
  - For each tier, `RewardAccumulator(denom,tier)` is increased by `X(tier) / TotalLocked(denom,tier)`.
- When a user's `Locked(address,denom,tier) = Amount` is updated for any reason, `RewardBasis(address,denom,tier)` is set to the current `RewardAccumulator(denom,tier)` and `PendingRewards(addr)` is increased by the difference of the two, times the user's locked amount. (This uses the locked amount _before_ it is updated.)
- When a user claims rewards, they receive `PendingRewards(addr)` and set it to zero (which clears it from state).

The algorithm above uses an approach similar to [F1 Fee Distribution](https://drops.dagstuhl.de/opus/volltexte/2020/11974/) in that it uses an exchange rate (in our case, HistoricalReward) to track each denom's hypothetical rewards since genesis, and determines actual reward amounts by recording the previous exchange rate (RewardBasis) at which each user made their previous claim.

The F1 algoritm only works if users are forced to claim rewards every time their locked amount increases or decreases (thus, locked amount is known to have stayed constant between any two claims). Our implementation is less complex than F1 because there is no equivalent to slashing  in `x/incentive`, and we move rewards to the `PendingRewards` instead of claiming them directly. The same mathematical effect is achieved, where `Locked(addr,denom,tier)` remains constant in the time `RewardAccumulator(denom,tier)` has increased to its current value from `RewardBasis(denom,tier)`, allowing reward calculation on demand using only those three values.

### Claiming Rewards

There will be a message type which manually claims rewards.

```go
type MsgClaim struct {
  Lender      sdk.AccAddress
}
```

This message type gathers `PendingRewards` by updating `RewardBasis` for each nonzero `Locked(addr,denom,tier)` associated with the lender's address, then claims all pending rewards.

### Funding Programs

Incentive programs are passed via governance, but the funding itself must be placed in the `x/incentive` module account so it can be distributed.

To fund programs, we first require them to pass governance (setting denoms, dates, and tier weights) then assign them an ID, allowing permissionless funding. 

```go
type MsgFundProgram struct {
  Sponsor sdk.AccAddress
  ID      uint32
  Amount  sdk.Coin
}
```

This would allow external incentive programs, once passed, to be funded permissionlessly by the contributing bodies. `MsgFundProgram` should fail if the program has a different `RewardDenom` than input, or if the program is past its start date.

Programs created this way do not have their `ID` or total reward amount known at the time of voting - those fields would be zero when voted on.

Since only approved programs can be funded, we avoid the edge case of returning funds to their sponsors if governance rejects an incentive program proposal.

## Alternative Approaches

- Funding of programs could be done within governance proposals for incentives, if there is a suitable community pool balance to draw from. External incentives, in this case, would require the total rewards (e.g. `300000 ATOM`) to be transferred to the community pool by the third party before voting though, and edge cases (refunding after `no` vote, insufficient funds in pool when proposal passes) must be considered.
- The current approach of passing individual, potentially overlapping programs and letting them run could be replaced with an approach which combines internal/external rewards into the same struct and allows new funding end date extension with a separate governance vote. This would increase complexity compared to isolated programs.

## Consequences

The general purpose of this module is to incentivize liquidity over time and guard against bank runs in the case of black swan events.

The proposed algorithm avoids iteration, at the cost of a moderate amount of complexity and the requirement of a claim transaction.

It allows for external and overlapping incentive programs, but does not provide a means to alter or cancel programs already in progress. 

### Positive
- Locking funds encourages stability and provides our primary defense against bank runs 
- No iteration
- Allows external (non-`UMEE`-denominated) incentive programs to further engourage liquidity

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
