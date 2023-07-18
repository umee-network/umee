# Incentive Module

## Abstract

This document specifies the `x/incentive` module of the Umee chain.

The incentive module allows users to `Bond` collateral `uTokens` from the `x/leverage` module, and governance to create and fund `Incentive Programs` which distribute rewards to users with bonded uTokens over time.

Users can `Unbond` tokens over a period of time, after which they can be withdrawn. `UnbondingDuration` is a module parameter (the same for all tokens) set by governance. Typical values might be `1 day`, `7 days`, or `0 days` (instant unbonding).

The incentive module depends on the `x/leverage` module for information about users' bonded collateral, and also requires that the leverage module prevent bonded or currently unbonding collateral from being withdrawn.
There are also a few more advanced interactions, such as instantly unbonding collateral when it is liquidated, and registering an `exponent` when a uToken denom is incentivized for the first time.

## Contents

1. **[Concepts](#concepts)**
   - [Bonding Collateral](#bonding-collateral)
   - [Incentive Programs](#incentive-programs)
   - [Claiming Rewards](#claiming-rewards)
   - [Reward Accumulators](#reward-accumulators)
   - [Reward Trackers](#reward-trackers)
2. **[State](#state)**
3. **[Messages](#messages)**

## Concepts

### Bonding Collateral

A user can bond their `x/leverage` collaterized `uTokens` in a `x/incentive` module to receive extra rewards.

Bonding prevents the user from using any `leverage.MsgDecollateralize` or `leverage.MsgWithdraw` which would reduce the user's collateral below the bonded amount.

**Example:** a user has `100 u/UMEE` in their wallet and `50 u/UMEE` collateral in the leverage module. `40 u/UMEE` from that `50 u/UMEE` is bonded in the incentive module. Their maximum `leverage.MsgDecollateralize` allowed by their bond is `10 u/UMEE` and their maximum `leverage.MsgWithdraw` is `110u/UMEE`.

Bonded collateral is eligible for incentive program rewards as long as it is not currently unbonding.

When the user starts unbonding a uToken, the module's `UnbondingDuration` determines the time after which the tokens are unlocked to the user.

Unbonding uTokens are not eligible for incentive rewards while they unbond, but are still subject to the same restrictions on `leverage.MsgWithdraw` and `leverage.MsgDecollateralize` as bonded tokens.

For example, if a user has `10u/UMEE` bonded and `3u/UMEE` more unbonding, out of a total of `20u/UMEE` collateral, then their current max withdraw is `7u/UMEE`, and they are earning incentive rewards on only `10u/UMEE` uTokens.

The module parameter `MaxUnbondings` limits how many concurrent unbondings a user can have of the same uToken denom, to prevent spam.

Additionally, `MsgEmergencyUnbond` can instantly unbond collateral, starting with in-progress unbondings then bonded tokens.
This costs a fee - for example, if the parameter `EmergencyUnbondFee` is `0.01`, then 1% of the uTokens unbonded would be donated to the `x/leverage` module reserves while the other 99% are returned to the user.

### Incentive Programs

An `IncentiveProgram` is a fixed-duration program which distributes a predetermined amount of one reward token to users which have bonded selected uTokens during its duration.

For example, the following incentive program would, at each block during the `864000 seconds` after its start at `unix time 1679659746`, distribute a portion of its total `1000 UMEE` rewards to users who have bonded (but are not currently unbonding) `u/uumee` to the incentive module.

```go
ip := IncentiveProgram {
   StartTime: 1679659746, // Mar 24 2023
   Duration: 864000, // 10 days
   UToken: "u/uumee",
   TotalRewards: sdk.Coin{"uumee",1000_000000},
}
```

Reward distribution math is

- **Constant Rate:** Regardless of how much `u/uumee` is bonded to the incentive module at any given block, the program distributes the fraction `blockTime / duration` of its total rewards across users every block it is active (with some corrective rounding).
- **Weighted by Amount:** A user with `200u/uumee` bonded received twice as many rewards on a given block as a user with `100u/uumee` bonded.

When multiple incentive programs are active simultaneously, they compute their rewards independently.

Additionally, if no users are bonded while it is active, a program refrains from distributing any rewards until bonded users exist.

For example, an incentive program which saw no bonded users for the first 25% of its duration would distribute 100% of its rewards over the remaining 75% duration.

If no users are bonded at the end of an incentive program, it will end with nonzero `RemainingRewards` and those rewards will stay in the module balance.

### Claiming Rewards

A user can claim rewards for all of their bonded uTokens at once using `MsgClaim`. When a user claims rewards, an appropriate amount of tokens are sent from the `x/incentive` module account to their wallet.

There are also times where rewards must be claimed automatically to maintain rewards-tracking math. These times are:

- On `MsgBond`
- On `MsgBeginUnbonding`
- When a `leverage.MsgLiquidate` forcefully reduces the user's bonded collateral

Any of the actions above cause the same tokens to be transferred to the user as would have been generated by a `MsgClaim` at the same moment.

By automatically claiming rewards whenever a user's bonded amount changes, the module guarantees the following invariant:

> Between any two consecutive reward claims by an account associated with a specific bonded uToken denom, the amount of bonded `uTokens` of the given denom for that account remained constant.

### Reward Accumulators

The incentive module must calculate rewards each block without iterating over accounts and past incentive programs. It does so by storing a number of `RewardAccumulator`

```go
ra := RewardAccumulator{
   denom: "u/uumee",
   exponent: 6,
   rewards: "0.00023uumee, 0.00014ibc/1234",
}
```

The example reward accumulator above can be interpreted as:

> If `10^6` `u/uumee` was bonded at genesis and had remained bonded since, it would have accumulated `0.00023uumee` and `0.00014ibc/1234` in rewards.

The incentive module must store one `RewardAccumulator` for each uToken denom that is currently (or has been previously) incentivized.

During `EndBlock` when rewards are being distributed by incentive programs, the module divides the amount of tokens to distribute by the current `TotalBonded` which it stores for each uToken denom, to increment the `rewards` field in each `RewardAccumulator`.

The `exponent` field is based on the exponent of the base asset associated with the uToken in `denom`. It reduces the loss of precision that may result from dividing the (relatively small) amount of rewards distributed in a single block by the (relatively large) total amount of uTokens bonded.

Note that the `rewards` field must never decrease, nor can any nonzero `RewardAccumulator` be deleted, even after associated incentive programs have ended. The `exponent` field should also never be changed.

### Reward Trackers

`RewardTracker` is stored per-user, and is used in combination with `RewardAccumulator` to calculate rewards since the user's last claim.

```go
rt := RewardTracker{
   account: "umee1s84d29zk3k20xk9f0hvczkax90l9t94g72n6wm",
   denom: "u/uumee",
   rewards: "0.00020uumee, 0.00014ibc/1234",
}
```

The example reward tracker above can be interpreted as:

> The last time account `umee1s84d29zk3k20xk9f0hvczkax90l9t94g72n6wm` claimed rewards for bonded `u/uumee`, the value of `RewardAccumulator` (not tracker) for that denom was `"0.00020uumee, 0.00014ibc/1234"`. Therefore, the rewards to claim this time should be based on the _increase_ since then.

Because the amount of bonded uTokens for this user was constant between the previous `RewardTracker` increase and the current moment, the following simple calculation determines rewards:

> rewards to claim = (RewardAccumulator - RewardTracker) * (AmoundBonded / 10^Exponent)

Whenever an account claims rewards (including the times when rewards are claimed automatically due to `MsgBond`, `MsgBeginUnbonding`, or `leverage.MsgLiquidate`), the account's `RewardTracker` of the affected denom must be updated to the current value of `RewardAccumulator`.

A `RewardTracker` can also be deleted from the module's store once an account's bonded amount has reached zero for a uToken `denom`. The tracker will be initialized to the accumulator's current value if the account decides to bond again later.

## State

The incentive module stores the following in its KVStore and genesis state:

- `Params`
- Every upcoming, ongoing, and completed `IncentiveProgram`
- The next available program `ID` (an integer)
- The last unix time rewards were calculated
- `RewardAccumulator` of every bonded uToken denom, unless zero
- `RewardTracker` for each user associated with nonzero bonded uTokens with a nonzero a reward accumulator above
- All nonzero bonded uTokens amounts for each account (excldes unbondings)
- All unbondings in progress for each account

Additionally, some mathematically redundant information is maintained in `KVStore` but not genesis state for efficient operations:

- `TotalBonded` for every bonded uToken denom (total excludes unbondings).
- `TotalUnbonding` for every bonded uToken denom.
- `AmountUnbonding` for each account with one or more unbondings in progress.

These totals are kept in sync with the values they track by the functions in `keeper/store.go`, which update the totals whenever any of the values they reference are changed for any reason (including when importing genesis state).

## Messages

See [proto](https://github.com/umee-network/umee/blob/main/proto/umee/incentive/v1/tx.proto) for detailed fields.

Permissionless:

- `MsgClaim` Claims any pending rewards for all bonded uTokens
- `MsgBond` Bonds uToken collateral (and claims pending rewarda)
- `MsgBeginUnbonding` Starts unbonding uToken collateral (and claims pending rewards)
- `MsgEmergencyUnbond` Instantly unbonds uToken collateral for a fee based on amount unbonded (and claims pending rewards)
- `MsgSponsor` Funds an entire incentive program with rewards, if it has been passed by governance but not yet funded.

Governance controlled:

- `MsgGovSetParams` Sets module parameters
- `MsgGovCreatePrograms` Creates one or more incentive programs. These programs can be set for community funding or manual funding in the proposal.
