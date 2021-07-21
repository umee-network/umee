package e2e

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

type orchestrator struct {
	index    int
	mnemonic string
	keyInfo  keyring.Info
}

func (o *orchestrator) instanceName() string {
	return fmt.Sprintf("orchestrator%d", o.index)
}
