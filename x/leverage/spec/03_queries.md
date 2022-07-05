# Queries

The following queries are available on the leverage module:

General queries:
- **Registered Tokens** returns the entire [Token Registry](02_state.md#Token-Registry)
- **Params** returns the module's current [parameters](07_params.md)
- **Liquidation Targets** queries a list of all borrowers eligible for liquidation

Queries on accepted asset types:
- **Borrow APY** queries for the [Borrow APY](01_concepts.md#Borrow-APY) of a specified denomination.
- **Supply APY** queries for the [Supplying APY](01_concepts.md#Supplying-APY) of a specified denomination.
- **Reserve Amount** queries for the amount reserved of a specified denomination.
- **Total Collateral** queries for the total collateral amount of a specified uToken denomination.
- **Exchange Rate** queries the [uToken Exchange Rate](01_concepts.md#uToken-Exchange-Rate) of a given uToken denomination.
- **Market Size** queries the [Market Size](01_concepts.md#Market-Size) of a specified denomination.
- **Token Market Size** queries the [Market Size](01_concepts.md#Market-Size) of a specified denomination, but denominated in base tokens instead of USD. This amounts to _total loaned by all suppliers + interest accrued._
- **Market Summary** combines several asset-specifying queries for more efficient frontend access.

Queries on account addresses:
- **Borrowed** queries for the amount of a given token denomination borrowed by a user. If a denomination is not supplied, the total for each borrowed token is returned.
- **BorrowedValue** queries for the USD value of the amount of a given token denomination borrowed by a user. If a denomination is not supplied, the total across all of that user's borrowed tokens is returned.
- **Loaned** queries for the amount  of a given token denomination loaned by a user. If a denomination is not supplied, the total sum of all of that user's loaned tokens is returned.
- **LoanedValue** queries for the USD value of the amount  of a given token denomination loaned by a user. If a denomination is not supplied, the total across all of that user's loaned tokens is returned.
- **Collateral Setting** queries a borrower's collateral setting (enabled or disabled) of a specified uToken denomination.
- **Collateral** queries a user's collateral amount by token denomination. If a denomination is not supplied, the total for each collateral token is returned.
- **CollateralValue** queries a user's collateral value in USD by token denomination. If a denomination is not supplied, the sum over all collateral tokens is returned.
- **Borrow Limit** queries the [Borrow Limit](01_concepts.md#Borrow-Limit) in USD of a given user.
- **Liquidation Limit** queries the [Borrow Limit](01_concepts.md#Liquidation-Limit) in USD of a given user.