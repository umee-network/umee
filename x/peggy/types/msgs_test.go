package types_test

import (
	"bytes"
	"crypto/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v040auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v040"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/x/peggy/types"
)

func TestValidateMsgSetOrchestratorAddresses(t *testing.T) {
	var (
		ethAddress                   = common.HexToAddress("0xb462864E395d88d6bc7C5dd5F3F5eb4cc2599255")
		cosmosAddress sdk.AccAddress = bytes.Repeat([]byte{0x1}, v040auth.AddrLen)
		valAddress    sdk.AccAddress = bytes.Repeat([]byte{0x1}, v040auth.AddrLen)
	)

	ethSig := make([]byte, 65)
	rand.Read(ethSig)

	specs := map[string]struct {
		srcCosmosAddr sdk.AccAddress
		srcValAddr    sdk.AccAddress
		srcETHAddr    common.Address
		srcETHSig     []byte
		expErr        bool
	}{
		"all good": {
			srcCosmosAddr: cosmosAddress,
			srcValAddr:    valAddress,
			srcETHAddr:    ethAddress,
			srcETHSig:     ethSig,
		},
		"empty validator address": {
			srcETHAddr:    ethAddress,
			srcCosmosAddr: cosmosAddress,
			srcETHSig:     ethSig,
			expErr:        true,
		},
		"invalid account address": {
			srcValAddr:    nil,
			srcCosmosAddr: cosmosAddress,
			srcETHAddr:    ethAddress,
			srcETHSig:     ethSig,
			expErr:        true,
		},
		"empty cosmos address": {
			srcValAddr: valAddress,
			srcETHAddr: ethAddress,
			srcETHSig:  ethSig,
			expErr:     true,
		},
		"invalid cosmos address": {
			srcCosmosAddr: nil,
			srcValAddr:    valAddress,
			srcETHAddr:    ethAddress,
			srcETHSig:     ethSig,
			expErr:        true,
		},
		"empty ethereum signature": {
			srcCosmosAddr: cosmosAddress,
			srcValAddr:    valAddress,
			srcETHAddr:    ethAddress,
			expErr:        true,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			msg := types.NewMsgSetOrchestratorAddress(spec.srcValAddr, spec.srcCosmosAddr, spec.srcETHAddr, spec.srcETHSig)

			err := msg.ValidateBasic()
			if spec.expErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
