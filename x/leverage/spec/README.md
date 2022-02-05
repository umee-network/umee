# Leverage Module

## Abstract

This document specifies the `x/leverage` module of the Umee chain.

The leverage module allows users to lend and borrow assets, and implements various features to support this, such as a token accept-list, a dynamic interest rate module, incentivized liquidation of undercollateralized debt, and automatic reserve-based repayment of bad debt.

The leverage module depends directly on `x/oracle` for asset prices, and interacts indirectly with `x/ibctransfer`, `x/peggy`, and the cosmos `x/bank` module as these all affect account balances.

## Contents

1. **[Concepts](01_concepts.md)**
    - [Accepted Assets](01_concepts.md#Accepted-Assets)
        - [uTokens](01_concepts.md#uTokens)
    - [Lending and Borrowing](01_concepts.md#Lending-and-Borrowing)
    - [Reserves](01_concepts.md#Reserves)
    - [Liquidation](01_concepts.md#Liquidation)
    - Important Derived Values:
        - [Adjusted Borrow Amounts](01_concepts.md#Adjusted-Borrow-Amounts)
        - [uToken Exchange Rate](01_concepts.md#uToken-Exchange-Rate)
        - [Borrow Utilization](01_concepts.md#Borrow-Utilization)
        - [Borrow Limit](01_concepts.md#Borrow-Limit)
        - [Borrow APY](01_concepts.md#Borrow-APY)
        - [Lending APY](01_concepts.md#Lending-APY)
        - [Close Factor](01_concepts.md#Close-Factor)
        - [Market Size](01_concepts.md#Market-Size)
2. **[State](02_state.md)**
3. **[Queries](03_queries.md)**
4. **[Messages](04_messages.md)**
    - [MsgLendAsset](04_messages.md#MsgLendAsset)
    - [MsgWithdrawAsset](04_messages.md#MsgWithdrawAsset)
    - [MsgSetCollateral](04_messages.md#MsgSetCollateral)
    - [MsgBorrowAsset](04_messages.md#MsgBorrowAsset)
    - [MsgRepayAsset](04_messages.md#MsgRepayAsset)
    - [MsgLiquidate](04_messages.md#MsgLiquidate)
5. **[EndBlock](05_endblock.md)**
    - [Bad Debt Sweeping](05_endblock.md#Sweep-Bad-Debt)
    - [Interest Accrual](05_endblock.md#Accrue-Interest)
6. **[Events](06_events.md)**
7. **[Parameters](07_params.md)**
