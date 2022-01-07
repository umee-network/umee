# Events

The leverage module emits the following events:

## Handlers

### MsgLendAsset

| Type     | Attribute Key | Attribute Value  |
| -------- | ------------- | ---------------- |
| lend     | sender        | {lenderAddress}  |
| lend     | amount        | {amount}         |
| message  | module        | leverage         |
| message  | action        | lend             |
| message  | sender        | {lenderAddress}  |

### MsgWithdrawAsset

| Type     | Attribute Key | Attribute Value  |
| -------- | ------------- | ---------------- |
| withdraw | sender        | {lenderAddress}  |
| withdraw | amount        | {amount}         |
| message  | module        | leverage         |
| message  | action        | withdraw         |
| message  | sender        | {lenderAddress}  |

### MsgSetCollateral

| Type           | Attribute Key | Attribute Value   |
| -------------- | ------------- | ----------------- |
| set_collateral | sender        | {borrowerAddress} |
| set_collateral | denom         | {denom}           |
| set_collateral | enable        | {enable}          |
| message        | module        | leverage          |
| message        | action        | set_collateral    |
| message        | sender        | {borrowerAddress} |

### MsgBorrowAsset

| Type    | Attribute Key | Attribute Value   |
| ------- | ------------- | ----------------- |
| borrow  | sender        | {borrowerAddress} |
| borrow  | amount        | {amount}          |
| message | module        | leverage          |
| message | action        | borrow            |
| message | sender        | {borrowerAddress} |

### MsgRepayAsset

| Type    | Attribute Key | Attribute Value   |
| ------- | ------------- | ----------------- |
| repay   | sender        | {borrowerAddress} |
| repay   | amount        | {amount}*         |
| message | module        | leverage          |
| message | action        | repay             |
| message | sender        | {borrowerAddress} |

* Amount successfully repaid may be lower than the amount requested in the message if the original amount would exceed full repayment.

### MsgLiquidate

| Type      | Attribute Key | Attribute Value     |
| --------- | ------------- | ------------------- |
| liquidate | sender        | {liquidatorAddress} |
| liquidate | amount        | {amount}*           |
| liquidate | reward_denom  | {rewardDenom}       |
| message   | module        | leverage            |
| message   | action        | liquidate           |
| message   | sender        | {liquidatorAddress} |

* Amount successfully liquidated may be lower than the amount requested in the message if the original amount exceeds full repayment, exceeds the value of desired collateral rewards, or is otherwise restricted by `CloseFactor`.

## Keeper events

In addition to handlers events, the leverage keeper will produce events from the following functions.

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