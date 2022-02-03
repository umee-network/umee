# Test Network `umeeverse-party-1`

## Software Information
You can build from the same tag as the network or you can run the docker image (note: you will need to map config files and systemd scripts if you want to use systemd).

### Beta Build
To enable the beta build you just need to set an environment variable before running make build:
```bash
UMEE_ENABLE_BETA=true make build
```

### Version
This testnet will run the beta enabled binary from [v0.7.3](https://github.com/umee-network/umee/tree/v0.7.3)

### Running the Docker Image
```bash
docker run --entrypoint umeed-beta -it us-docker.pkg.dev/umeedefi/stack/node:v0.7.3
```

### Getting x86_64 binaries
You can also just extract the binaries from the docker image (if you are running x86_64) and work with things that way if you prefer.
```bash
export CONTAINER=$(docker create us-docker.pkg.dev/umeedefi/stack/node:v0.7.3)
docker cp "$CONTAINER:/usr/local/bin" umee-bin
```

```bash
$ ls umee-bin
gaiad        hermes        peggo        price-feeder    umeed        umeed-beta
```

## Gravity Contract Information (Goerli)
Gravity Address: [0xB76197AF55D294994Fcec380964131B824132Ec6](https://goerli.etherscan.io/address/0xB76197AF55D294994Fcec380964131B824132Ec6)

UMEE ERC20: [0x850b72fce82e0bccfbe6aaed2db792be5c9e9973](https://goerli.etherscan.io/token/0x850b72fce82e0bccfbe6aaed2db792be5c9e9973)

ATOM ERC20: [0x260edffa7648ddc398b91884d78485612fc6d246](https://goerli.etherscan.io/token/0x260edffa7648ddc398b91884d78485612fc6d246)

## IBC tokens
ATOM DENOM: `ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2`

UMEE DENOM: `ibc/9F53D255F5320A4BE124FF20C29D46406E126CE8A09B00CA8D3CFF7905119728`

## Apps
[Block Explorer](https://explorer-git-stacks-umeeverse-party-1-umee.vercel.app)

Multisig Wallet: Coming soon

## Umee Chain Information
Chain ID: `umeeverse-party-1`

[Peers](umee-peers.txt)

### Umee Keplr Wallet Setup
Note: try doing this from the developer console in the [Block Explorer](#apps)

```javascript
await fetch("https://raw.githubusercontent.com/umee-network/umee/main/networks/umeeverse-party-1/keplr-umee-config.json")
  .then(r => r.json())
  .then(keplr.experimentalSuggestChain.bind(keplr))
  .then(() => keplr.enable('umeeverse-party-1'))
```

### Umee Cosmos API
* `https://api.biplosion.umeeverse-party-1.network.umee.cc`
* `https://api.coppicing.umeeverse-party-1.network.umee.cc`
* `https://api.scurrility.umeeverse-party-1.network.umee.cc`

### Umee Cosmos GRPC
* `grpc.biplosion.umeeverse-party-1.network.umee.cc:443`
* `grpc.coppicing.umeeverse-party-1.network.umee.cc:443`
* `grpc.scurrility.umeeverse-party-1.network.umee.cc:443`

### Umee Tendermint RPC
* `https://rpc.biplosion.umeeverse-party-1.network.umee.cc`
* `https://rpc.coppicing.umeeverse-party-1.network.umee.cc`
* `https://rpc.scurrility.umeeverse-party-1.network.umee.cc`

## Gaia Chain Information
Chain ID: `gaia-umeeverse-party-1`

[Peers](gaia-peers.txt)

### Gaia Keplr Wallet Setup
Note: try doing this from the developer console in the [Block Explorer](#apps)

```javascript
await fetch("https://raw.githubusercontent.com/umee-network/umee/main/networks/umeeverse-party-1/keplr-gaia-config.json")
  .then(r => r.json())
  .then(keplr.experimentalSuggestChain.bind(keplr))
  .then(() => keplr.enable('gaia-umeeverse-party-1'))
```

### Gaia Cosmos API
* `https://api.sphaeralcea.gaia-umeeverse-party-1.network.umee.cc`

### Gaia Cosmos GRPC
* `grpc.sphaeralcea.gaia-umeeverse-party-1.network.umee.cc:443`

### Gaia Tendermint RPC
* `https://rpc.sphaeralcea.gaia-umeeverse-party-1.network.umee.cc`

