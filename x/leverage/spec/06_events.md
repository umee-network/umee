# Events

The leverage module emits the following events:

## Handlers

### MsgSupply

| Type     | Attribute Key | Attribute Value                                 |
| -------- | ------------- | ----------------------------------------------- |
| supply   | sender        | {supplierAddress}                               |
| supply   | amount        | {amount}                                        |
| message  | module        | leverage                                        |
| message  | action        | /umeenetwork.umee.leverage.v1beta1.MsgSupply    |
| message  | sender        | {supplierAddress}                               |

### MsgWithdrawAsset

| Type     | Attribute Key | Attribute Value                                     |
| -------- | ------------- | --------------------------------------------------- |
| withdraw | sender        | {supplierAddress}                                   |
| withdraw | amount        | {amount}                                            |
| message  | module        | leverage                                            |
| message  | action        | /umeenetwork.umee.leverage.v1beta1.MsgWithdrawAsset |
| message  | sender        | {supplierAddress}                                   |

### MsgSetCollateral

| Type           | Attribute Key | Attribute Value                                     |
| -------------- | ------------- | --------------------------------------------------- |
| set_collateral | sender        | {borrowerAddress}                                   |
| set_collateral | denom         | {denom}                                             |
| set_collateral | enable        | {enable}                                            |
| message        | module        | leverage                                            |
| message        | action        | /umeenetwork.umee.leverage.v1beta1.MsgSetCollateral |
| message        | sender        | {borrowerAddress}                                   |

### MsgBorrowAsset

| Type    | Attribute Key | Attribute Value                                   |
| ------- | ------------- | ------------------------------------------------- |
| borrow  | sender        | {borrowerAddress}                                 |
| borrow  | amount        | {amount}                                          |
| message | module        | leverage                                          |
| message | action        | /umeenetwork.umee.leverage.v1beta1.MsgBorrowAsset |
| message | sender        | {borrowerAddress}                                 |

### MsgRepayAsset

| Type    | Attribute Key | Attribute Value                                  |
| ------- | ------------- | ------------------------------------------------ |
| repay   | sender        | {borrowerAddress}                                |
| repay   | amount        | {amount}*                                        |
| message | module        | leverage                                         |
| message | action        | /umeenetwork.umee.leverage.v1beta1.MsgRepayAsset |
| message | sender        | {borrowerAddress}                                |

* Amount successfully repaid may be lower than the amount requested in the message if the original amount would exceed full repayment.

### MsgLiquidate

| Type      | Attribute Key | Attribute Value                                 |
| --------- | ------------- | ----------------------------------------------- |
| liquidate | sender        | {liquidatorAddress}                             |
| liquidate | amount        | {amount}*                                       |
| liquidate | reward_denom  | {rewardDenom}                                   |
| message   | module        | leverage                                        |
| message   | action        | /umeenetwork.umee.leverage.v1beta1.MsgLiquidate |
| message   | sender        | {liquidatorAddress}                             |

* Amount successfully liquidated may be lower than the amount requested in the message if the original amount exceeds full repayment, exceeds the value of desired collateral rewards, or is otherwise restricted by `CloseFactor`.

## Keeper Events

In addition to handlers events, the leverage keeper will produce events from the following functions which may occur during `EndBlock`.

### AccrueAllInterest

| Type             | Attribute Key  | Attribute Value        |
| ---------------- | -------------- | ---------------------- |
| interest_accrual | block_height   | {ctx.BlockHeight}      |
| interest_accrual | unix_time      | {ctx.BlockTime.Unix()} |
| interest_accrual | total_interest | {totalInterest}        |
| interest_accrual | reserved       | {newReserves}          |

Interest accrual emits an event with the current block height and time, as well as total interest accrued across all borrows and the amount of each token added to reserves. Occurs every block.

### FundOracle

| Type        | Attribute Key | Attribute Value |
| ----------- | ------------- | --------------- |
| fund_oracle | amount        | {reward}        |

Oracle rewards sent from the leverage module are tracked at the moment of transfer. Occurs every block.

### RepayBadDebt

| Type           | Attribute Key | Attribute Value     |
| -------------- | ------------- | ------------------- |
| repay_bad_debt | borrower      | {borrowerAddress}   |
| repay_bad_debt | denom         | {denom}             |
| repay_bad_debt | amount        | {amount}            |

Bad debt repayments are tracked by borrower address and the amount successfully repaid.

| Type               | Attribute Key | Attribute Value     |
| ------------------ | ------------- | ------------------- |
| reserves_exhausted | borrower      | {borrowerAddress}   |
| reserves_exhausted | denom         | {denom}             |
| reserves_exhausted | amount        | {amount}            |

Reserve exhaustion is tracked by the address of the last borrower partially repaid, and the remaining borrow amount in the relevant denom.