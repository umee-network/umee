package keeper

import "github.com/umee-network/umee/v4/util"

var (
	// Regular state
	keyPrefixIndex                 = []byte{0x01}
	keyPrefixBalances              = []byte{0x02}
	keyPrefixNextRebalancingTime   = []byte{0x03}
	keyPrefixNextInterestClaimTime = []byte{0x04}
	// keyPrefixParams is the key to query all gov params
	keyPrefixParams = []byte{0x05}
)

// keyIndex returns a KVStore key for index parameters for specific Index.
func keyIndex(meTokendenom string) []byte {
	// keyPrefixIndex | meTokendenom
	return util.ConcatBytes(1, keyPrefixIndex, []byte(meTokendenom))
}

// keyBalance returns a KVStore key for balane of a specific Index.
func keyBalance(meTokendenom string) []byte {
	// keyPrefixBalances | meTokendenom
	return util.ConcatBytes(1, keyPrefixBalances, []byte(meTokendenom))
}
