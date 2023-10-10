package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	tmcfg "github.com/cometbft/cometbft/config"
	tmos "github.com/cometbft/cometbft/libs/os"
	p2p "github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkcrypto "github.com/cosmos/cosmos-sdk/crypto"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	umeeapp "github.com/umee-network/umee/v6/app"
	appparams "github.com/umee-network/umee/v6/app/params"
)

type validator struct {
	chain        *chain
	index        int
	moniker      string
	mnemonic     string
	KeyInfo      keyring.Record
	privateKey   cryptotypes.PrivKey
	consensusKey privval.FilePVKey
	nodeKey      p2p.NodeKey
}

// instanceName is <v.moniker><v.index>
func (v *validator) instanceName() string {
	return fmt.Sprintf("%s%d", v.moniker, v.index)
}

// configDir is <chain.configDir()>/<val.instanceName>
func (v *validator) configDir() string {
	return fmt.Sprintf("%s/%s", v.chain.configDir(), v.instanceName())
}

// createConfigDir makes the validator's config directory at <chain.configDir()>/<val.instanceName>/config
func (v *validator) createConfigDir() error {
	p := path.Join(v.configDir(), "config")
	return os.MkdirAll(p, 0o755)
}

// init creates the validator's config directory and creates its genesis.json and config.toml
func (v *validator) init(cdc codec.Codec) error {
	if err := v.createConfigDir(); err != nil {
		return err
	}

	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(v.configDir())
	config.Moniker = v.moniker

	// TODO: We just created the directory. Isn't this just returning empty genesis doc? If so, replace this code.
	genDoc, err := getGenDoc(v.configDir())
	if err != nil {
		return err
	}

	appState, err := json.MarshalIndent(umeeapp.ModuleBasics.DefaultGenesis(cdc), "", " ")
	if err != nil {
		return fmt.Errorf("failed to JSON encode app genesis state: %w", err)
	}

	genDoc.ChainID = v.chain.ID
	genDoc.Validators = nil
	genDoc.AppState = appState

	if err = genutil.ExportGenesisFile(genDoc, config.GenesisFile()); err != nil {
		return fmt.Errorf("failed to export app genesis state: %w", err)
	}

	tmcfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)
	return nil
}

// loads or generates a node key in the validator's config directory
func (v *validator) createNodeKey() error {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(v.configDir())
	config.Moniker = v.moniker

	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return err
	}

	v.nodeKey = *nodeKey
	return nil
}

// loads or generates a consensus key in the validator's config directory
func (v *validator) createConsensusKey() error {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(v.configDir())
	config.Moniker = v.moniker

	pvKeyFile := config.PrivValidatorKeyFile()
	if err := tmos.EnsureDir(filepath.Dir(pvKeyFile), 0o777); err != nil {
		return err
	}

	pvStateFile := config.PrivValidatorStateFile()
	if err := tmos.EnsureDir(filepath.Dir(pvStateFile), 0o777); err != nil {
		return err
	}

	filePV := privval.LoadOrGenFilePV(pvKeyFile, pvStateFile)
	v.consensusKey = filePV.Key

	return nil
}

// createKey generates a random key and saves to keyring-backend-test in the validator's config directory
// Uses test keyring passphrase so that the price feeder can vote using this key.
func (v *validator) createKey(cdc codec.Codec, name string) error {
	mnemonic, err := createMnemonic()
	if err != nil {
		return err
	}

	kb, err := keyring.New(keyringAppName, keyring.BackendTest, v.configDir(), nil, cdc)
	if err != nil {
		return err
	}

	keyringAlgos, _ := kb.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
	if err != nil {
		return err
	}

	info, err := kb.NewAccount(name, mnemonic, "", sdk.FullFundraiserPath, algo)
	if err != nil {
		return err
	}

	privKeyArmor, err := kb.ExportPrivKeyArmor(name, keyringPassphrase)
	if err != nil {
		return err
	}

	privKey, _, err := sdkcrypto.UnarmorDecryptPrivKey(privKeyArmor, keyringPassphrase)
	if err != nil {
		return err
	}

	v.KeyInfo = *info
	v.mnemonic = mnemonic
	v.privateKey = privKey

	return nil
}

func (v *validator) buildCreateValidatorMsg(amount sdk.Coin) (sdk.Msg, error) {
	description := stakingtypes.NewDescription(v.moniker, "", "", "", "")
	commissionRates := stakingtypes.CommissionRates{
		Rate:          sdk.MustNewDecFromStr("0.1"),
		MaxRate:       sdk.MustNewDecFromStr("0.2"),
		MaxChangeRate: sdk.MustNewDecFromStr("0.01"),
	}

	// get the initial validator min self delegation
	minSelfDelegation, _ := sdk.NewIntFromString("1")

	valPubKey, err := cryptocodec.FromTmPubKeyInterface(v.consensusKey.PubKey)
	if err != nil {
		return nil, err
	}
	valAddr, err := v.KeyInfo.GetAddress()
	if err != nil {
		return nil, err
	}

	return stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(valAddr),
		valPubKey,
		amount,
		description,
		commissionRates,
		minSelfDelegation,
	)
}

func (v *validator) signMsg(cdc codec.Codec, msgs ...sdk.Msg) (*sdktx.Tx, error) {
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return nil, err
	}

	txBuilder.SetMemo(fmt.Sprintf("%s@%s:26656", v.nodeKey.ID(), v.instanceName()))
	txBuilder.SetFeeAmount(sdk.NewCoins())
	txBuilder.SetGasLimit(appparams.DefaultGasLimit)

	signerData := authsigning.SignerData{
		ChainID:       v.chain.ID,
		AccountNumber: 0,
		Sequence:      0,
	}

	// For SIGN_MODE_DIRECT, calling SetSignatures calls setSignerInfos on
	// TxBuilder under the hood, and SignerInfos is needed to generate the sign
	// bytes. This is the reason for setting SetSignatures here, with a nil
	// signature.
	//
	// Note: This line is not needed for SIGN_MODE_LEGACY_AMINO, but putting it
	// also doesn't affect its generated sign bytes, so for code's simplicity
	// sake, we put it here.
	pubKey, err := v.KeyInfo.GetPubKey()
	if err != nil {
		return nil, err
	}
	sig := txsigning.SignatureV2{
		PubKey: pubKey,
		Data: &txsigning.SingleSignatureData{
			SignMode:  txsigning.SignMode_SIGN_MODE_DIRECT,
			Signature: nil,
		},
		Sequence: 0,
	}

	if err := txBuilder.SetSignatures(sig); err != nil {
		return nil, err
	}

	bytesToSign, err := encodingConfig.TxConfig.SignModeHandler().GetSignBytes(
		txsigning.SignMode_SIGN_MODE_DIRECT,
		signerData,
		txBuilder.GetTx(),
	)
	if err != nil {
		return nil, err
	}

	sigBytes, err := v.privateKey.Sign(bytesToSign)
	if err != nil {
		return nil, err
	}

	sig = txsigning.SignatureV2{
		PubKey: pubKey,
		Data: &txsigning.SingleSignatureData{
			SignMode:  txsigning.SignMode_SIGN_MODE_DIRECT,
			Signature: sigBytes,
		},
		Sequence: 0,
	}
	if err := txBuilder.SetSignatures(sig); err != nil {
		return nil, err
	}

	signedTx := txBuilder.GetTx()
	bz, err := encodingConfig.TxConfig.TxEncoder()(signedTx)
	if err != nil {
		return nil, err
	}

	return decodeTx(cdc, bz)
}
