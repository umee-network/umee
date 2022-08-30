package ante

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	evidence "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/stretchr/testify/assert"

	leverage "github.com/umee-network/umee/v3/x/leverage/types"
)

func TestPriority(t *testing.T) {
	tcs := []struct {
		name     string
		oracle   bool
		msgs     []sdk.Msg
		priority int64
	}{
		{"empty priority 0", false, []sdk.Msg{}, 0},
		{"when oracleOrGravity is set, then tx is max", true, []sdk.Msg{}, 100},
		{"evidence1", true, []sdk.Msg{&evidence.MsgSubmitEvidence{}}, 100},
		{"evidence2", false, []sdk.Msg{&evidence.MsgSubmitEvidence{}}, 90},
		{"evidence3", false, []sdk.Msg{&evidence.MsgSubmitEvidence{}, &evidence.MsgSubmitEvidence{}}, 90},
		{"leverage1", false, []sdk.Msg{&leverage.MsgLiquidate{}}, 80},
		{"leverage-evidence1", false, []sdk.Msg{&leverage.MsgLiquidate{}, &evidence.MsgSubmitEvidence{}}, 80},
		{"leverage-evidence2", false, []sdk.Msg{&evidence.MsgSubmitEvidence{}, &leverage.MsgLiquidate{}}, 80},
		{"mixed1", false, []sdk.Msg{&evidence.MsgSubmitEvidence{}, &leverage.MsgLiquidate{}, &bank.MsgSend{}}, 0},
		{"mixed2", false, []sdk.Msg{&bank.MsgSend{}, &evidence.MsgSubmitEvidence{}, &leverage.MsgLiquidate{}}, 0},
	}

	for _, tc := range tcs {
		p := getTxPriority(tc.oracle, tc.msgs)
		assert.Equal(t, tc.priority, p, tc.name)
	}
}
