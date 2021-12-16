package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	leveragetypes "github.com/umee-network/umee/x/leverage/types"
	"github.com/umee-network/umee/x/oracle/types"
)

// Hooks defines a structure around the x/oracle Keeper that implements various
// Hooks interface defined by other modules such as x/leverage.
type Hooks struct {
	k Keeper
}

var _ leveragetypes.Hooks = Hooks{}

// Hooks returns a new Hooks instance that wraps the x/oracle keeper.
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// AfterTokenRegistered implements the x/leverage Hooks interface. Specifically,
// it checks if the provided Token should be added to the existing accepted list
// of assets for the x/oracle module.
func (h Hooks) AfterTokenRegistered(ctx sdk.Context, token leveragetypes.Token) {
	acceptList := h.k.AcceptList(ctx)

	var tokenExists bool
	for _, t := range acceptList {
		if t.BaseDenom == token.BaseDenom {
			tokenExists = true
			break
		}
	}

	if !tokenExists {
		acceptList = append(acceptList, types.Denom{
			BaseDenom:   token.BaseDenom,
			SymbolDenom: token.SymbolDenom,
			Exponent:    token.Exponent,
		})
	}

	h.k.SetAcceptList(ctx, acceptList)
}

// AfterRegisteredTokenRemoved implements the x/leverage Hooks interface. Currently,
// it performs a no-op, however, we may want to remove tokens from the accept
// list in the future when they're removed from the x/leverage Token registry.
func (h Hooks) AfterRegisteredTokenRemoved(ctx sdk.Context, token leveragetypes.Token) {}
