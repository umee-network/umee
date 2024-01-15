package incentive_test

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v6/util/coin"
	"github.com/umee-network/umee/v6/x/incentive"
)

const (
	uumee   = "uumee"
	govAddr = "umee10d07y265gmmuvt4z0w9aw880jnsr700jg5w6jp"
)

var (
	testAddr, _ = sdk.AccAddressFromBech32("umee1s84d29zk3k20xk9f0hvczkax90l9t94g72n6wm")
	uToken      = sdk.NewInt64Coin(coin.UumeeDenom, 10)
	token       = sdk.NewInt64Coin(uumee, 10)
	program     = incentive.NewIncentiveProgram(0, 4, 5, uToken.Denom, token, coin.Zero(token.Denom), false)
)

func TestMsgs(t *testing.T) {
	t.Parallel()

	userMsgs := []sdk.LegacyMsg{
		incentive.NewMsgBond(testAddr, uToken),
		incentive.NewMsgBeginUnbonding(testAddr, uToken),
		incentive.NewMsgEmergencyUnbond(testAddr, uToken),
		incentive.NewMsgClaim(testAddr),
		incentive.NewMsgSponsor(testAddr, 3),
	}

	govMsgs := []sdk.LegacyMsg{
		incentive.NewMsgGovCreatePrograms(govAddr, []incentive.IncentiveProgram{program}),
		incentive.NewMsgGovSetParams(govAddr, incentive.DefaultParams()),
	}

	for _, msg := range userMsgs {
		if m, ok := msg.(sdk.HasValidateBasic); ok {
			err := m.ValidateBasic()
			assert.NilError(t, err, msg.String())
		}

		// check signers
		assert.Equal(t, len(msg.GetSigners()), 1)
		assert.Equal(t, msg.GetSigners()[0].String(), testAddr.String())
	}

	for _, msg := range govMsgs {
		if m, ok := msg.(sdk.HasValidateBasic); ok {
			err := m.ValidateBasic()
			assert.NilError(t, err, msg.String())
		}

		// check signers
		assert.Equal(t, len(msg.GetSigners()), 1)
		assert.Equal(t, msg.GetSigners()[0].String(), govAddr)
	}
}

func TestLegacyMsg(t *testing.T) {
	t.Parallel()

	msgs := []legacytx.LegacyMsg{
		incentive.NewMsgBond(testAddr, uToken),
		incentive.NewMsgBeginUnbonding(testAddr, uToken),
		incentive.NewMsgEmergencyUnbond(testAddr, uToken),
		incentive.NewMsgClaim(testAddr),
		incentive.NewMsgSponsor(testAddr, 3),
		incentive.NewMsgGovCreatePrograms(govAddr, []incentive.IncentiveProgram{program}),
		incentive.NewMsgGovSetParams(govAddr, incentive.DefaultParams()),
	}

	for _, msg := range msgs {
		assert.Assert(t, len(msg.GetSignBytes()) != 0)
		// assert.Equal(t,
		// 	addV1ToType(fmt.Sprintf("/umee.%T", msg)),
		// 	msg.Type(),
		// )
	}
}

// addV1ToType replaces "incentive." with "incentive.v1."
func addV1ToType(s string) string {
	return strings.Replace(s, "*incentive", "incentive.v1", 1)
}
