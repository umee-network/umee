package uibc

import "github.com/umee-network/umee/v3/util"

const (
	// ModuleName defines the module name
	ModuleName = "uibc"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for uibc
	RouterKey = ModuleName
)

var (
	KeyPrefixForIBCDenom = []byte{0x01}
	KeyTotalOutflowSum   = []byte{0x02}
	// ParamsKey is the key to query all gov params
	ParamsKey = []byte{0x03}
	// QuotaExpiresKey is the key to store the next expire time of ibc-transfer quota
	QuotaExpiresKey = []byte{0x04}
)

func CreateKeyForQuotaOfIBCDenom(ibcDenom string) []byte {
	//  keyPrefixForIBCDenom| denom | 0x00
	return util.ConcatBytes(0, KeyPrefixForIBCDenom, []byte(ibcDenom))
}
