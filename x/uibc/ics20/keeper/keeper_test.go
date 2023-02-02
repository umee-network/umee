package keeper_test

import (
	"strconv"
	"strings"
	"testing"

	gravitytypes "github.com/Gravity-Bridge/Gravity-Bridge/module/x/gravity/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"gotest.tools/v3/assert"

	umeeapp "github.com/umee-network/umee/v4/app"
	"github.com/umee-network/umee/v4/tests/util"
)

type IntegrationSuite struct {
	coordinator *ibctesting.Coordinator
	chainA      *ibctesting.TestChain
	chainB      *ibctesting.TestChain

	queryClient ibctransfertypes.QueryClient
}

func initIntegrationSuite(t *testing.T) *IntegrationSuite {
	s := &IntegrationSuite{}
	s.coordinator = ibctesting.NewCoordinator(t, 0)

	chains := make(map[string]*ibctesting.TestChain)
	for i := 0; i < 2; i++ {
		ibctesting.DefaultTestingAppInit = SetupTestingApp

		// create a chain with the temporary coordinator that we'll later override
		chainID := ibctesting.GetChainID(i)
		chain := ibctesting.NewTestChain(t, ibctesting.NewCoordinator(t, 0), chainID)

		balance := banktypes.Balance{
			Address: chain.SenderAccount.GetAddress().String(),
			Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000000000000))),
		}

		// create application and override files in the IBC test chain
		app := ibctesting.SetupWithGenesisValSet(
			t,
			chain.Vals,
			[]authtypes.GenesisAccount{
				chain.SenderAccount.(authtypes.GenesisAccount),
			},
			chainID,
			sdk.DefaultPowerReduction,
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

		// set gravity bridge delegate keys
		umeApp := app.(*umeeapp.UmeeApp)
		for _, val := range chain.Vals.Validators {
			_, _, ethAddr, err := util.GenerateRandomEthKey()
			assert.NilError(t, err)

			gravityEthAddr, err := gravitytypes.NewEthAddress(ethAddr.Hex())
			assert.NilError(t, err)

			umeApp.GravityKeeper.SetOrchestratorValidator(
				chain.GetContext(),
				sdk.ValAddress(val.Address),
				sdk.AccAddress(val.Address),
			)
			umeApp.GravityKeeper.SetEthAddressForValidator(
				chain.GetContext(),
				sdk.ValAddress(val.Address),
				*gravityEthAddr,
			)
		}

		chain.Coordinator = s.coordinator
		s.coordinator.CommitBlock(chain)

		chains[chainID] = chain
	}

	s.coordinator.Chains = chains
	s.chainA = s.coordinator.GetChain(ibctesting.GetChainID(0))
	s.chainB = s.coordinator.GetChain(ibctesting.GetChainID(1))

	umeeApp := s.GetUmeeApp(s.chainA, t)

	queryHelper := baseapp.NewQueryServerTestHelper(s.chainA.GetContext(), umeeApp.InterfaceRegistry())
	ibctransfertypes.RegisterQueryServer(queryHelper, umeeApp.UIBCTransferKeeper)
	s.queryClient = ibctransfertypes.NewQueryClient(queryHelper)

	return s
}

func (k *IntegrationSuite) GetUmeeApp(c *ibctesting.TestChain, t *testing.T) *umeeapp.UmeeApp {
	umeeApp, ok := c.App.(*umeeapp.UmeeApp)
	assert.Equal(t, true, ok)

	return umeeApp
}

func TestTrackMetadata(t *testing.T) {
	t.Skip("ibctransfer integration tests require further investigation, currently it breaks on connection handshake")
	s := initIntegrationSuite(t)
	pathAtoB := NewTransferPath(s.chainA, s.chainB)
	s.coordinator.Setup(pathAtoB)

	t.Run("OnRecvPacketA", func(t *testing.T) {
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

		err := s.GetUmeeApp(s.chainA, t).UIBCTransferKeeper.OnRecvPacket(s.chainA.GetContext(), packet, data)
		assert.NilError(t, err)
	})

	t.Run("OnRecvPacketB", func(t *testing.T) {
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

		err := s.GetUmeeApp(s.chainB, t).UIBCTransferKeeper.OnRecvPacket(s.chainB.GetContext(), packet, data)
		assert.NilError(t, err)
	})

	t.Run("SendTransfer", func(t *testing.T) {
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
		assert.NilError(t, err)

		denomTrace := ibctransfertypes.ParseDenomTrace(data.Denom)
		ibcDenom := denomTrace.IBCDenom()

		registeredenom := func() {
			denomTrace := ibctransfertypes.ParseDenomTrace(denom)
			traceHash := denomTrace.Hash()
			if !s.GetUmeeApp(s.chainB, t).UIBCTransferKeeper.HasDenomTrace(s.chainB.GetContext(), traceHash) {
				s.GetUmeeApp(s.chainB, t).UIBCTransferKeeper.SetDenomTrace(s.chainB.GetContext(), denomTrace)
			}
		}

		registeredenom()

		amount, err := strconv.Atoi(data.Amount)
		assert.NilError(t, err)

		err = s.GetUmeeApp(s.chainB, t).UIBCTransferKeeper.SendTransfer(
			s.chainB.GetContext(),
			packet.SourcePort,
			packet.SourceChannel,
			sdk.NewCoin(ibcDenom, sdk.NewInt(int64(amount))),
			sender,
			data.Receiver,
			clienttypes.NewHeight(0, 110),
			0,
		)
		assert.NilError(t, err)
	})

	s.coordinator.CommitBlock(s.chainA, s.chainB)

	_, ok := s.GetUmeeApp(s.chainA, t).BankKeeper.GetDenomMetaData(s.chainA.GetContext(), "ibc/DB6D78EC2E51C8B6AAF6DA64E660911491DC1A67C64DA69ED6945FE6DB552A5C")
	assert.Equal(t, true, ok)

	_, ok = s.GetUmeeApp(s.chainB, t).BankKeeper.GetDenomMetaData(s.chainB.GetContext(), "ibc/10180B5BF0701A3E34A5F818607D7E57ECD35CD9D673ABCCD174F157DFC06C0F")
	assert.Equal(t, true, ok)
}
