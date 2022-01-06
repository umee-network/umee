# Leverage Module

## Abstract

This document specifies the `x/leverage` module of the Umee chain.

The leverage module allows users to lend and borrow assets, and implements various features to support this, such as a token accept-list, a dynamic interest rate module, incentivized liquidation of undercollateralized debt, and automatic reserve-based repayment of bad debt.

The leverage module depends directly on `x/oracle` for asset prices, and interacts indirectly with `x/ibctransfer`, `x/peggy`, and the cosmos `x/bank` module as these all affect account balances.

## Contents

1. **[Overview](01_overview.md)**
    - [Accepted Assets](01_overview.md#Accepted-Assets)
        - [Token Parameters](01_overview.md#Token-Parameters)
        - [uTokens](01_overview.md#uTokens)
    - [Lending and Borrowing](01_overview.md#Lending-and-Borrowing)
    - [Reserves](01_overview.md#Reserves)
    - [Liquidation](01_overview.md#Liquidation)
    - [Bad Debt](01_overview.md#Bad-Debt)
    - Important Derived Values:
        - [uToken Exchange Rate](01_overview.md#uToken-Exchange-Rate)
        - [Borrow Utilization](01_overview.md#Borrow-Utilization)
        - [Dynamic Interest Rate](01_overview.md#Dynamic-Interest-Rate)
        - [Borrow Limit](01_overview.md#Borrow-Limit)
        - [Borrow APY](01_overview.md#Borrow-APY)
        - [Lending APY](01_overview.md#Lending-APY)
        - [Close Factor](01_overview.md#Close-Factor)
        - [Market Size](01_overview.md#Market-Size)
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
3. **[Queries](03_queries.md)**
    - TODO
4. **[Messages](04_messages.md)**
    - [MsgLendAsset](04_messages.md#MsgLendAsset)
    - [MsgWithdrawAsset](04_messages.md#MsgWithdrawAsset)
    - [MsgSetCollateral](04_messages.md#MsgSetCollateral)
    - [MsgBorrowAsset](04_messages.md#MsgBorrowAsset)
    - [MsgRepayAsset](04_messages.md#MsgRepayAsset)
    - [MsgLiquidate](04_messages.md#MsgLiquidate)
5. **[Periodic Functions](05_periodic.md)**
    - [Bad Debt Sweeping](05_periodic.md#Sweep-Bad-Debt)
    - [Interest Accrual](05_periodic.md#Accrue-Interest)
    - [Exchange Rate Updates](05_periodic.md#Update-Exchange-Rates)
6. **[Events](06_events.md)**
    - [Periodic](06_events.md#Periodic)
    - [Handlers](06_events.md#Handlers)
7. **[Parameters](07_params.md)**
8. **[Module Interactions](08_interactions)**
    - Direct Interactions:
        - [x/Oracle](08_interactions#Oracle)
    - Indirect Interactions:
        - [x/Bank](08_interactions#Bank)
        - [x/IBCtransfer](08_interactions#IBC-Transfer)
        - [x/Peggy](08_interactions#Peggy)

### TODO
- Cover governance, especially UpdateRegistryProposal
- Discuss interactions with `bank`, `ibctransfer` and `peggy`. Specifically the implications of transferring tokens/utokens/collateral at various points, and interactions with `Module Accounts`
- a `Keeper` section
- invariants
- "Lending and borrowing" section
  - lend / withdraw for uTokens
  - mark as collateral (moved to module to can no longer transfer it)
  - borrow / repay
