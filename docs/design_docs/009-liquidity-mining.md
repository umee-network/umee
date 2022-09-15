# Design Doc 009: Liquidity Mining

## Changelog

- May 23 2022: Initial Draft (@toteki)

## Status

Proposed

## Abstract

Umee wishes to add support for liquidity mining incentives; i.e. additional rewards on top of the normal `x/leverage` APY for supplying base assets.

For example, a user might "lock" some of their `u/ATOM` collateral held in the leverage module for 14 days, earning an additional 12% APY of the collateral's value, received as `UMEE` tokens.

Locked tokens will be unavailable for `x/leverage` withdrawal until unlocked, but will still be able to be liquidated. There will be 3 locking tiers, differing in unlocking duration.

Incentive programs will be created by governance proposals, get funded with tokens, then (from `StartDate` to `EndDate`) distribute those tokens to suppliers of based on the suppliers' locked value and lock tier. APY will vary as fixed reward amounts are divided amongst all participating suppliers.

A message type `MsgSponsor` will allow for internal (multisig account) or external (ordinary wallet) funding of incentive programs with the `RewardDenom` that was set by governance. This message type will not require special permissions to run, only control of the funds to be used.

### Backwards Compatibility

This will introduce a new module as well as new behavior in the `x/leverage` module. The changes will take place during the mainnet `calypso` upgrade.

## Decision

The general approach will be to create an `x/incentive` module with a small surface of interaction with `x/leverage`.

The incentive module will support message types which allow users to lock and unlock uTokens, similar to staking.

Locked funds must be collateral-enabled uTokens. Locked funds will not leave their original place in the leverage module account. Locking will prevent withdrawal and collateral-disabling (but not liquidation) of the locked uTokens until they are successfully unlocked.

There will be three tiers of locking, differing in their unlocking durations, which may receive differing incentives. The tiers will be of fixed durations and will exist independent of active incentive programs.

The intended structure of what is described as a single incentive program is as follows:

- A fixed sum of a single token denomination
- To be distributed evenly over time between a start date and an end date
- To all addresses which have locked a specified uToken denomination
- Proportional to their total value locked but not currently unlocking
- Then weighted by locking tier
- As calculated at the moment of distribution (each block)

Incentive programs are created by governance. No message types to alter or halt incentive programs once voted on are planned, and any number of incentive programs should be capable of being active simultaneously regardless of parameters, including overlapping dates and denominations.

Incentives funding will be stored in the `x/incentive` module account.

## Detailed Design

The `x/incentive` module will have a fixed number (3) of lock tiers, which will be the same for all asset classes at any given time. Typical unlocking durations for the tiers might be 1,7, and 14 days.

The module will govern the lock durations of the tiers (in seconds) using parameters:

```go
LockDurationShort uint64
LockDurationMedium uint64
LockDurationLong uint64
MiddleTierWeight sdk.Dec
ShortTierWeight  sdk.Dec
```

Valid tier weights range from 0 to 1.

For example, a `MiddleTierWeight` of `0.8` means that collateral locked at the middle duration tier accrues 80% of the rewards that the same amount would accrue at the longest tier.

### Locking and Unlocking

Users lock and unlock `uTokens` they have enabled as collateral in `x/leverage` by sumbitting `x/incentive` message types.

Locked uTokens remain in the `x/leverage` module account.

Locking of funds is independent of active incentive programs, and can even be done in their absence.

```go
type MsgLock struct {
  User sdk.AccAddress
  Asset sdk.Coin
  Tier   uint32
}

type MsgUnlock struct {
  User sdk.AccAddress
  Asset sdk.Coin
  Tier   uint32
}
```

Assets are `uToken`, exact integers which will not experience rounding errors.

On receiving a `MsgLock`, the module must perform the following steps:

- Validate tier and uToken amount
- Verify user has sufficient unlocked uTokens
- Distribute the user's current `x/incentive` rewards for the selected denom and tier, if any
- Record the new locked utoken amount for the selected denom and tier

See later sections for reward mechanics - it is mathematically necessary to update pending rewards when updating locked amounts.

On receiving a `MsgUnlock`, the module must perform the following steps:

- Validate tier and uToken amount
- Verify user has sufficient locked uTokens of the selected tier that are not currently unlocking
- Start an unlocking for the user in question

Unlockings are defined as a struct:

```go
type CollateralUnlocking struct {
  User sdk.AccAddress
  Asset sdk.Coin
  Tier   uint32
  End    uint64
}

type CollateralUnlockings struct {
    Unlockings []CollateralUnlocking
}
```

Ongoing unlockings are stored together in state. Concurrent unlockings are added to the end of the slice as they are created.

```go
UnlockingPrefix | addr => CollateralUnlockings
```

To avoid spam, a parameter in the module will limit the number of concurrent unlockings per address

```go
MaxUnlockings uint32
```

When an unlocking period ends, the restrictions on withdrawing and disabling collateral release, but the `uTokens` remain in the `x/leverage` module account as collateral. Since no transfer occurs at the moment funds unlock, there is no need for any action in `EndBlock`.

Completed unlockings for an address are cleared from state the next time withdraw restrictions are calculated during a transaction - that is, during a `MsgLock`, `MsgUnlock`, `MsgWithdraw`, `MsgSetCollateral` sent by the user, or a `MsgLiquidate` where the user is being liquidated.

Any collateral can potentially be seized during `MsgLiquidate`, whether it is locked, unlocking, or unlocked. In the event that the target of liquidation has collateral in various such states, it will be liquidated in this order:

1. Unlocked collateral
2. Locked collateral, starting from the least incentivized tier
3. Unlocking collateral, starting from the least incentivized tier and breaking ties within tiers by choosing the unlockings created first. The seized collateral is subtracted from the `Asset.Denom` of any active unlocking affected.

### Incentive Programs

The `x/incentive` module exists to incentivize users to lock collateral assets at the various tiers.

The basic unit of incentivization is the `IncentiveProgram`, which, continuously between a start time and end time, distributes a fixed amount of a given asset type across all holders of a given denom of locked `uToken` collateral, weighted by amount held and locking tier.

```go
type IncentiveProgram struct {
  ID               uint32
  LockedDenom      string
  RewardDenom      string
  StartTime        uint64
  Duration         uint64
}
```

Start time is unix time, and duration is measured in seconds.

TotalRewards is the total amount of RewardDenom to be distributed as a result of the incentive program.

Incentive programs, both active and historical, will be stored in state using protobuf marshaling and a simple uint32 ID once their initial governance proposal passes. Their total rewards, which are set when a valid program is funded, will also be stored by program ID.

```go
IncentiveProgramPrefix | ID => IncentiveProgram
ProgramRewardsPrefix   | ID => sdk.Coin
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

Note: After Cosmos SDK 0.46, this can be simplified to use a concrete message type only usabel by governance.

See _Funding Proposals_ for additional considerations.

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

The F1 algorithm only works if users are forced to claim rewards every time their locked amount increases or decreases (thus, locked amount is known to have stayed constant between any two claims).
Our implementation is less complex than F1 because there is no equivalent to slashing in `x/incentive`, and we move rewards to the `PendingRewards` instead of claiming them directly.
The same mathematical effect is achieved, where `Locked(addr,denom,tier)` remains constant in the time `RewardAccumulator(denom,tier)` has increased to its current value from `RewardBasis(denom,tier)`, allowing reward calculation on demand using only those three values.

### Claiming Rewards

There will be a message type which manually claims rewards.

```go
type MsgClaim struct {
  User      sdk.AccAddress
}
```

This message type gathers `PendingRewards` by updating `RewardBasis` for each nonzero `Locked(addr,denom,tier)` associated with the user's address, then claims all pending rewards.

### Funding Programs

Incentive programs are passed via governance, but the funding itself must be placed in the `x/incentive` module account so it can be distributed.

To fund programs, we first require them to pass governance (setting denoms, dates, and tier weights) then assign them an ID, allowing permissionless funding.

```go
type MsgFundProgram struct {
  Sponsor sdk.AccAddress
  ID      uint32
  Asset   sdk.Coin
}
```

This would allow external incentive programs, once passed, to be funded permissionlessly by the contributing bodies. `MsgFundProgram` should fail if the program has a different `RewardDenom` than input, or if the program is past its start date.

Programs created this way do not have their `ID` or total reward amount known at the time of voting - those fields would be zero when voted on.

Since only approved programs can be funded, we avoid the edge case of returning funds to their sponsors if governance rejects an incentive program proposal.

## Alternative Approaches

- Funding of programs could be done within governance proposals for incentives, if there is a suitable community pool balance to draw from. External incentives, in this case, would require the total rewards (e.g. `300000 ATOM`) to be transferred to the community pool by the third party before voting.
- The current approach of passing individual, potentially overlapping programs and letting them run could be replaced with an approach which combines internal/external rewards into the same struct.

## Consequences

The general purpose of this module is to incentivize liquidity over time and guard against bank runs in the case of black swan events.

The proposed algorithm avoids iteration, at the cost of a moderate amount of complexity and the requirement of a claim transaction.

It allows for external and overlapping incentive programs, but does not provide a means to alter or cancel programs already in progress.

### Positive

- Locking funds encourages stability and provides our primary defense against bank runs.
- No iteration.
- Allows external (non-`UMEE`-denominated) incentive programs to further engourage liquidity.

### Negative

- Changing lock tier durations after launch would blindside existing users.
- Rewards must be claimed via claim transaction (or automatically on other action).
- Fairly complex.

### Neutral

- Lock tiers are not adjustable per program.
- Rewards are liquid (not locked).
- Allows overlapping incentive programs in the same denom.

## References

- [F1 Fee Distribution](https://drops.dagstuhl.de/opus/volltexte/2020/11974/)
