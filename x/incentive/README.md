# Incentive Module

## Abstract

This document specifies the `x/incentive` module of the Umee chain.

The incentive module allows users to `Bond` collateral `uTokens` from the `x/leverage` module, and governance to create and fund `Incentive Programs` which distribute rewards to users with bonded uTokens over time. Users can `Unbond` tokens over a period of time, after which they can be withdrawn.

The incentive module depends on the `x/leverage` module for information about users' bonded collateral, and also requires that the leverage module prevent bonded or currently unbonding collateral from being withdrawn.
There are also a few more advanced interactions, such as instantly unbonding collateral when it is liquidated, and registering an `exponent` when a uToken denom is incentivized for the first time.
