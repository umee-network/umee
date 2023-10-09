package setup

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"

	appparams "github.com/umee-network/umee/v6/app/params"
	"github.com/umee-network/umee/v6/x/metoken/mocks"
)

const (
	ATOM          = "ATOM"
	ATOMBaseDenom = "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
	ATOMExponent  = 6
	PhotonDenom   = "photon"

	// TODO: validator initial balances can be only uumee. Non-validaotr accounts get other tokens for testing.
	InitBalanceStr = "510000000000" + appparams.BondDenom + ",100000000000" + PhotonDenom + ",100000000000" + mocks.USDTBaseDenom

	GaiaChainID = "test-gaia-chain"

	PriceFeederContainerRepo  = "ghcr.io/umee-network/price-feeder-umee"
	PriceFeederServerPort     = "7171/tcp"
	PriceFeederMaxStartupTime = 20 // seconds
)

var (
	minGasPrice = appparams.ProtocolMinGasPrice.String()

	// TODO: stake less on the validators, and instead delegate from a non-validator account
	stakeAmount, _  = sdk.NewIntFromString("100000000000")
	stakeAmountCoin = sdk.NewCoin(appparams.BondDenom, stakeAmount)

	stakeAmount2, _  = sdk.NewIntFromString("500000000000")
	stakeAmountCoin2 = sdk.NewCoin(appparams.BondDenom, stakeAmount2)
)

// createMnemonic generates a random mnemonic to be used in key generation
func createMnemonic() (string, error) {
	entropySeed, err := bip39.NewEntropy(256)
	if err != nil {
		return "", err
	}

	mnemonic, err := bip39.NewMnemonic(entropySeed)
	if err != nil {
		return "", err
	}

	return mnemonic, nil
}

// createMemoryKey generates a random key which will be stored only in memory
func createMemoryKey(cdc codec.Codec) (mnemonic string, info *keyring.Record, err error) {
	mnemonic, err = createMnemonic()
	if err != nil {
		return "", nil, err
	}

	kb, err := keyring.New("testnet", keyring.BackendMemory, "", nil, cdc)
	if err != nil {
		return "", nil, err
	}

	keyringAlgos, _ := kb.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
	if err != nil {
		return "", nil, err
	}

	account, err := kb.NewAccount("", mnemonic, "", sdk.FullFundraiserPath, algo)
	if err != nil {
		return "", nil, err
	}

	return mnemonic, account, nil
}
