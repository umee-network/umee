package keeper

import (
	"context"

	"fmt"
	"log"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ardanlabs/ethereum"
	"github.com/ardanlabs/ethereum/currency"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func ToAaave(ctx1 context.Context, ghoAmount sdk.Int, recipient common.Address) error {
	ctx := context.Background()

	backend, err := ethereum.CreateDialedBackend(ctx, "https://opt-goerli.g.alchemy.com/v2/euAaBF09KINxu0q4nEtfEVEtrK1BmotU")
	if err != nil {
		panic(err)
	}
	defer backend.Close()

	privateKey, err := crypto.HexToECDSA("e93cf48f1b161895893f61a482bdad21a557255773ef084850ead61d4cb0d044")
	if err != nil {
		log.Fatal(err)
	}

	clt, err := ethereum.NewClient(backend, privateKey)
	if err != nil {
		panic(err)
	}

	// =========================================================================

	const gasLimit = 1600000
	const gasPriceGwei = 39.576
	const valueGwei = 0.0
	tranOpts, err := clt.NewTransactOpts(ctx, gasLimit, currency.GWei2Wei(big.NewFloat(gasPriceGwei)), big.NewFloat(valueGwei))
	if err != nil {
		panic(err)
	}

	// =========================================================================

	address := common.HexToAddress("0x5E9464a09F301af854c546c3aEE3762f7146d683")
	//recipient := common.HexToAddress("0x610A34ed4F715F62faa86BA5A20a7602A63bc98a")
	facilitator, err := NewReFiFacilitator(address, backend)
	tx, err := facilitator.OnAxelarGmp(tranOpts, recipient, big.NewInt(100))

	if err != nil {
		panic(err)
	}

	fmt.Println("TX sent %s", tx)
	return nil
}
