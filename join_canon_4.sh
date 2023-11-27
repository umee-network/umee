if [ -z "$1" ]; then
  echo "Please provide a moniker"
  exit 1
fi

read -r -p "Are you sure you want to reset your .umee folder? [y/n]" response
if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]
then
    umeed tendermint unsafe-reset-all
    rm -r $HOME/.umee
    umeed init $1

    content=$(curl https://canon-4.rpc.network.umee.cc/genesis)
    genesis=$(jq '.result.genesis' <<<"$content")

    rm -r $HOME/.umee/config/genesis.json
    echo "$genesis" > $HOME/.umee/config/genesis.json

    # update and replace state sync, seeds

    SNAP_RPC="https://canon-4.rpc.network.umee.cc:443/" ; \
    BLOCK_HEIGHT=630000 ; \
    TRUST_HASH="EB7DF8672C9725C9E65AE055D6EFD0983ADEFB9344C4914226D32E7FDAA569E9" ; \
    PEERS="ee7d691781717cbd1bf6f965dc45aad19c7af05f@canon-4.network.umee.cc:10000,dfd1d83b668ff2e59dc1d601a4990d1bd95044ba@canon-4.network.umee.cc:10001,e25008728d8f800383561d5ce68cff2d6bfc3826@canon-4.network.umee.cc:10002"

    sed -i.bak -e "s/^seeds *=.*/seeds = \"$PEERS\"/" $HOME/.umee/config/config.toml
    sed -i '' 's/enable = false/enable = true/g' $HOME/.umee/config/config.toml
    sed -i '' "s|rpc_servers = \"\"|rpc_servers = \"$SNAP_RPC,$SNAP_RPC\"|g" $HOME/.umee/config/config.toml
    sed -i '' "s/trust_height = 0/trust_height = $BLOCK_HEIGHT/g" $HOME/.umee/config/config.toml
    sed -i '' "s|trust_hash = \"\"|trust_hash = \"$TRUST_HASH\"|g" $HOME/.umee/config/config.toml
    sed -i '' 's|minimum-gas-prices = ""|minimum-gas-prices = "0.01uumee"|g' $HOME/.umee/config/app.toml

    umeed start
else
    exit 1
fi
