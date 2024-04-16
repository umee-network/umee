package keeper

import (
	"github.com/umee-network/umee/v6/util"
)

var (
	keyRewardsParams    = []byte{0x01}
	keyRwardsCurrentID  = []byte{0x02}
	keyPrefixRewardsBid = []byte{0x03}
)

func (k Keeper) keyRewardsBid(id uint32) []byte {
	return util.KeyWithUint32(keyPrefixRewardsBid, id)
}
