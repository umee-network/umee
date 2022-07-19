# State

The `x/leverage` module keeps the following objects in state:

- Registered Token: `0x01 | denom -> ProtocolBuffer(Token)`
- Adjusted Borrowed Amount: `0x02 | borrowerAddress | denom -> sdk.Dec`
- Collateral Setting: `0x03 | borrowerAddress | denom -> 0x01`
- Collateral Amount: `0x04 | borrowerAddress | denom -> sdk.Int`
- Reserved Amount: `0x05 | denom -> sdk.Int`
- Last Interest Accrual (Unix Time): `0x06 -> int64`
- Bad Debt Instance: `0x07 | borrowerAddress | denom -> 0x01`
- Interest Scalar: `0x08 | denom -> sdk.Dec`
- Total Borrowed: `0x09 | denom -> sdk.Dec`
- Totak UToken Supply:  `0x0A | denom -> sdk.Int`

The following serialization methods are used unless otherwise stated:
- `sdk.Dec.Marshal()` and `sdk.Int.Marshal()` for numeric types
- `[]byte(denom) | 0x00` for asset and uToken denominations (strings)
- `address.MustLengthPrefix(sdk.Address)` for account addresses
- `cdc.Marshal` and `cdc.Unmarshal` for `gogoproto/types.Int64Value` wrapper around int64

Note that collateral settings and instances of bad debt are both tracked using a value of `0x01`. In both cases, the `0x01` means `true` ("enabled" or "present") and a missing or deleted entry means `false`. No value besides `0x01` is ever stored.

## Adjusted Total Borrowed

Unlike all other quantities in state, `AdjustedTotalBorrowed` values are not present in imported and exported genesis state.

Instead, every time an individual `AdjustedBorrow` is set during `ImportGenesis`, its respective token's `AdjustedTotalBorrowed` is increased by the same amount. Thus, it is indirectly imported as the sum of individual positions.

Similarly, `AdjustedTotalBorrowed` is never set independently during regular operations. It is modified during calls to `setAdjustedBorrow`, always increasing or decreasing by the change in the individual borrow being set.

## Token Registry

The `0x01` prefix above allows a governance-controlled `Token Registry` to be stored in state. The token registry is a list of all accepted base asset types and their parameters:

```go
type Token struct {
    BaseDenom              string
    ReserveFactor          sdk.Dec
    CollateralWeight       sdk.Dec
    LiquidationThreshold   sdk.Dec
    BaseBorrowRate         sdk.Dec
    KinkBorrowRate         sdk.Dec
    MaxBorrowRate          sdk.Dec
    KinkUtilization        sdk.Dec
    LiquidationIncentive   sdk.Dec
    SymbolDenom            string
    Exponent               uint32
    EnableMsgSupply        bool
    EnableMsgBorrow        bool
    Blacklist              bool
    MaxCollateralShare     sdk.Dec
    MaxSupplyUtilization   sdk.Dec
    MinCollateralLiquidity sdk.Dec
}
```
