# Queries

The following queries are available on the leverage module:

General queries:
- **Registered Tokens** returns the entire [Token Registry](02_state.md#Token-Registry)
- **Params** returns the module's current [parameters](07_params.md)
- **Liquidation Targets** queries a list of all borrowers eligible for liquidation

Queries on accepted asset types:
- **Borrow APY** queries for the [Borrow APY](01_concepts.md#Borrow-APY) of a specified denomination.
- **Lend APY** queries for the [Lending APY](01_concepts.md#Lending-APY) of a specified denomination.
- **Reserve Amount** queries for the amount reserved of a specified denomination.
- **Exchange Rate** queries the [uToken Exchange Rate](01_concepts.md#uToken-Exchange-Rate) of a given uToken denomination.
- **Market Size** queries the [Market Size](01_concepts.md#Market-Size) of a specified denomination

Queries on account addresses:
- **Borrowed** queries for the borrowed amount of a user by token denomination. If the denomination is not supplied, the total for each borrowed token is returned.
- **Collateral Setting** queries a borrower's collateral setting of a specified uToken denomination.
- **Collateral** queries the collateral amount of a user by token denomination. If the denomination is not supplied, the total for each collateral token is returned.
- **Borrow Limit** queries the [Borrow Limit](01_concepts.md#Borrow-Limit) in USD of a given borrower.