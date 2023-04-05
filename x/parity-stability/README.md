# Parity Stability Module

## Abstract

This document specifies the `x/parity-stabillity` module of the Umee chain.

The parity stability module allows users to swap and redeem stable assets for a stable index Token. This index Token will maintain the parity between underlying assets given a specific configuration.

The parity stability module depends directly on `x/leverage` for supplying, collateralizing, borrowing and reserving 
assets, and the cosmos `x/bank` module as these all affect account balances.

## Contents

1. **[Concepts](#concepts)**
    - [Accepted Assets](#accepted-assets)
    - [Swapping and Redeeming](#swapping-and-redeeming)
    - Important Derived Values:
        - [Incentivizing Displacement](#incentivizing-displacement)
        - [Index Exchange Rate](#index-exchange-rate)
2. **[State](#state)**
3. **[Queries](#queries)**
4. **[Messages](#messages)**
5. **[Update Registry Proposal](#update-registry-proposal)**
6. **[Events](#events)**
7. **[Parameters](#params)**
8. **[EndBlock](#end-block)**
    - [Bad Debt Sweeping](#sweep-bad-debt)
    - [Interest Accrual](#accrue-interest)

## Concepts

### Accepted Assets

At the foundation of the `parity stability` module is the _Index Registry_, which contains a list of accepted assets 
and their percentage distribution across the index.

This list is controlled by governance and an emergency group. Assets that are not in the index registry are not 
available for swapping or redeeming for the index's Token.

Once added to the index registry, assets cannot be removed. In the rare case where an asset would need to be phased 
out, its index percentage can be changed to zero.

### Swapping and Redeeming

Users have the following actions available to them:

- Swap accepted asset for index Token with the [Incentivizing Displacement](#incentivizing-displacement) applied to 
  the initial 1:1 [Index Exchange Rate](#index-exchange-rate) to calculate the index Token amount.
  
  Calculated index Token amount will be minted and transferred to the user's account, meanwhile the accepted asset for 
  the swap will be transferred to the `leverage` module pools and the `leverage` module reserves. The portion to be 
  transferred to each one is determined by the _Index Registry_ configuration of each accepted asset.

- Redeem index Token for accepted asset with the [Incentivizing Displacement](#incentivizing-displacement) applied to
  the initial 1:1 [Index Exchange Rate](#index-exchange-rate) to calculate the asset amount.

  The index Token amount will be withdrawn from the user's account and burned, meanwhile the chosen asset to redeem 
  will be transferred from the `leverage` module to the user's account.

### Derived Values