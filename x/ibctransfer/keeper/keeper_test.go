package keeper_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v2/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v2/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v2/testing"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/umee-network/umee/app"
)

type KeeperTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator
	chainA      *ibctesting.TestChain
	chainB      *ibctesting.TestChain

	queryClient types.QueryClient
}

func (s *KeeperTestSuite) SetupTest() {
	s.coordinator = ibctesting.NewCoordinator(s.T(), 0)
	ibctesting.DefaultTestingAppInit = SetupTestingApp

	chains := make(map[string]*ibctesting.TestChain)
	for i := 0; i < 2; i++ {
		// create a chain with the temporary coordinator that we'll later override
		chainID := ibctesting.GetChainID(i)
		chain := ibctesting.NewTestChain(s.T(), ibctesting.NewCoordinator(s.T(), 0), chainID)

		balance := banktypes.Balance{
			Address: chain.SenderAccount.GetAddress().String(),
			Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000000000000))),
		}

		// create application and override files in the IBC test chain
		app := ibctesting.SetupWithGenesisValSet(
			s.T(),
			chain.Vals,
			[]authtypes.GenesisAccount{
				chain.SenderAccount.(authtypes.GenesisAccount),
			},
			balance,
		)

		chain.App = app
		chain.QueryServer = app.GetIBCKeeper()
		chain.TxConfig = app.GetTxConfig()
		chain.Codec = app.AppCodec()
		chain.CurrentHeader = tmproto.Header{
			ChainID: chainID,
			Height:  1,
			Time:    s.coordinator.CurrentTime.UTC(),
		}

		chain.Coordinator = s.coordinator
		s.coordinator.CommitBlock(chain)

		chains[chainID] = chain
	}

	s.coordinator.Chains = chains
	s.chainA = s.coordinator.GetChain(ibctesting.GetChainID(0))
	s.chainB = s.coordinator.GetChain(ibctesting.GetChainID(1))

	umeeApp := s.GetUmeeApp(s.chainA)

	queryHelper := baseapp.NewQueryServerTestHelper(s.chainA.GetContext(), umeeApp.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, umeeApp.TransferKeeper)
	s.queryClient = types.NewQueryClient(queryHelper)
}

func (s *KeeperTestSuite) GetUmeeApp(c *ibctesting.TestChain) *app.UmeeApp {
	umeeApp, ok := c.App.(*app.UmeeApp)
	s.Require().True(ok)

	return umeeApp
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) TestGetTransferAccount() {
	expectedModAccAddr := sdk.AccAddress(crypto.AddressHash([]byte(types.ModuleName)))
	macc := s.GetUmeeApp(s.chainA).TransferKeeper.GetTransferAccount(s.chainA.GetContext())

	s.Require().NotNil(macc)
	s.Require().Equal(types.ModuleName, macc.GetName())
	s.Require().Equal(expectedModAccAddr, macc.GetAddress())
}

func (s *KeeperTestSuite) TestTrackMetadata() {
	pathAtoB := NewTransferPath(s.chainA, s.chainB)
	s.coordinator.Setup(pathAtoB)

	s.Run("OnRecvPacketA", func() {
		denom := strings.Join([]string{
			s.chainB.ChainID,
			s.chainA.ChainID,
			"quark",
		}, "/")

		data := ibctransfertypes.NewFungibleTokenPacketData(
			denom,
			strconv.Itoa(1),
			AddressFromString("a3"),
			AddressFromString("a4"),
		)

		packet := channeltypes.NewPacket(
			data.GetBytes(),
			uint64(1),
			"transfer",
			"channel-0",
			"transfer",
			"channel-0",
			clienttypes.NewHeight(0, 100),
			0,
		)

		err := s.GetUmeeApp(s.chainA).TransferKeeper.OnRecvPacket(s.chainA.GetContext(), packet, data)
		s.Require().NoError(err)
	})

	s.Run("OnRecvPacketB", func() {
		denom := strings.Join([]string{
			s.chainA.ChainID,
			s.chainB.ChainID,
			"photon",
		}, "/")

		data := ibctransfertypes.NewFungibleTokenPacketData(
			denom,
			strconv.Itoa(1),
			AddressFromString("a2"),
			AddressFromString("a1"),
		)

		packet := channeltypes.NewPacket(
			data.GetBytes(),
			uint64(1),
			"transfer",
			"channel-0",
			"transfer",
			"channel-0",
			clienttypes.NewHeight(0, 100),
			0,
		)

		err := s.GetUmeeApp(s.chainB).TransferKeeper.OnRecvPacket(s.chainB.GetContext(), packet, data)
		s.Require().NoError(err)
	})

	s.Run("SendTransfer", func() {
		denom := strings.Join([]string{
			pathAtoB.EndpointA.ChannelConfig.PortID,
			pathAtoB.EndpointA.ChannelID,
			s.chainA.ChainID,
			s.chainB.ChainID,
			"photon",
		}, "/")

		data := ibctransfertypes.NewFungibleTokenPacketData(
			denom,
			strconv.Itoa(1),
			AddressFromString("a1"),
			AddressFromString("a2"),
		)

		packet := channeltypes.NewPacket(
			data.GetBytes(),
			uint64(1),
			"transfer",
			"channel-0",
			"transfer",
			"channel-0",
			clienttypes.NewHeight(0, 100),
			0,
		)

		sender, err := sdk.AccAddressFromBech32(data.Sender)
		s.Require().NoError(err)

		denomTrace := types.ParseDenomTrace(data.Denom)
		ibcDenom := denomTrace.IBCDenom()

		registerDenom := func() {
			denomTrace := types.ParseDenomTrace(denom)
			traceHash := denomTrace.Hash()
			if !s.GetUmeeApp(s.chainB).TransferKeeper.HasDenomTrace(s.chainB.GetContext(), traceHash) {
				s.GetUmeeApp(s.chainB).TransferKeeper.SetDenomTrace(s.chainB.GetContext(), denomTrace)
			}
		}

		registerDenom()

		amount, err := strconv.Atoi(data.Amount)
		s.Require().NoError(err)

		err = s.GetUmeeApp(s.chainB).TransferKeeper.SendTransfer(
			s.chainB.GetContext(),
			packet.SourcePort,
			packet.SourceChannel,
			sdk.NewCoin(ibcDenom, sdk.NewInt(int64(amount))),
			sender,
			data.Receiver,
			clienttypes.NewHeight(0, 110),
			0,
		)
		s.Require().NoError(err)
	})

	s.coordinator.CommitBlock(s.chainA, s.chainB)

	_, ok := s.GetUmeeApp(s.chainA).BankKeeper.GetDenomMetaData(s.chainA.GetContext(), "ibc/DB6D78EC2E51C8B6AAF6DA64E660911491DC1A67C64DA69ED6945FE6DB552A5C")
	s.Require().True(ok)

	_, ok = s.GetUmeeApp(s.chainB).BankKeeper.GetDenomMetaData(s.chainB.GetContext(), "ibc/10180B5BF0701A3E34A5F818607D7E57ECD35CD9D673ABCCD174F157DFC06C0F")
	s.Require().True(ok)
}
