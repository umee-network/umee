package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Hooks defines pre or post processing hooks for various actions the x/leverage
// module takes.
type Hooks interface {
	// AfterTokenRegistered defines a hook another keeper can execute after the
	// x/leverage registers a token.
	AfterTokenRegistered(ctx sdk.Context, token Token)

	// AfterRegisteredTokenRemoved defines a hook another keeper can execute after
	// the x/leverage module removes a registered token.
	AfterRegisteredTokenRemoved(ctx sdk.Context, token Token)
}

var _ Hooks = MultiHooks{}

// MultiHooks defines a type alias for multiple hooks, i.e. for multiple keepers
// to execute x/leverage hooks.
type MultiHooks []Hooks

func NewMultiHooks(hooks ...Hooks) MultiHooks {
	return hooks
}

func (mh MultiHooks) AfterTokenRegistered(ctx sdk.Context, token Token) {
	for _, h := range mh {
		h.AfterTokenRegistered(ctx, token)
	}
}

func (mh MultiHooks) AfterRegisteredTokenRemoved(ctx sdk.Context, token Token) {
	for _, h := range mh {
		h.AfterRegisteredTokenRemoved(ctx, token)
	}
}
