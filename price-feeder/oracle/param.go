package oracle

import oracletypes "github.com/umee-network/umee/v2/x/oracle/types"

const (
	// paramsCacheInterval represents the amount of blocks
	// during which we will cache the oracle params.
	paramsCacheInterval = int64(200)
)

// ParamCache is used to cache oracle param data for
// an amount of blocks, defined by paramsCacheInterval.
type ParamCache struct {
	params           *oracletypes.Params
	lastUpdatedBlock int64
}

// Update retrieves the most recent oracle params and
// updates the instance.
func (onChainData *ParamCache) Update(currentBlockHeigh int64, params oracletypes.Params) {
	onChainData.lastUpdatedBlock = currentBlockHeigh
	onChainData.params = &params
}

// IsParamsOutdated checks whether or not the current
// param data was fetched in the last 200 blocks.
func (onChainData *ParamCache) IsParamsOutdated(currentBlockHeigh int64) bool {
	if onChainData.params == nil {
		return true
	}

	if currentBlockHeigh < paramsCacheInterval {
		return false
	}

	// This is an edge case, which should never happen.
	// The current blockchain height is lower
	// than the last updated block, to fix we should
	// just update the cached params again.
	if currentBlockHeigh < onChainData.lastUpdatedBlock {
		return true
	}

	return (currentBlockHeigh - onChainData.lastUpdatedBlock) > paramsCacheInterval
}
