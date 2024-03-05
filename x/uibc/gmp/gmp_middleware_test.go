package gmp

import (
	"encoding/json"
	"testing"

	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"
)

func TestGmpMemoHandler(t *testing.T) {
	gmpHandler := NewHandler()
	logger := log.NewNopLogger()
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, logger)

	// invalid memo
	invalidMemo := "invalid memp"
	err := gmpHandler.OnRecvPacket(ctx, sdk.Coin{}, invalidMemo, nil)
	assert.ErrorContains(t, err, "invalid character")

	// valid memo
	validMemo := Message{
		SourceChain:   "source_chain",
		SourceAddress: "source_addr",
		Payload:       nil,
		Type:          int64(1),
	}
	m, err := json.Marshal(validMemo)
	assert.NilError(t, err)
	err = gmpHandler.OnRecvPacket(ctx, sdk.Coin{}, string(m), nil)
	assert.NilError(t, err)
}
