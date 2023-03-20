package incentive

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v4/util/coin"
)

func TestMsgs(t *testing.T) {
	govAddr := "umee10d07y265gmmuvt4z0w9aw880jnsr700jg5w6jp"
	addr, err := sdk.AccAddressFromBech32("umee1s84d29zk3k20xk9f0hvczkax90l9t94g72n6wm")
	assert.NilError(t, err, "test address")

	tier := uint32(BondTierLong)
	uToken := sdk.NewInt64Coin("u/uumee", 10)
	token := sdk.NewInt64Coin("uumee", 10)
	program := NewIncentiveProgram(0, 4, 5, uToken.Denom, token, coin.Zero(token.Denom), false)

	msgs := []sdk.Msg{
		NewMsgBond(addr, tier, uToken),
		NewMsgBeginUnbonding(addr, tier, uToken),
		NewMsgClaim(addr),
		NewMsgSponsor(addr, 3, token),
		NewMsgGovCreatePrograms(govAddr, "title", "desc", []IncentiveProgram{program}),
		NewMsgGovSetParams(govAddr, "title", "desc", DefaultParams()),
	}

	for _, msg := range msgs {
		err := msg.ValidateBasic()
		assert.NilError(t, err, msg.String())
	}
}
