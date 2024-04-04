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
	logger := log.NewNopLogger()
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, logger)

	tests := []struct {
		name   string
		memo   func() string
		errMsg string
	}{
		{
			name: "empty memo",
			memo: func() string {
				return ""
			},
			errMsg: "",
		},
		{
			name: "invalid memo",
			memo: func() string {
				return "invalid memo"
			},
			errMsg: "invalid character",
		},
		{
			name: "unsupport msg",
			memo: func() string {
				validMemo := Memo{
					SourceChain:   "source_chain",
					SourceAddress: "source_addr",
					Payload:       nil,
					Type:          int64(4),
				}
				m, err := json.Marshal(validMemo)
				assert.NilError(t, err)
				return string(m)
			},
			errMsg: "unrecognized gmp message type",
		},
		{
			name: "unsupport msg",
			memo: func() string {
				validMemo := Memo{
					SourceChain:   "source_chain",
					SourceAddress: "source_addr",
					Payload:       nil,
					Type:          int64(1),
				}
				m, err := json.Marshal(validMemo)
				assert.NilError(t, err)
				return string(m)
			},
			errMsg: "msg.type=1 (TypeGeneralMessage) is not supported. Supported types include: 2",
		},
		{
			name: "valid memo",
			memo: func() string {
				validMemo := Memo{
					SourceChain:   "source_chain",
					SourceAddress: "source_addr",
					Payload:       nil,
					Type:          int64(2),
				}
				m, err := json.Marshal(validMemo)
				assert.NilError(t, err)
				return string(m)
			},
			errMsg: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseMemo(ctx, sdk.Coin{}, tc.memo(), nil)
			if len(tc.errMsg) != 0 {
				assert.ErrorContains(t, err, tc.errMsg)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}
