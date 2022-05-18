# Messages

## MsgLendAsset

A user lends assets the the module.

```protobuf
message MsgLendAsset {
  string                   lender = 1;
  cosmos.base.v1beta1.Coin amount = 2;
}
```

The message will fail under the following conditions:
- `amount` is not a valid amount of an accepted asset
- `lender` balance is insufficient

## MsgWithdrawAsset

A user withdraws lent assets.

```protobuf
message MsgWithdrawAsset {
  string                   lender = 1;
  cosmos.base.v1beta1.Coin amount = 2;
}
```

The message will fail under the following conditions:
- `amount` is not a valid amount of an accepted asset's corresponding uToken
- The sum of `lender` uToken balance and uToken collateral (if enabled) is insufficient

The following additional failures are only possible for collateral-enabled _uTokens_
- Withdrawing the required uToken collateral would reduce `lender`'s `BorrowLimit` below their total borrowed value
- Borrow value or borrow limit cannot be computed due to a missing `x/oracle` price

## MsgSetCollateral

A user enables or disables a uToken denomination as collateral for their account.

```protobuf
message MsgSetCollateral {
  string borrower = 1;
  string denom = 2;
  bool   enable = 3;
}
```

The message will fail under the following conditions:
- `denom` is not a valid uToken

The following additional failures are only possible for collateral-enabled _uTokens_
- Disabling the required _uTokens_ as collateral would reduce `borrower`'s `BorrowLimit` below their total borrowed value
- Borrow value or borrow limit cannot be computed due to a missing `x/oracle` price

## MsgBorrowAsset

A user borrows base assets from the module.

```protobuf
message MsgBorrowAsset {
  string                   borrower = 1;
  cosmos.base.v1beta1.Coin amount = 2;
}
```

The message will fail under the following conditions:
- `amount` is not a valid amount of an accepted asset
- Borrowing the requested amount would cause `borrower` to exceed their `BorrowLimit`
- Borrow value or borrow limit cannot be computed due to a missing `x/oracle` price

## MsgRepayAsset

A user fully or partially repays one of their borrows. If the requested amount would overpay, it is reduced to the full repayment amount before attempting.

```protobuf
message MsgRepayAsset {
  string                   borrower = 1;
  cosmos.base.v1beta1.Coin amount = 2;
}
```

The message will fail under the following conditions:
- `amount` is not a valid amount of an accepted asset
- `borrower` balance is insufficient
- `borrower` has not borrowed any of the specified asset

## MsgLiquidate

A user liquidates all or part of an undercollateralized borrower's borrow positions in exchange for an equivalent value of the borrower's collateral, plus liquidation incentive. If the requested repayment amount would overpay or is limited by available collateral rewards or the dynamic `CloseFactor`, the repayment amount will be reduced to the maximum acceptable value before liquidation is attempted.

The user specifies a minimum reward amount (in a base token denom) that they would accept for the full repayment amount. This is used to compute a ratio of actual repayment (which could be lower than intended) to token equivalent of actual uToken reward. Transactions that would result in a reward:repayment amount lower than the minimum will fail instead.

A minimum reward amount of zero ignores this check and trusts oracle prices.

```protobuf
message MsgLiquidate {
  string                   liquidator = 1;
  string                   borrower = 2;
  cosmos.base.v1beta1.Coin repayment = 3;
  cosmos.base.v1beta1.Coin reward = 4;
}
```

The message will fail under the following conditions:
- `repayment` is not a valid amount of an accepted base asset
- `reward` is not a valid amount of an accepted base asset
- `borrower` has not borrowed any of the specified asset to repay
- `borrower` has no collateral of the requested reward denom
- `borrower`'s total borrowed value does not exceed their `LiquidationThreshold`
- `liquidator` balance is insufficient
- the message's ratio of `reward` to `repayment` is higher than the ratio that would result from liquidation at the current oracle prices and liquidation incentives
- Borrowed value or `LiquidationThreshold` cannot be computed due to a missing `x/oracle` price