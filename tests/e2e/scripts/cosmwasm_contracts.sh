#!/bin/bash

echo "-----------------------"
echo "## Add new CosmWasm CW20 contract"

SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
ARTIFACTS_PATH=$SCRIPTPATH/../artifacts/
CHAIN_ID="umee-local-beta-testnet"

RESP=$(umeed tx wasm store $ARTIFACTS_PATH/cw20_base.wasm --chain-id $CHAIN_ID --from alice --keyring-backend test --gas 100000000 -y)
CODE_ID=$(echo "$RESP" | jq -r '.logs[0].events[1].attributes[-1].value')

echo "* Code id: $CODE_ID"
echo "* Download code"
TMPDIR=$(mktemp -t wasmdXXXXXX)
umeed q wasm code "$CODE_ID" "$TMPDIR"
echo "-----------------------"
echo "## List code"
umeed query wasm list-code --chain-id $CHAIN_ID

echo "-----------------------"
echo "## Create new contract instance"
INIT="{\"decimals\": 2, \"name\":\"CashName\", \"symbol\": \"CAShING\", \"initial_balances\":[{\"address\": \"$(umeed keys show alice --keyring-backend=test -a)\", \"amount\": \"64534\"}]}"
echo "----------INIT: $INIT"
(umeed tx wasm instantiate "$CODE_ID" "$INIT" --admin="$(umeed keys show alice -a --keyring-backend=test)" \
  --from alice --keyring-backend test --amount="100uumee" --label "test-cw20-rafilx" \
  --gas 1000000 -y --chain-id $CHAIN_ID) > $SCRIPTPATH/contract_definition.json


echo "Contract Information"

CONTRACT=$(umeed query wasm list-contract-by-code "$CODE_ID" -o json --chain-id $CHAIN_ID  | jq -r '.contracts[-1]')
echo "* Contract address: $CONTRACT"
echo "### Query all"
RESP=$(umeed query wasm contract-state all "$CONTRACT" -o json --chain-id $CHAIN_ID )
# echo "$RESP" | jq
# echo "### Query smart"
# umeed query wasm contract-state smart "$CONTRACT" '{"verifier":{}}' -o json --chain-id $CHAIN_ID  | jq
# echo "### Query raw"
# KEY=$(echo "$RESP" | jq -r ".models[0].key")
# umeed query wasm contract-state raw "$CONTRACT" "$KEY" -o json --chain-id $CHAIN_ID  | jq

# umee1a8vuh2wk0ugmdcnfsev8szfkcnrkcswgk6qxravwe0x77ee7g6cqw2pzkk
# umeevaloper1zypqa76je7pxsdwkfah6mu9a583sju6xjettez
# umee1zypqa76je7pxsdwkfah6mu9a583sju6xjavygg
# terra1rz5chzn0g07hp5jx63srpkhv8hd7x8pss20w2e

# umeed tx wasm execute umee1nc5tatafv6eyq7llkr2gv50ff9e22mnf70qgjlv737ktmt4eswrq59gzzq "{"transfer": {"recipient": "umeevaloper159yrfahn6n6xnrzqvz7c6dngctsvc5wfua9vsc", "amount": "123"}}" --from alice --keyring-backend test