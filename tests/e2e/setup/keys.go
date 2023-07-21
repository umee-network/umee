package setup

import (
	"github.com/cosmos/go-bip39"
	appparams "github.com/umee-network/umee/v5/app/params"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	PhotonDenom    = "photon"
	InitBalanceStr = "510000000000" + appparams.BondDenom + ",100000000000" + PhotonDenom
	GaiaChainID    = "test-gaia-chain"

	EthChainID uint = 15
	EthMinerPK      = "0xb1bab011e03a9862664706fc3bbaa1b16651528e5f0e7fbfcbfdd8be302a13e7"

	PriceFeederContainerRepo  = "ghcr.io/umee-network/price-feeder-umee"
	PriceFeederServerPort     = "7171/tcp"
	PriceFeederMaxStartupTime = 20 // seconds
)

var (
	minGasPrice     = appparams.ProtocolMinGasPrice.String()
	stakeAmount, _  = sdk.NewIntFromString("100000000000")
	stakeAmountCoin = sdk.NewCoin(appparams.BondDenom, stakeAmount)

	stakeAmount2, _  = sdk.NewIntFromString("500000000000")
	stakeAmountCoin2 = sdk.NewCoin(appparams.BondDenom, stakeAmount2)
)

var (
	ATOM          = "ATOM"
	ATOMBaseDenom = "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
	ATOMExponent  = 6
)

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

func createMemoryKey(cdc codec.Codec) (mnemonic string, info *keyring.Record, err error) {
	mnemonic, err = createMnemonic()
	if err != nil {
		return "", nil, err
	}

	account, err := createMemoryKeyFromMnemonic(cdc, mnemonic)
	if err != nil {
		return "", nil, err
	}

	return mnemonic, account, nil
}

func createMemoryKeyFromMnemonic(cdc codec.Codec, mnemonic string) (*keyring.Record, error) {
	kb, err := keyring.New("testnet", keyring.BackendMemory, "", nil, cdc)
	if err != nil {
		return nil, err
	}

	keyringAlgos, _ := kb.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
	if err != nil {
		return nil, err
	}

	account, err := kb.NewAccount("", mnemonic, "", sdk.FullFundraiserPath, algo)
	if err != nil {
		return nil, err
	}

	return account, nil
}
