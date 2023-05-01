package incentive_test

import (
	"fmt"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v4/util/coin"
	"github.com/umee-network/umee/v4/x/incentive"
)

const (
	govAddr = "umee10d07y265gmmuvt4z0w9aw880jnsr700jg5w6jp"
)

var (
	testAddr, _ = sdk.AccAddressFromBech32("umee1s84d29zk3k20xk9f0hvczkax90l9t94g72n6wm")
	uToken      = sdk.NewInt64Coin("u/uumee", 10)
	token       = sdk.NewInt64Coin("uumee", 10)
	program     = incentive.NewIncentiveProgram(0, 4, 5, uToken.Denom, token, coin.Zero(token.Denom), false)
)

func TestMsgs(t *testing.T) {
	userMsgs := []sdk.Msg{
		incentive.NewMsgBond(testAddr, uToken),
		incentive.NewMsgBeginUnbonding(testAddr, uToken),
		incentive.NewMsgEmergencyUnbond(testAddr, uToken),
		incentive.NewMsgClaim(testAddr),
		incentive.NewMsgSponsor(testAddr, 3),
	}

	govMsgs := []sdk.Msg{
		incentive.NewMsgGovCreatePrograms(govAddr, []incentive.IncentiveProgram{program}),
		incentive.NewMsgGovSetParams(govAddr, incentive.DefaultParams()),
	}

	for _, msg := range userMsgs {
		err := msg.ValidateBasic()
		assert.NilError(t, err, msg.String())
		// check signers
		assert.Equal(t, len(msg.GetSigners()), 1)
		assert.Equal(t, msg.GetSigners()[0].String(), testAddr.String())
	}

	for _, msg := range govMsgs {
		err := msg.ValidateBasic()
		assert.NilError(t, err, msg.String())
		// check signers
		assert.Equal(t, len(msg.GetSigners()), 1)
		assert.Equal(t, msg.GetSigners()[0].String(), govAddr)
	}
}

// functions required in msgs.go which are not part of sdk.Msg
type sdkmsg interface {
	Route() string
	Type() string
	GetSignBytes() []byte
}

func TestRoutes(t *testing.T) {
	msgs := []sdkmsg{
		*incentive.NewMsgBond(testAddr, uToken),
		*incentive.NewMsgBeginUnbonding(testAddr, uToken),
		*incentive.NewMsgEmergencyUnbond(testAddr, uToken),
		*incentive.NewMsgClaim(testAddr),
		*incentive.NewMsgSponsor(testAddr, 3),
		*incentive.NewMsgGovCreatePrograms(govAddr, []incentive.IncentiveProgram{program}),
		*incentive.NewMsgGovSetParams(govAddr, incentive.DefaultParams()),
	}

	for _, msg := range msgs {
		assert.Equal(t, "incentive", msg.Route())
		// check for non-empty returns for now
		assert.Assert(t, len(msg.GetSignBytes()) != 0)
		// exact match required
		assert.Equal(t,
			// example: "/umee.incentive.v1.MsgBond"
			// with %T returning "incentive.MsgBond"
			addV1ToType(fmt.Sprintf("/umee.%T", msg)),
			msg.Type(),
		)
	}
}

// addV1ToType replaces "incentive." with "incentive.v1."
func addV1ToType(s string) string {
	return strings.Replace(s, "incentive", "incentive.v1", 1)
}
