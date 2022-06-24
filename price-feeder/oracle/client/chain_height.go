package client

import (
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	rpcclient "github.com/cosmos/cosmos-sdk/client/rpc"
)

const (
	// nodeStatsCacheInterval represents the amount of seconds
	// to update the node stats info from.
	nodeStatsCacheInterval = time.Second * 5
)

// ChainHeight is used to cache the chain height of the
// current node which keeps getting updated at the amount
// of interval nodeStatsCacheInterval.
type ChainHeight struct {
	clientCtx         client.Context
	mtx               sync.RWMutex
	errGetChainHeight error
	lastChainHeight   int64
}

var (
	// keep the instance thread safe.
	once sync.Once

	// keeps the single instance of ChainHeight.
	instance *ChainHeight
)

// ChainHeightInstance returns the single instance of ChainHeight.
func ChainHeightInstance(clientCtx client.Context) *ChainHeight {
	once.Do(func() {
		instance = &ChainHeight{
			clientCtx:         clientCtx,
			errGetChainHeight: nil,
			lastChainHeight:   0,
		}
		go instance.KeepUpdating()
	})

	return instance
}

// Update retrieves the most recent oracle params and
// updates the instance.
func (chainHeight *ChainHeight) UpdateChainHeight() {
	chainHeight.mtx.Lock()
	defer chainHeight.mtx.Unlock()

	blockHeight, err := rpcclient.GetChainHeight(chainHeight.clientCtx)
	if err != nil {
		chainHeight.errGetChainHeight = err
		return
	}
	chainHeight.lastChainHeight = blockHeight
	chainHeight.errGetChainHeight = nil
}

// KeepUpdating keeps asking for the current chain
// height from the node at each nodeStatsCacheInterval.
func (chainHeight *ChainHeight) KeepUpdating() {
	for {
		chainHeight.UpdateChainHeight()
		time.Sleep(nodeStatsCacheInterval)
	}
}

// GetChainHeight returns the last chain height available.
func (chainHeight *ChainHeight) GetChainHeight() (int64, error) {
	chainHeight.mtx.RLock()
	defer chainHeight.mtx.RUnlock()

	return chainHeight.lastChainHeight, chainHeight.errGetChainHeight
}
