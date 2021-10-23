package types_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/umee-network/umee/x/peggy/types"

	"github.com/ethereum/go-ethereum/common"
)

func TestValsetConfirmSig(t *testing.T) {
	var (
		correctSig = "e108a7776de6b87183b0690484a74daef44aa6daf907e91abaf7bbfa426ae7706b12e0bd44ef7b0634710d99c2d81087a2f39e075158212343a3b2948ecf33d01c"
		invalidSig = "fffff7776de6b87183b0690484a74daef44aa6daf907e91abaf7bbfa426ae7706b12e0bd44ef7b0634710d99c2d81087a2f39e075158212343a3b2948ecf33d01c"
		ethAddress = common.HexToAddress("0xc783df8a850f42e7f7e57013759c285caa701eb6")
		hash       = common.HexToHash("0x88165860d955aee7dc3e83d9d1156a5864b708841965585d206dbef6e9e1a499")
	)

	specs := map[string]struct {
		srcHash      common.Hash
		srcSignature string
		srcETHAddr   common.Address
		expErr       bool
	}{
		"all good": {
			srcHash:      hash,
			srcSignature: correctSig,
			srcETHAddr:   ethAddress,
		},
		"invalid signature": {
			srcHash:      hash,
			srcSignature: invalidSig,
			srcETHAddr:   ethAddress,
			expErr:       true,
		},
		"empty signature": {
			srcHash:    hash,
			srcETHAddr: ethAddress,
			expErr:     true,
		},
		"signature too short": {
			srcHash:      hash,
			srcSignature: correctSig[0:64],
			srcETHAddr:   ethAddress,
			expErr:       true,
		},
		"empty eth address": {
			srcHash:      hash,
			srcSignature: correctSig,
			expErr:       true,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			sigBytes, err := hex.DecodeString(spec.srcSignature)
			require.NoError(t, err)

			// when
			err = ValidateEthereumSignature(spec.srcHash, sigBytes, spec.srcETHAddr)
			if spec.expErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
