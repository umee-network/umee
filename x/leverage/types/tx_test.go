package types_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/umee-network/umee/v6/x/leverage/types"
	"gotest.tools/v3/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

const (
	denom  = "uumee"
	uDenom = "u/uumee"
)

var (
	testAddr, _ = sdk.AccAddressFromBech32("umee1s84d29zk3k20xk9f0hvczkax90l9t94g72n6wm")
	uToken      = sdk.NewInt64Coin(uDenom, 10)
	token       = sdk.NewInt64Coin(denom, 10)
)

func TestTxs(t *testing.T) {
	txs := []sdk.Msg{
		types.NewMsgSupply(testAddr, token),
		types.NewMsgWithdraw(testAddr, uToken),
		types.NewMsgMaxWithdraw(testAddr, denom),
		types.NewMsgCollateralize(testAddr, uToken),
		types.NewMsgSupplyCollateral(testAddr, token),
		types.NewMsgDecollateralize(testAddr, uToken),
		types.NewMsgBorrow(testAddr, token),
		types.NewMsgMaxBorrow(testAddr, denom),
		types.NewMsgRepay(testAddr, token),
		types.NewMsgLiquidate(testAddr, testAddr, token, uDenom),
		types.NewMsgLeveragedLiquidate(testAddr, testAddr, token.Denom, uDenom, sdk.OneDec()),
		types.NewMsgLeveragedLiquidate(testAddr, testAddr, "", "", sdk.ZeroDec()), // empty optional fields
	}

	for _, tx := range txs {
		err := tx.ValidateBasic()
		assert.NilError(t, err, tx.String())
		// check signers
		assert.Equal(t, len(tx.GetSigners()), 1)
		assert.Equal(t, tx.GetSigners()[0].String(), testAddr.String())
	}
}

func TestLegacyMsg(t *testing.T) {
	txs := []legacytx.LegacyMsg{
		types.NewMsgSupply(testAddr, token),
		types.NewMsgWithdraw(testAddr, uToken),
		types.NewMsgMaxWithdraw(testAddr, denom),
		types.NewMsgCollateralize(testAddr, uToken),
		types.NewMsgSupplyCollateral(testAddr, token),
		types.NewMsgDecollateralize(testAddr, uToken),
		types.NewMsgBorrow(testAddr, token),
		types.NewMsgMaxBorrow(testAddr, denom),
		types.NewMsgRepay(testAddr, token),
		types.NewMsgLiquidate(testAddr, testAddr, token, uDenom),
		types.NewMsgLeveragedLiquidate(testAddr, testAddr, token.Denom, uDenom, sdk.OneDec()),
	}

	for _, tx := range txs {
		assert.Assert(t, len(tx.GetSignBytes()) != 0)
		assert.Equal(t,
			addV1ToType(fmt.Sprintf("/umee.%T", tx)),
			tx.Type(),
		)
	}
}

func addV1ToType(s string) string {
	return strings.Replace(s, "*types", "leverage.v1", 1)
}
