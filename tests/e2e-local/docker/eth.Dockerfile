FROM ethereum/client-go:v1.10.13

RUN apk add --no-cache curl
COPY eth_genesis.json eth_genesis.json
RUN geth --identity "UmeeTestnet" \
  --nodiscover \
  --networkid 15 init eth_genesis.json

# NOTE:
# - etherbase is where rewards get sent
# - private key for this address is 0xb1bab011e03a9862664706fc3bbaa1b16651528e5f0e7fbfcbfdd8be302a13e7
ENTRYPOINT geth --identity "UmeeTestnet" --nodiscover \
  --networkid 15 \
  --mine \
  --http \
  --http.port "8545" \
  --http.addr "0.0.0.0" \
  --http.corsdomain "*" \
  --http.vhosts "*" \
  --miner.threads=1 \
  --nousb \
  --verbosity=3 \
  --miner.etherbase=0xBf660843528035a5A4921534E156a27e64B231fE \
  --rpc.allow-unprotected-txs
