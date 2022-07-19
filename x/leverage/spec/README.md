# Leverage Module

## Abstract

This document specifies the `x/leverage` module of the Umee chain.

The leverage module allows users to supply and borrow assets, and implements various features to support this, such as a token accept-list, a dynamic interest rate module, incentivized liquidation of undercollateralized debt, and automatic reserve-based repayment of bad debt.

The leverage module depends directly on `x/oracle` for asset prices, and interacts indirectly with `x/ibctransfer`, `x/peggy`, and the cosmos `x/bank` module as these all affect account balances.

## Contents

1. **[Concepts](01_concepts.md)**
   - [Accepted Assets](01_concepts.md#Accepted-Assets)
     - [uTokens](01_concepts.md#uTokens)
   - [Supplying and Borrowing](01_concepts.md#Supplying-and-Borrowing)
   - [Reserves](01_concepts.md#Reserves)
   - [Liquidation](01_concepts.md#Liquidation)
   - Important Derived Values:
     - [Adjusted Borrow Amounts](01_concepts.md#Adjusted-Borrow-Amounts)
     - [uToken Exchange Rate](01_concepts.md#uToken-Exchange-Rate)
     - [Supply Utilization](01_concepts.md#Supply-Utilization)
     - [Borrow Limit](01_concepts.md#Borrow-Limit)
     - [Liquidation Limit](01_concepts.md#Liquidation-Limit)
     - [Borrow APY](01_concepts.md#Borrow-APY)
     - [Supplying APY](01_concepts.md#Supplying-APY)
     - [Close Factor](01_concepts.md#Close-Factor)
     - [Total Supplied](01_concepts.md#Total-Supplied)
2. **[State](02_state.md)**
3. **[Queries](03_queries.md)**
4. **[Messages](04_messages.md)**
    - [MsgSupply](04_messages.md#MsgSupply)
    - [MsgWithdraw](04_messages.md#MsgWithdraw)
    - [MsgCollateralize](04_messages.md#MsgCollateralize)
    - [MsgDecollateralize](04_messages.md#MsgDecollateralize)
    - [MsgBorrow](04_messages.md#MsgBorrow)
    - [MsgRepay](04_messages.md#MsgRepay)
    - [MsgLiquidate](04_messages.md#MsgLiquidate)
5. **[EndBlock](05_endblock.md)**
    - [Bad Debt Sweeping](05_endblock.md#Sweep-Bad-Debt)
    - [Interest Accrual](05_endblock.md#Accrue-Interest)
6. **[Events](06_events.md)**
7. **[Parameters](07_params.md)**
