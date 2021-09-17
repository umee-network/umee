<!-- markdownlint-disable MD013 -->
# Upgrade Instructions

This document contains a rough outline for the steps required to upgrade a node
running on `umee-betanet-1` to `umee-betanet-2`:

1. Stop the `gorc` service (if validator)

    ```bash
    $ systemctl stop gorc
    ```

2. Stop the `geth` service (if validator)

    ```bash
    $ systemctl stop geth
    ```

3. Update `geth` service to run on Rinkeby and restart service (if validator)

    ```bash
    $ vim /etc/systemd/system/geth.service
    # use --rinkeby instead of --goerli
    $ systemctl daemon-reload
    $ systemctl start geth
    ```

4. Stop the `umeed` service

    ```bash
    $ systemctl stop umeed
    ```

5. Reset state (this removes data and other files)

    ```bash
    $ umeed unsafe-reset-all
    ```

6. Remove the old genesis file

    ```bash
    $ rm -fr $HOME/.umee/config/genesis.json
    ```

7. (Optional) If you plan on using different Ethereum or Umee keys for `gorc`, then remove
   `gorc` keystore (since we're redeploying gravity on **Rinkeby**)

    ```bash
    $ rm -fr $HOME/gorc/keystore/**
    ```

8. Update `umeed` binary to version v0.2.0
9. Update `gorc` binary to version v0.2.10+
10. (Optional) Re-import keys into `gorc`
11. Update `gorc` config (if applicable)

    ```bash
    [gravity]
    contract = "0xc846512f680a2161D2293dB04cbd6C294c5cFfA7"
    fees_denom = "uumee"

    [ethereum]
    key_derivation_path = "m/44'/60'/0'/0/0"
    rpc = "..."
    gas_price_multiplier = 1.0

    [cosmos]
    key_derivation_path = "m/44'/118'/0'/0/0"
    grpc = "..."
    prefix = "umee"

    [cosmos.gas_price]
    amount = 0.00001
    denom = "uumee"

    [metrics]
    listen_addr = "127.0.0.1:3000"
    ```

12. Download the new genesis file

    ```bash
    $ wget https://raw.githubusercontent.com/umee-network/umee/main/networks/umee-betanet-2/genesis.json $HOME/.umee/config/genesis.json
    ```

13. Restart the `umeed` service

    ```bash
    $ systemctl start umeed
    ```

14. (Optional) Proceed with validator creation, key delegation, and restarting `geth` and `gorc`
