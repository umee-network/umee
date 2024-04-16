package keeper

import (
	"github.com/umee-network/umee/v6/util"
	"github.com/umee-network/umee/v6/x/auction"
)

var (
	keyRewardsParams    = []byte{0x01}
	keyRwardsCurrentID  = []byte{0x02}
	keyPrefixRewardsBid = []byte{0x03}
)

func (k Keeper) keyRewardsBid(id uint32) []byte {
	return util.ConcatBytes(0, keyPrefixRewardsBid)
}
