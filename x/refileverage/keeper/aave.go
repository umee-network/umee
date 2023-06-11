package keeper

import (
	"context"
	"crypto/ecdsa"

	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ardanlabs/ethereum"
	"github.com/ardanlabs/ethereum/currency"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var facilitatorAddr = common.HexToAddress("0x88cd7D0C712f912C4f88b6f89516dFF7F434fD95")
var key *ecdsa.PrivateKey
var e18 = big.NewInt(1000000000000000000)

func init() {
	var err error
	// pubkey: 0x610a34ed4f715f62faa86ba5a20a7602a63bc98a
	key, err = crypto.HexToECDSA("e93cf48f1b161895893f61a482bdad21a557255773ef084850ead61d4cb0d044")
	if err != nil {
		panic(err)
	}
}

func ToAaave(ctx context.Context, ghoAmount sdk.Int, recipient common.Address) error {
	backend, err := ethereum.CreateDialedBackend(ctx, "https://opt-goerli.g.alchemy.com/v2/euAaBF09KINxu0q4nEtfEVEtrK1BmotU")
	if err != nil {
		return err
	}
	defer backend.Close()

	clt, err := ethereum.NewClient(backend, key)
	if err != nil {
		return err
	}

	// =========================================================================

	const gasLimit = 1600000
	const gasPriceGwei = 39.576
	const valueGwei = 0.0
	tranOpts, err := clt.NewTransactOpts(ctx, gasLimit, currency.GWei2Wei(big.NewFloat(gasPriceGwei)), big.NewFloat(valueGwei))
	if err != nil {
		return err
	}

	// =========================================================================

	facilitator, err := NewReFiFacilitator(facilitatorAddr, backend)
	i := ghoAmount.BigInt()
	// i = i.Mul(i, e18)
	tx, err := facilitator.OnAxelarGmp(tranOpts, recipient, i)
	if err != nil {
		return err
	}

	fmt.Println("TX sent %s", tx)
	return nil
}
