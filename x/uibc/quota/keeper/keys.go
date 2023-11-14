package keeper

import (
	"github.com/umee-network/umee/v6/util"
)

var (
	keyPrefixDenomOutflows = []byte{0x01}
	keyTotalOutflows       = []byte{0x02}
	keyParams              = []byte{0x03}
	keyQuotaExpires        = []byte{0x04}
	keyPrefixDenomInflows  = []byte{0x05}
	keyTotalInflows        = []byte{0x06}
)

func keyTotalOutflow(ibcDenom string) []byte {
	//  keyPrefixDenomOutflows | denom
	return util.ConcatBytes(0, keyPrefixDenomOutflows, []byte(ibcDenom))
}

func keyTokenInflow(ibcDenom string) []byte {
	//  keyPrefixDenomInflows | denom
	return util.ConcatBytes(0, keyPrefixDenomInflows, []byte(ibcDenom))
}

// denomFromKey extracts denom from a key with the form
// prefix | denom
func denomFromKey(key, prefix []byte) string {
	return string(key[len(prefix):])
}
