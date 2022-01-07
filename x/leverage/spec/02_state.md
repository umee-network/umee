# State

The `x/leverage` module keeps the following objects in state:

- Registered Token: `0x01 | denom -> ProtocolBuffer(Token)`
- Borrowed Amount: `0x02 | borrowerAddress | denom -> sdk.Int`
- Collateral Setting: `0x03 | borrowerAddress | denom -> 0x01`
- Collateral Amount: `0x04 | borrowerAddress | denom -> sdk.Int`
- Reserved Amount: `0x05 | denom -> sdk.Int`
- Last Interest Epoch (Unix Time): `0x06 -> sdk.Int`
- uToken Exchange Rate: `0x07 | denom -> sdk.Dec`
- Bad Debt Instance: `0x08 | borrowerAddress | denom -> 0x01`
- Borrow APY: `0x09 | denom -> sdk.Dec`
- Lend APY: `0x0A | denom -> sdk.Dec`

The following serialization methods are used unless otherwise stated:
- `sdk.Dec.Marshal()` and `sdk.Int.Marshal()` for numeric types
- `[]byte(denom) | 0x00 ` for asset and uToken denominations (strings)
- `address.MustLengthPrefix(sdk.Address)` for account addresses

Note that collateral settings and instances of bad debt are both tracked using a value of `0x01`. In both cases, the `0x01` means `true` ("enabled" or "present") and a missing or deleted entry means `false`. No value besides `0x01` is ever stored.

## Token Registry

The `0x01` prefix above allows a governance-controlled `Token Registry` to be stored in state. The token registry is a list of all accepted base asset types and their parameters:

```go
type Token struct {
    BaseDenom            string
    ReserveFactor        sdk.Dec
    CollateralWeight     sdk.Dec
    BaseBorrowRate       sdk.Dec
    KinkBorrowRate       sdk.Dec
    MaxBorrowRate        sdk.Dec
    KinkUtilizationRate  sdk.Dec
    LiquidationIncentive sdk.Dec
    SymbolDenom          string
    Exponent             uint32
}
```
