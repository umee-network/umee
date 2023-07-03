# Cosmwasm

## Example smart contract to interact with umee native modules (leverage and oracle)

- [umee-cosmwasm](https://github.com/umee-network/umee-cosmwasm)

## Cosmwasm Built-in capabilities

- [Built-in capabilities](https://github.com/CosmWasm/cosmwasm/blob/main/docs/CAPABILITIES-BUILT-IN.md) - iterator, staking, stargate, cosmwasm_1_1, cosmwasm_1_2
- Custom capability of umee chain: `umee`

## Allowed native module queries

Queries for all native Umee modules:

- [ugov](https://github.com/umee-network/umee/blob/main/proto/umee/ugov/v1/query.proto)
- [leverage](https://github.com/umee-network/umee/blob/main/proto/umee/leverage/v1/query.proto)
- [oracle](https://github.com/umee-network/umee/blob/main/proto/umee/oracle/v1/query.proto)
- [uibc](https://github.com/umee-network/umee/blob/main/proto/umee/uibc/v1/quota.proto)

JSON input to query the native modules with custom query

```json
{
  "chain": {
    "custom": {
      "leverage_parameters": {}
    }
  }
}
```

JSON input to query the all modules with stargate queries
> You will find more about stargate queries [here](https://github.com/CosmWasm/wasmvm/blob/v1.2.3/types/queries.go#L339-L350)

```json
{
  "chain":{
    "stargate":{
      "path":"/umee.leverage.v1.Query/Params",
      "data": ""
    }
  }
}
```

Example command to execute a query:

```bash
$ umeed q wasm contract-state smart ${json_input}
$ umeed q wasm contract-state smart umee14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9scsdqqx '{"chain":{"custom":{"leverage_params":{}}}}'
```

## Allowed native module transactions

Only [leverage module transactions](https://github.com/umee-network/umee/blob/main/proto/umee/leverage/v1/tx.proto) are allowed. Example JSON input for Umee native module:

```json
{
  "umee": {
    "leverage": {
      "supply": {
        "supplier": "",
        "asset": {
          "denom": "uumee",
          "amount": "123123123"
        }
      }
    }
  }
}
```

Example commands to execute a transaction:

```bash
$ umeed tx wasm execute ${contract_id} ${json_input}
$ umeed tx wasm execute umee14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9scsdqqx '{"umee":{"leverage":{"supply_collateral":{"supplier":"umee1s84d29zk3k20xk9f0hvczkax90l9t94g72n6wm","asset":{"denom":"uumee","amount":"1234"}}}}}'
```
