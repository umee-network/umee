package tests

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"

	"github.com/umee-network/umee/x/peggy/client/cli"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func NewIntegrationTestSuite(cfg network.Config) *IntegrationTestSuite {
	return &IntegrationTestSuite{cfg: cfg}
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	s.network = network.New(s.T(), s.cfg)
	kb := s.network.Validators[0].ClientCtx.Keyring

	_, _, err := kb.NewMnemonic(
		"newAccount",
		keyring.English,
		sdk.FullFundraiserPath,
		keyring.DefaultBIP39Passphrase,
		hd.Secp256k1,
	)
	s.Require().NoError(err)

	account1, _, err := kb.NewMnemonic(
		"newAccount1",
		keyring.English,
		sdk.FullFundraiserPath,
		keyring.DefaultBIP39Passphrase,
		hd.Secp256k1,
	)
	s.Require().NoError(err)

	account2, _, err := kb.NewMnemonic(
		"newAccount2",
		keyring.English,
		sdk.FullFundraiserPath,
		keyring.DefaultBIP39Passphrase,
		hd.Secp256k1,
	)
	s.Require().NoError(err)

	multi := kmultisig.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{account1.GetPubKey(), account2.GetPubKey()})
	_, err = kb.SaveMultisig("multi", multi)
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestSetOrchestratorAddress() {
	val := s.network.Validators[0]

	ethPrivKey, err := ethcrypto.GenerateKey()
	s.Require().NoError(err)

	ethPubKey := ethPrivKey.Public()
	ethPubKeyECDSA, ok := ethPubKey.(*ecdsa.PublicKey)
	s.Require().True(ok)

	ethPrivKeyBz := crypto.FromECDSA(ethPrivKey)
	ethAddr := ethcrypto.PubkeyToAddress(*ethPubKeyECDSA)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			name: "missing Ethereum private key",
			args: []string{
				val.Address.String(),
				val.Address.String(),
				ethAddr.Hex(),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr: true,
			respType:  &sdk.TxResponse{},
		},
		{
			name: "invalid Ethereum signature",
			args: []string{
				val.Address.String(),
				val.Address.String(),
				ethAddr.Hex(),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				"--orch-eth-sig=foobar",
			},
			expectErr: true,
			respType:  &sdk.TxResponse{},
		},
		{
			name: "valid transaction",
			args: []string{
				val.Address.String(),
				val.Address.String(),
				ethAddr.Hex(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--eth-priv-key=%s", hexutil.Encode(ethPrivKeyBz)),
			},
			expectErr:    false,
			expectedCode: 0,
			respType:     &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			clientCtx := val.ClientCtx

			bz, err := clitestutil.ExecTestCLICmd(clientCtx, cli.CmdSetOrchestratorAddress(), tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(bz.Bytes(), tc.respType), bz.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}
