package quota

import (
	"github.com/umee-network/umee/v6/util"
)

var (
	keyPrefixDenomOutflows = []byte{0x01}
	keyOutflowSum          = []byte{0x02}
	keyParams              = []byte{0x03}
	keyQuotaExpires        = []byte{0x04}
	keyPrefixDenomInflows  = []byte{0x05}
	keyInflowSum           = []byte{0x06}
)

func keyTokenOutflow(ibcDenom string) []byte {
	//  keyPrefixDenomOutflows | denom
	return util.ConcatBytes(0, keyPrefixDenomOutflows, []byte(ibcDenom))
}

func keyTokenInflow(ibcDenom string) []byte {
	//  keyPrefixDenomInflows | denom
	return util.ConcatBytes(0, keyPrefixDenomInflows, []byte(ibcDenom))
}
