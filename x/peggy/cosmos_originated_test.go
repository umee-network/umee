package peggy_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v040auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v040"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/umee-network/umee/x/peggy"
	"github.com/umee-network/umee/x/peggy/keeper"
	"github.com/umee-network/umee/x/peggy/testpeggy"
	"github.com/umee-network/umee/x/peggy/types"
)

// Have the validators put in a erc20<>denom relation with ERC20DeployedEvent
// Send some coins of that denom into the cosmos module
// Check that the coins are locked, not burned
// Have the validators put in a deposit event for that ERC20
// Check that the coins are unlocked and sent to the right account

func TestCosmosOriginated(t *testing.T) {
	tv := initializeTestingVars(t)
	addDenomToERC20Relation(tv)
	lockCoinsInModule(tv)
	acceptDepositEvent(tv)
}

type testingVars struct {
	myOrchestratorAddr sdk.AccAddress
	myValAddr          sdk.ValAddress
	erc20              common.Address
	denom              string
	input              testpeggy.TestInput
	ctx                sdk.Context
	ms                 types.MsgServer
	t                  *testing.T
}

func initializeTestingVars(t *testing.T) *testingVars {
	var tv testingVars

	tv.t = t

	tv.myOrchestratorAddr = make([]byte, v040auth.AddrLen)
	tv.myValAddr = sdk.ValAddress(tv.myOrchestratorAddr) // revisit when proper mapping is impl in keeper

	tv.erc20 = common.HexToAddress("0x0bc529c00c6401aef6d220be8c6ea1667f6ad93e")
	tv.denom = "uatom"

	tv.input = testpeggy.CreateTestEnv(t)
	tv.ctx = tv.input.Context
	tv.input.PeggyKeeper.StakingKeeper = testpeggy.NewStakingKeeperMock(tv.myValAddr)
	tv.input.PeggyKeeper.SetOrchestratorValidator(tv.ctx, tv.myValAddr, tv.myOrchestratorAddr)
	tv.ms = keeper.NewMsgServerImpl(tv.input.PeggyKeeper)

	return &tv
}

func addDenomToERC20Relation(tv *testingVars) {
	tv.input.BankKeeper.SetDenomMetaData(tv.ctx, bank.Metadata{
		Description: "The native staking token of the Cosmos Hub.",
		DenomUnits: []*bank.DenomUnit{
			{Denom: "uatom", Exponent: uint32(0), Aliases: []string{"microatom"}},
			{Denom: "matom", Exponent: uint32(3), Aliases: []string{"milliatom"}},
			{Denom: "atom", Exponent: uint32(6), Aliases: []string{}},
		},
		Base:    "uatom",
		Display: "atom",
	})

	var (
		myNonce = uint64(1)
	)

	ethClaim := types.MsgERC20DeployedClaim{
		CosmosDenom:   tv.denom,
		TokenContract: tv.erc20.Hex(),
		Name:          "atom",
		Symbol:        "atom",
		Decimals:      6,
		EventNonce:    myNonce,
		Orchestrator:  tv.myOrchestratorAddr.String(),
	}

	_, err := tv.ms.ERC20DeployedClaim(sdk.WrapSDKContext(tv.ctx), &ethClaim)
	require.NoError(tv.t, err)

	NewBlockHandler(tv.input.PeggyKeeper).EndBlocker(tv.ctx)

	// check if attestation persisted
	attestation := tv.input.PeggyKeeper.GetAttestation(tv.ctx, myNonce, ethClaim.ClaimHash())
	require.NotNil(tv.t, attestation)

	// check if erc20<>denom relation added to db
	isCosmosOriginated, gotERC20, err := tv.input.PeggyKeeper.DenomToERC20Lookup(tv.ctx, tv.denom)

	require.NoError(tv.t, err)
	assert.True(tv.t, isCosmosOriginated)

	isCosmosOriginated, gotDenom := tv.input.PeggyKeeper.ERC20ToDenomLookup(tv.ctx, tv.erc20)
	assert.True(tv.t, isCosmosOriginated)

	assert.Equal(tv.t, tv.denom, gotDenom)
	assert.Equal(tv.t, tv.erc20, gotERC20)
}

func lockCoinsInModule(tv *testingVars) {
	var (
		userCosmosAddr, _  = sdk.AccAddressFromBech32("cosmos1990z7dqsvh8gthw9pa5sn4wuy2xrsd80mg5z6y")
		denom              = "uatom"
		startingCoinAmount = sdk.NewIntFromUint64(150)
		sendAmount         = sdk.NewIntFromUint64(50)
		feeAmount          = sdk.NewIntFromUint64(5)
		startingCoins      = sdk.Coins{sdk.NewCoin(denom, startingCoinAmount)}
		sendingCoin        = sdk.NewCoin(denom, sendAmount)
		feeCoin            = sdk.NewCoin(denom, feeAmount)
		ethDestination     = common.HexToAddress("0x3c9289da00b02dC623d0D8D907619890301D26d4")
	)

	// we start by depositing some funds into the users balance to send
	tv.input.BankKeeper.MintCoins(tv.ctx, types.ModuleName, startingCoins)
	tv.input.BankKeeper.SendCoinsFromModuleToAccount(tv.ctx, types.ModuleName, userCosmosAddr, startingCoins)
	balance1 := tv.input.BankKeeper.GetAllBalances(tv.ctx, userCosmosAddr)
	assert.Equal(tv.t, sdk.Coins{sdk.NewCoin(denom, startingCoinAmount)}, balance1)

	// send some coins
	msg := &types.MsgSendToEth{
		Sender:    userCosmosAddr.String(),
		EthDest:   ethDestination.Hex(),
		Amount:    sendingCoin,
		BridgeFee: feeCoin,
	}

	_, err := tv.ms.SendToEth(sdk.WrapSDKContext(tv.ctx), msg)
	require.NoError(tv.t, err)

	// Check that user balance has gone down
	balance2 := tv.input.BankKeeper.GetAllBalances(tv.ctx, userCosmosAddr)
	assert.Equal(tv.t, sdk.Coins{sdk.NewCoin(denom, startingCoinAmount.Sub(sendAmount).Sub(feeAmount))}, balance2)

	// Check that peggy balance has gone up
	peggyAddr := tv.input.AccountKeeper.GetModuleAddress(types.ModuleName)
	assert.Equal(tv.t,
		sdk.Coins{sdk.NewCoin(denom, sendAmount.Add(feeAmount))},
		tv.input.BankKeeper.GetAllBalances(tv.ctx, peggyAddr),
	)
}

func acceptDepositEvent(tv *testingVars) {
	var (
		myOrchestratorAddr sdk.AccAddress = make([]byte, v040auth.AddrLen)
		myCosmosAddr, _                   = sdk.AccAddressFromBech32("cosmos16ahjkfqxpp6lvfy9fpfnfjg39xr96qett0alj5")
		myNonce                           = uint64(2)
		anyETHAddr                        = common.HexToAddress("0xf9613b532673cc223aba451dfa8539b87e1f666d")
	)

	myErc20 := types.ERC20Token{
		Amount:   sdk.NewInt(12),
		Contract: tv.erc20.Hex(),
	}

	ethClaim := types.MsgDepositClaim{
		EventNonce:     myNonce,
		TokenContract:  myErc20.Contract,
		Amount:         myErc20.Amount,
		EthereumSender: anyETHAddr.Hex(),
		CosmosReceiver: myCosmosAddr.String(),
		Orchestrator:   myOrchestratorAddr.String(),
	}

	_, err := tv.ms.DepositClaim(sdk.WrapSDKContext(tv.ctx), &ethClaim)
	require.NoError(tv.t, err)
	NewBlockHandler(tv.input.PeggyKeeper).EndBlocker(tv.ctx)

	// check that attestation persisted
	a := tv.input.PeggyKeeper.GetAttestation(tv.ctx, myNonce, ethClaim.ClaimHash())
	require.NotNil(tv.t, a)

	// Check that user balance has gone up
	assert.Equal(tv.t,
		sdk.Coins{sdk.NewCoin(tv.denom, myErc20.Amount)},
		tv.input.BankKeeper.GetAllBalances(tv.ctx, myCosmosAddr))

	// Check that peggy balance has gone down
	peggyAddr := tv.input.AccountKeeper.GetModuleAddress(types.ModuleName)
	assert.Equal(tv.t,
		sdk.Coins{sdk.NewCoin(tv.denom, sdk.NewIntFromUint64(55).Sub(myErc20.Amount))},
		tv.input.BankKeeper.GetAllBalances(tv.ctx, peggyAddr),
	)
}
