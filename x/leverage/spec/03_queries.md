# Queries

The following queries are available on the leverage module:

General queries:
- **Registered Tokens** returns the entire [Token Registry](02_state.md#Token-Registry)
- **Params** returns the module's current [parameters](07_params.md)
- **Liquidation Targets** queries a list of all borrowers eligible for liquidation

Queries on accepted asset types:
- **Market Summary** collects data on a given `Token` denomination. A description of each response field can be found in the [QueryMarketSummaryResponse proto definition](../../../proto/umee/leverage/v1/query.proto)

Queries on account addresses:
- **Account Balances** gets the total supplied, collateralized, and borrowed tokens for an address. A description of each response field can be found in the [QueryAccountBalancesResponse proto definition](../../../proto/umee/leverage/v1/query.proto)
- **Account Summary** calculates the total value supplied, collateralized, and borrowed by an address as well as its borrowing limits. It will fail if the price oracle is down. A description of each response field can be found in the [QueryAccountSummaryResponse proto definition](../../../proto/umee/leverage/v1/query.proto)