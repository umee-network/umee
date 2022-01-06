### TODO:
- Cover governance
- Discuss interactions with `bank`, `ibctransfer` and `peggy`. Specifically the implications of transferring tokens/utokens/collateral at various points, and interactions with `Module Accounts`
- a `Keeper` section?
- invariants
- "Lending and borrowing" section
    - lend / withdraw for uTokens
    - mark as collateral (moved to module to can no longer transfer it)
    - borrow / repay

# Leverage Module

## Abstract

This document specifies the `x/leverage` module of the Umee chain.

TODO: Summary (or even business-level info) here.

## Contents

1. **[Overview](01_overview.md)**
    - [Accepted Assets](01_overview.md#Accepted-Assets)
        - [Token Parameters](01_overview.md#Token-Parameters)
        - [uTokens](01_overview.md#uTokens)
    - [Lending and Borrowing](01_overview.md#Lending-and-Borrowing)
    - [Interest Rate Model](01_overview.md#Interest-Rate-Model)
    - [Reserves](01_overview.md#Reserves)
    - [Liquidation](01_overview.md#Liquidation)
    - [Bad Debt](01_overview.md#Bad-Debt)
    - Important Derived Values:
        - [uToken Exchange Rate](01_overview.md#Exchange-Rate)
        - [Borrow Utilization](01_overview.md#Borrow-Utilization)
        - [Borrow Limit](01_overview.md#Borrow-Limit)
        - [Borrow APY](01_overview.md#Borrow-APY)
        - [Lending APY](01_overview.md#Lending-APY)
        - [Close Factor](01_overview.md#Close-Factor)
2. **[State](02_state.md)**
    - [Token Registry](02_state.md#Token-Registry)
    - [Borrows](02_state.md#Borrows)
    - [Reserves](02_state.md#Reserves)
    - [Collateral Settings](02_state.md#Collateral-Settings)
    - [Collateral Amounts](02_state.md#Collateral-Amounts)
    - [Exchange Rates](02_state.md#Exchange-Rates)
    - [Bad Debt Addresses](02_state.md#Bad-Debt-Addresses)
    - [Borrow APY](02_state.md#Borrow-APY)
    - [Lending APY](02_state.md#Lending-APY)
3. **[Messages](03_messages.md)**
    - [MsgLendAsset](03_messages.md#MsgLendAsset)
    - [MsgWithdrawAsset](03_messages.md#MsgWithdrawAsset)
    - [MsgSetCollateral](03_messages.md#MsgSetCollateral)
    - [MsgBorrowAsset](03_messages.md#MsgBorrowAsset)
    - [MsgRepayAsset](03_messages.md#MsgRepayAsset)
    - [MsgLiquidate](03_messages.md#MsgLiquidate)
4. **[Periodic Functions](04_periodic.md)**
    - [Bad Debt Sweeping](04_periodic.md#Sweep-Bad-Debt)
    - [Interest Accrual](04_periodic.md#Accrue-Interest)
    - [Exchange Rate Updates](04_periodic.md#Update-Exchange-Rates)
5. **[Events](05_events.md)**
    - [Periodic](05_events.md#Periodic)
    - [Handlers](05_events.md#Handlers)
6. **[Parameters](06_params.md)**
7. **[Module Interactions](07_interactions)**
    - Direct Interactions:
        - [x/Oracle](07_interactions#Oracle)
    - Indirect Interactions:
        - [x/Bank](07_interactions#Bank)
        - [x/IBCtransfer](07_interactions#IBC-Transfer)
        - [x/Peggy](07_interactions#Peggy)
