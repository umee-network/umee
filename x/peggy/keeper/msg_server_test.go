package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/umee-network/umee/app"
	"github.com/umee-network/umee/x/peggy/keeper"
	"github.com/umee-network/umee/x/peggy/types"
)

func TestMsgServer_RequestBatch_InvalidSender(t *testing.T) {
	var (
		umeeApp = app.Setup(t, false, 0)
		ctx     = umeeApp.NewContext(
			false,
			tmproto.Header{
				Height: 1234567,
				Time:   time.Date(2020, time.April, 22, 12, 0, 0, 0, time.UTC),
			},
		)

		orcAddr1, _ = sdk.AccAddressFromBech32("umee1dkfhxs87adz9ll6jfr0jr5jet6u8tjaqx4z8rg")
	)

	msgServer := keeper.NewMsgServerImpl(umeeApp.PeggyKeeper)

	msg := &types.MsgRequestBatch{Orchestrator: orcAddr1.String()}
	_, err := msgServer.RequestBatch(sdk.WrapSDKContext(ctx), msg)
	require.Error(t, err)
}

func TestMsgServer_SetOrchestratorAddresses(t *testing.T) {
	ethPrivKey1, err := ethcrypto.GenerateKey()
	require.NoError(t, err)

	ethPrivKey2, err := ethcrypto.GenerateKey()
	require.NoError(t, err)

	var (
		umeeApp = app.Setup(t, false, 0)
		ctx     = umeeApp.NewContext(
			false,
			tmproto.Header{
				Height: 1234567,
				Time:   time.Date(2020, time.April, 22, 12, 0, 0, 0, time.UTC),
			},
		)

		orcAddr1, _ = sdk.AccAddressFromBech32("umee1dkfhxs87adz9ll6jfr0jr5jet6u8tjaqx4z8rg")
		valAddr1    = sdk.ValAddress(orcAddr1)
		ethAddr1    = ethcrypto.PubkeyToAddress(ethPrivKey1.PublicKey)

		orcAddr2, _ = sdk.AccAddressFromBech32("umee185vwyk9xu2k8h0c4yn2uauxfecktfvd44urlcc")
		valAddr2    = sdk.ValAddress(orcAddr2)
		ethAddr2    = ethcrypto.PubkeyToAddress(ethPrivKey2.PublicKey)
	)

	// setup for getSignerValidator
	umeeApp.PeggyKeeper.StakingKeeper = NewStakingKeeperMock(valAddr1)

	// Set the sequence to 1 because the antehandler will do this in the full
	// chain.
	acc := umeeApp.AccountKeeper.NewAccountWithAddress(ctx, orcAddr1)
	acc.SetSequence(1)
	umeeApp.AccountKeeper.SetAccount(ctx, acc)

	msgServer := keeper.NewMsgServerImpl(umeeApp.PeggyKeeper)

	// ensure we cannot set the validator's orchestrator addresses with an invalid sig
	signMsgBz := umeeApp.AppCodec().MustMarshal(&types.SetOrchestratorAddressesSignMsg{
		ValidatorAddress: valAddr1.String(),
		Nonce:            45,
	})
	badSig, err := types.NewEthereumSignature(crypto.Keccak256Hash(signMsgBz), ethPrivKey1)
	require.NoError(t, err)

	msg := types.NewMsgSetOrchestratorAddress(sdk.AccAddress(valAddr1), orcAddr1, ethAddr1, badSig)
	_, err = msgServer.SetOrchestratorAddresses(sdk.WrapSDKContext(ctx), msg)
	require.Error(t, err)

	// ensure we can set the validator's orchestrator addresses
	signMsgBz = umeeApp.AppCodec().MustMarshal(&types.SetOrchestratorAddressesSignMsg{
		ValidatorAddress: valAddr1.String(),
		Nonce:            0,
	})
	goodSig, err := types.NewEthereumSignature(crypto.Keccak256Hash(signMsgBz), ethPrivKey1)
	require.NoError(t, err)

	msg = types.NewMsgSetOrchestratorAddress(sdk.AccAddress(valAddr1), orcAddr1, ethAddr1, goodSig)
	_, err = msgServer.SetOrchestratorAddresses(sdk.WrapSDKContext(ctx), msg)
	require.NoError(t, err)

	// ensure we cannot set the same ethereum key for another validator
	msg = types.NewMsgSetOrchestratorAddress(sdk.AccAddress(valAddr2), orcAddr2, ethAddr1, goodSig)
	_, err = msgServer.SetOrchestratorAddresses(sdk.WrapSDKContext(ctx), msg)
	require.Error(t, err)

	// ensure we cannot set the same orchestrator key for another validator
	msg = types.NewMsgSetOrchestratorAddress(sdk.AccAddress(valAddr2), orcAddr1, ethAddr2, goodSig)
	_, err = msgServer.SetOrchestratorAddresses(sdk.WrapSDKContext(ctx), msg)
	require.Error(t, err)
}
