package keeper

import (
	"github.com/umee-network/umee/v4/util"
)

var (
	keyPrefixDenomOutflows = []byte{0x01}
	keyTotalOutflows       = []byte{0x02}
	keyParams              = []byte{0x03}
	keyQuotaExpires        = []byte{0x04}
)

func KeyTotalOutflows(ibcDenom string) []byte {
	//  KeyPrefixDenomQuota | denom
	return util.ConcatBytes(0, keyPrefixDenomOutflows, []byte(ibcDenom))
}
