# Manual Liquidations

Running liquidations on Umee requires four main steps:

- Preparing a liquidator account
- Querying for eligible liquidation targets
- Executing a liquidation transaction
- Selling liquidation rewards for a profit

This document contains notes on how to do these manually - a more sophisticated approach would be to automate one or more steps and monitor the results.

## Preparing a Liquidator Account

Any wallet you control can be used as a liquidator, as long as it has a pool of assets and tokens for gas on the relevant chains.

Here is an example setup:

- `100 AXL-USDC` from axelar collateralized on [app.umee.cc](https://app.umee.cc/) with no active borrows
- `1 UMEE` on Umee
- `1 ATOM` on Cosmos Hub
- `1 AXL` on Axelar
- `1 OSMO` on Osmosis

The main collateral will allow you to borrow any tokens required during liquidation, and the other tokens are for gas on their native chains. (Gas tokens will vary based on which chains and exchanges are being used.)

## Querying for Eligible Liquidation Targets

Umee provides a query to list addresses that are eligible for liquidation:

- REST: `https://api.your.node.here/umee/leverage/v1/liquidation-targets`
- CLI: `umeed q leverage liquidation-targets`

There is also a courtesy endpoint at `https://api.mainnet.network.umee.cc/umee/leverage/v1/liquidation_targets` which provides a cached output for the query every 5 minutes when available.

### Enabling the `liquidation-targets` Query

For performance reasons, this query is disabled by default:

> Error: rpc error: code = Unknown desc = node has disabled liquidator queries

Because this particular query iterates over all borrower accounts on the chain, it should not be enabled on validators or nodes supporting important infrastructure like block explorers.

To enable the liquidation targets query on a node you control, start the node with the `-l` flag:

`umeed start -l`

The following examples were obtained using the CLI against a local umee node with the query enabled.

### Using the Query

When there are no borrowers eligible for liquidation, the return is empty:

```sh
% umeed q leverage liquidation-targets
targets: []
```

If targets are available, they will appear as addresses

```sh
% umeed q leverage liquidation-targets
targets:
- umee1ycgwkvza827rvx7lllsv93ysxegxsky4jeuayt
- umee1l2jv2mym7xd442cmeqka9yvd7vxelsplnn2qn8
```

### Choosing a Target

Select one address and use leverage module queries to learn more about their position

```sh
% umeed q leverage account-summary umee1l2jv2mym7xd442cmeqka9yvd7vxelsplnn2qn8
borrow_limit: "0.862311887319476919"
borrowed_value: "0.981963999713893304"
collateral_value: "1.077889859149346149"
liquidation_threshold: "0.916206380276944227"
supplied_value: "1.077889859149346149"
```

The values above indicate that the borrower is small, with borrowed and supplied values of approximately 1 USD. Their borrowed value (`$0.98`) is higher than their liquidation threshold (`$0.91`) so they can be liquidated.

Next we want to see what tokens they have borrowed, and what collateral tokens of theirs can be claimed as a reward.

```sh
% umeed q leverage account-balances umee1l2jv2mym7xd442cmeqka9yvd7vxelsplnn2qn8
borrowed:
- amount: "83568"
  denom: ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9
collateral:
- amount: "1073682"
  denom: u/ibc/49788C29CD84E08D25CA7BE960BC1F61E88FEFC6333F58557D236D693398466A
supplied:
- amount: "1077906"
  denom: ibc/49788C29CD84E08D25CA7BE960BC1F61E88FEFC6333F58557D236D693398466A
```

Here (and with the help of `umeed q leverage registered-tokens` to interpret ibc denoms) we see that the user has borrowed `0.083 ATOM` with collateral of `1.077 USDC`.

If the borrower had multiple tokens collateralized or borrowed, this would be the time to choose which ones to repay and receive as a reward. In this case, the only option for the liquidator is to repay `ATOM` and receive `USDC`.

## Executing a Liquidation Transaction

Our intent is to repay `0.083 ATOM` on the borrower's behalf, and receive `USDC` in return. However, our liquidator account doesn't currently have `ATOM` on the Umee chain - it only has collateralized `USDC` and some `UMEE` left over for gas.

Rather than keeping every possible repayment token on hand, one strategy is to keep a base amount of collateral on chain, then borrow any tokens required for liquidation right before use.
Then, liquidation rewards can be withdrawn, sold on an exchange, and used to repay the borrowed tokens with some profit left over.

### Preparing Repayment

In the strategy above, the first step is to borrow `0.084 ATOM` (slightly more then the target owes, to ensure we have enough).

```sh
% umeed tx leverage borrow 84000ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9 --from my-key --chain-id umee-1 --gas auto --gas-adjustment 3.0 --gas-prices 0.1uumee -y --broadcast-mode block
```

Appropriate flags are provided, the most imporant of which is `--broadcast-mode block`, which waits for a response to your transaction and displays the result (or error).

Now the required `ATOM` is in our account.

Another option would have been to buy `ATOM` on an exchange, or have an initial balance prepared before we selected our liquidation target.

### Liquidation Transaction

In a liquidation transaction, you need to specify three things:

- target address (`umee1l2jv...`)
- repayment coin (`84000ibc/C4CFF...`, which is `0.084 ATOM`)
- reward token (`ibc/49788...`, which is `USDC`)

```sh
% umeed tx leverage liquidate umee1l2jv2mym7xd442cmeqka9yvd7vxelsplnn2qn8 84000ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9 ibc/49788C29CD84E08D25CA7BE960BC1F61E88FEFC6333F58557D236D693398466A --from my-key --chain-id umee-1 --gas auto --gas-adjustment 10.0 --gas-prices 0.1uumee -y --broadcast-mode block
```

The chain automatically reduces the repayment amount to what the target currently owes, so there is no risk of overpaying. Additionally, the reward amount is calculated using the `liquidation_incentive` of the reward token.

In this case (from `umeed q leverage registered-tokens`) USDC has `"liquidation_incentive": "0.05"` which means the reward will be 105% the value of the tokens repaid.

Repaying `$0.98` worth of borrowed ATOM rewarded us `$1.03` in USDC, out of the borrower's `$1.07` in collateral.

### Partial Liquidation

If the module determines that only part of a borrower's position can be liquidated, then the repayment amount will be automatically reduced as if the liquidator had set a lower maximum repay amount in the transaction. No computations are required on the liquidator's side.

Information on the parameters that determine partial liquidation can be found [here](./README.md#CloseFactor)

## Selling Liquidation Rewards for a Profit

To lock in our profits, we should trade the liquidation rewards we received back into the token denom we repaid.

In our case, we

- Transfer `1.03 USDC` from UMEE to Axelar using IBC
- Transfer `1.03 USDC` from Axelar to Osmosis using IBC (requires `AXL` gas)
- Trade `1.03 USDC` for `0.088 ATOM` on Osmosis
- Transfer `0.088 ATOM` from Osmosis to Cosmos Hub (requires `OSMO` gas)
- Transfer `0.088 ATOM` from Cosmos Hub to Umee (requires `ATOM` gas)
- Repay `0.084001 ATOM` on Umee (our borrow accrued a few minutes of interest)

For reference, the repayment looks like:

```sh
% umeed tx leverage repay 84001ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9 --from my-key --chain-id umee-1 --gas auto --gas-adjustment 3.0 --gas-prices 0.1uumee -y --broadcast-mode block
```

The liquidator account has returned to its initial state, except:

- An extra `0.004 ATOM` in its balance on Umee (profit)
- Paid some `ATOM`, `OSMO`, `AXL`, `UMEE` in gas on their respective chains

Note that most of the effort of liquidation comes from IBC transferring the reward and repayment tokens around.
Each token (`ATOM` and `USDC` in this case) must be sent to its original chain before being transferred to its exchange or back to Umee.

The primary obstacle for automation is to manage those IBC transfers, and keep a wallet on each relevant chain supplied with gas.
