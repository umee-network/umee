# umeeverse-party-1

## Software Information

You can build from the same tag as the network or you can run the docker image (note: you will need to map config files and systemd scripts if you want to use systemd).

### Beta Build
To enable the beta build you just need to set an environment variable before running make build:
```bash
UMEE_ENABLE_BETA=true make build
```

### Version
This testnet will run the beta enabled binary from [v0.7.2](https://github.com/umee-network/umee/tree/v0.7.2)

### Running the Docker Image
```bash
docker run --entrypoint umeed-beta -it us-docker.pkg.dev/umeedefi/stack/node:v0.7.2
```

### Getting x86_64 binaries
You can also just extract the binaries from the docker image (if you are running x86_64) and work with things that way if you prefer.
```bash
export CONTAINER=$(docker create us-docker.pkg.dev/umeedefi/stack/node:v0.7.2)
docker cp "$CONTAINER:/usr/local/bin" umee-bin
```

```bash
$ ls umee-bin
gaiad        hermes        peggo        price-feeder    umeed        umeed-beta
```

## Gravity Contract Information (Goerli)

Gravity Address: [0x568645530B377903e59EeC004C4dc18f94FD4501](https://goerli.etherscan.io/address/0x568645530B377903e59EeC004C4dc18f94FD4501)

## Umee Chain Information

Chain ID: `umeeverse-party-1`

[Peers](umee-peers.txt)


### Umee Cosmos API
 * `https://api.placebo.umeeverse-party-1.network.umee.cc`
 * `https://api.myosynizesis.umeeverse-party-1.network.umee.cc`
 * `https://api.semicupium.umeeverse-party-1.network.umee.cc`

### Umee Cosmos GRPC
 * `grpc.placebo.umeeverse-party-1.network.umee.cc:443`
 * `grpc.myosynizesis.umeeverse-party-1.network.umee.cc:443`
 * `grpc.semicupium.umeeverse-party-1.network.umee.cc:443`

### Umee Tendermint RPC
 * `https://rpc.placebo.umeeverse-party-1.network.umee.cc`
 * `https://rpc.myosynizesis.umeeverse-party-1.network.umee.cc`
 * `https://rpc.semicupium.umeeverse-party-1.network.umee.cc`

## Gaia Chain Information

Chain ID: `gaia-umeeverse-party-1`

[Peers](gaia-peers.txt)

### Gaia Cosmos API
 * `https://api.toot.gaia-umeeverse-party-1.network.umee.cc`

### Gaia Cosmos GRPC
 * `grpc.toot.gaia-umeeverse-party-1.network.umee.cc:443`

### Gaia Tendermint RPC
 * `https://rpc.toot.gaia-umeeverse-party-1.network.umee.cc`

