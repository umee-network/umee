# GMP Handler

Cross chain transfers using Axelar General Message Passing (GMP)
More info about cosmos GMP find [here](https://docs.axelar.dev/dev/cosmos-gmp)

## Axelar IBC Memo

```go
type Memo struct {
    SourceChain   string `json:"source_chain"`
    SourceAddress string `json:"source_address"`
    Payload       []byte `json:"payload"`
    Type          int64  `json:"type"`
}
```

When user wants to execute leverage supply or supply collateral or liquidate transaction using Axelar GMP , user have to send `Payload=[]byte(Msg)` in `Memo` ,
in below example contract change [payload](https://github.com/axelarnetwork/evm-cosmos-gmp-sample/blob/main/native-integration/multi-send/solidity/contracts/MultiSend.sol) feild and version ,for now we are supporting `uint32(2)` version.

## Supported Leverage Msgs through Axelar GMP

### Supply the asset

```go
type MsgSupply struct {
    // Supplier is the account address supplying assets and the signer of the message.
    Supplier string     `protobuf:"bytes,1,opt,name=supplier,proto3" json:"supplier,omitempty"`
    Asset    types.Coin `protobuf:"bytes,2,opt,name=asset,proto3" json:"asset"`
}
```

### Supply and Collateral Msg

```go
type MsgSupplyCollateral struct {
    // Supplier is the account address supplying assets and the signer of the message.
    Supplier string     `protobuf:"bytes,1,opt,name=supplier,proto3" json:"supplier,omitempty"`
    Asset    types.Coin `protobuf:"bytes,2,opt,name=asset,proto3" json:"asset"`
}
```

### Liquidate Msg

```go
type MsgLiquidate struct {
    // Liquidator is the account address performing a liquidation and the signer
    // of the message.
    Liquidator string `protobuf:"bytes,1,opt,name=liquidator,proto3" json:"liquidator,omitempty"`
    // Borrower is the account whose borrow is being repaid, and collateral consumed,
    // by the liquidation. It does not sign the message.
    Borrower string `protobuf:"bytes,2,opt,name=borrower,proto3" json:"borrower,omitempty"`
    // Repayment is the maximum amount of base tokens that the liquidator is willing
    // to repay.
    Repayment types.Coin `protobuf:"bytes,3,opt,name=repayment,proto3" json:"repayment"`
    // RewardDenom is the denom that the liquidator will receive as a liquidation reward.
    // If it is a uToken, the liquidator will receive uTokens from the borrower's
    // collateral. If it is a base token, the uTokens will be redeemed directly at
    // a reduced Liquidation Incentive, and the liquidator will receive base tokens.
    RewardDenom string `protobuf:"bytes,4,opt,name=reward_denom,json=rewardDenom,proto3" json:"reward_denom,omitempty"`
}
```

## Example transfer from fantom EVM to umee using Axelar GMP

```bash
# clone the repo
$ git clone https://github.com/axelarnetwork/evm-cosmos-gmp-sample.git
$ cd evm-cosmos-gmp-sample/native-integration/multi-send/solidity

# Build the contracts on this repo
$ npm run build

> multisend-solidity@1.0.0 build
> rm -rf artifacts && hardhat compile

Compiled 13 Solidity files successfully
```

## Deploy `MultiSend` contract on fantom evm testnet

> Add below code into `deploy-umee.js`

```js
'use strict'

const {
    providers: { JsonRpcProvider },
    Wallet,
    ContractFactory,
    constants: { AddressZero },
} = require('ethers')


const fantom = {
    name: 'fantom',
    url: 'https://rpc.testnet.fantom.network',
    confirmation: 1,
    privateKey: 'cf469f1c4b06a6204bb9f977fa2865271a17a4ed2028ba4c064fea4754e81c83',
    gateway: '0x97837985Ec0494E7b9C71f5D3f9250188477ae14',
    gasService: '0xbE406F0189A0B4cf3A05C286473D23791Dd44Cc6'
}

const MultiSend = require('./artifacts/contracts/MultiSend.sol/MultiSend.json');

(async () => {
    const wallet = new Wallet(
        fantom.privateKey,
        new JsonRpcProvider(fantom.url),
    );

    console.log("Address: " + wallet.address);
    const factory = ContractFactory.fromSolidity(MultiSend, wallet);

    const contract = await factory.deploy(fantom.gateway, fantom.gasService)
    const tx = await contract.deployed();

    console.log(`multi send contract deployed on ${tx.address}`);
})();
```

### Deploy the contract

```bash
$ node deploy-umee.js 
Address: 0x68B93045fe7D8794a7cAF327e7f855CD6Cd03BB8
multi send contract deployed on 0xf3bF4B57c5cf252BF940C80Df41d02834935FF3b
```

Deployed contract address : `0xf3bF4B57c5cf252BF940C80Df41d02834935FF3b`

## Execute the transfer from fantom evm to umee using axelar gmp

> Add below code into `interact-umee.js`

```js
'use strict'

const {
    providers: { JsonRpcProvider },
    Contract,
    Wallet,
    ethers,
} = require('ethers')


const MultiSend = require('./artifacts/contracts/MultiSend.sol/MultiSend.json');
const IERC20 = require('./artifacts/@axelar-network/axelar-gmp-sdk-solidity/contracts/interfaces/IERC20.sol/IERC20.json');

const tokenAddr = '0x75Cc4fDf1ee3E781C1A3Ee9151D5c6Ce34Cf5C61';
const contract = '0xf3bF4B57c5cf252BF940C80Df41d02834935FF3b';

const fantom = {
    name: 'fantom',
    url: 'https://rpc.testnet.fantom.network',
    confirmation: 1,
    privateKey: 'cf469f1c4b06a6204bb9f977fa2865271a17a4ed2028ba4c064fea4754e81c83',
    gateway: '0x97837985Ec0494E7b9C71f5D3f9250188477ae14',
    gasService: '0xbE406F0189A0B4cf3A05C286473D23791Dd44Cc6'
}

// args
const destChain = 'Umee';
const destAddress = "umee1grechyg9el4fp36vk4typzrwyfqk4cpemmy6ya"
const receiver = ["umee1ffzd88cg4qj0jndjaqnv3py2u3vp9tv0jnupty"]
const symbol = 'aUSDC';
const amount = 2000000;

(async () => {
    const wallet = new Wallet(
        fantom.privateKey,
        new JsonRpcProvider(fantom.url),
    );

    const multiSend = new Contract(contract, MultiSend.abi, wallet);
    const usda = new Contract(tokenAddr, IERC20.abi, wallet);

    console.log(`wallet has ${(await usda.balanceOf(wallet.address)) / 1e6} ${symbol} balance`)
    console.log(`gateway is ${(await multiSend.gateway())}`)

    const approveTx = await usda.approve(multiSend.address, amount);
    try {
        const approved =  await approveTx.wait();
        console.log("Tx is approved on ",`https://testnet.ftmscan.com/tx/${approved.transactionHash}`)
    } catch (e) {
        console.log("err at approve ", e)
    }

    try {
        const sendTx = await multiSend.multiSend(destChain, destAddress, receiver, symbol, amount, {
            value: ethers.utils.parseEther('0.01'),
        });
        const tx = await sendTx.wait();
        console.log(`transaction hash is ${tx.transactionHash}`);
        console.log("Tx is approved on ",`https://testnet.axelarscan.io/gmp/${tx.transactionHash}`)

    } catch (e) {
        console.log("Error sent tx ", e)
    }
})();
```

### Execute the code

```bash
$ node interact-umee.js 
wallet has 488.85 aUSDC balance
gateway is 0x97837985Ec0494E7b9C71f5D3f9250188477ae14
Tx is approved on  https://testnet.ftmscan.com/tx/0x5f13fd087c5aebbeed073f4ee6e16faffbb12e289d7f0e0289c99fffc7299e69
transaction hash is 0x811795d093273036bd7c7cd984b0805b17f80310b0c52fb2d5c2b58a62dfbfb1
Tx is approved on  https://testnet.axelarscan.io/gmp/0x811795d093273036bd7c7cd984b0805b17f80310b0c52fb2d5c2b58a62dfbfb1
```

> You will get 2 USDC from fantom to umee receiver address

### Tested on `canon-4` network

Received  by [umee1grechyg9el4fp36vk4typzrwyfqk4cpemmy6ya](https://canon-4.api.network.umee.cc/cosmos/bank/v1beta1/balances/umee1grechyg9el4fp36vk4typzrwyfqk4cpemmy6ya)

IBC Denom trace : [6F34E1BD664C36CE49ACC28E60D62559A5F96C4F9A6CCE4FC5A67B2852E24CFE](https://canon-4.api.network.umee.cc/ibc/apps/transfer/v1/denom_traces/6F34E1BD664C36CE49ACC28E60D62559A5F96C4F9A6CCE4FC5A67B2852E24CFE)

Axelar GMP Txn: [0x332ee321c3d18435e440f0b814ecc153c6904922bc8af957ff93c13dc677ecce](https://testnet.axelarscan.io/gmp/0x332ee321c3d18435e440f0b814ecc153c6904922bc8af957ff93c13dc677ecce)

FTM Txn : [0x56116dfb03df5a1dc46e767380c26dd32db68469e64b05d46be8cd6a2d03dcb1](https://testnet.ftmscan.com/tx/0x56116dfb03df5a1dc46e767380c26dd32db68469e64b05d46be8cd6a2d03dcb1)
