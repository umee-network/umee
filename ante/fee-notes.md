# Notes about Gas prices and gas used

NOTES:

- the gas used depends on the current state of the chain
- the gas records below are based on the empty state
- user is charged the gas-limit, even if tx consumed less gas

| operation                                      | gas used |
| :--------------------------------------------- | -------: |
| x/bank send                                    |   31'029 |
| x/group create                                 |   68'908 |
| x/oracle MsgAggregateExchangeRateVote (3 curr) |   66'251 |
| x/oracle MsgAggregateExchangeRateVote (6 curr) |   69'726 |
| default gas limit                              |  200'000 |

## Target price (in USD cent) for x/bank send

| umee price (usd cent) | gas price in uumee | fee (usd cent) |
| --------------------: | -----------------: | -------------: |
|                     2 |                0.2 |         0.0124 |
|                     2 |                0.1 |         0.0062 |
|                     2 |               0.02 |        0.00124 |
|                     5 |                0.2 |          0.031 |
|                     5 |                0.1 |         0.0155 |
|                     5 |               0.02 |         0.0031 |

## Target price (in USD cent) for validator oracle txs (with 6 currencies) per day

There are roughly 10tx / minute and 14400 per day.
Validator has to do 2 tx (prevote and vote) every 5 blocks.
Validator will need to do `5760 = 2*14400/5` tx.
In table below we assume the same price scheme as above, but in the code most likely we will apply a fixed discount (eg 10x).

The prices are indicative. For some transactions (especially oracle) fees can be disabled.
See fee.go file for details.

| umee price (usd cent) | gas price in uumee | fee (usd cent) |
| --------------------: | -----------------: | -------------: |
|                     2 |                0.2 |         161.28 |
|                     2 |                0.1 |          80.64 |
|                     2 |               0.02 |         16.128 |
|                     5 |                0.2 |          403.2 |
|                     5 |                0.1 |          201.6 |
|                     5 |               0.02 |          40.32 |

## Target price (in USD) for default gas limit

| umee price (usd cent) | gas price in uumee | fee (usd cent) |
| --------------------: | -----------------: | -------------: |
|                     2 |                0.2 |           0.08 |
|                     2 |                0.1 |           0.04 |
|                     2 |               0.02 |          0.008 |
|                     5 |                0.2 |            0.2 |
|                     5 |                0.1 |            0.1 |
|                     5 |               0.02 |           0.02 |
