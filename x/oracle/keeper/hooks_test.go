package keeper_test

import (
	umeeapp "github.com/umee-network/umee/v2/app"
	leveragetypes "github.com/umee-network/umee/v2/x/leverage/types"
)

func (s *IntegrationTestSuite) TestHooks_AfterTokenRegistered() {
	h := s.app.OracleKeeper.Hooks()
	s.Require().Len(s.app.OracleKeeper.AcceptList(s.ctx), 1)

	// TODO: These hooks need to respond to the Token.Blacklist field
	// (as well as update symbol denom and exponent if they ever change).
	// Blacklisting in particular should actually eliminate the base
	// denom from the oracle (and un-blacklisting, for whatever reason,
	// should do the opposite.)

	// require that an existing token does not change the accept list
	h.AfterTokenRegistered(s.ctx, leveragetypes.Token{
		BaseDenom:   umeeapp.BondDenom,
		SymbolDenom: umeeapp.DisplayDenom,
		Exponent:    6,
	})
	s.Require().Len(s.app.OracleKeeper.AcceptList(s.ctx), 1)

	// require a new registered token updates the accept list
	h.AfterTokenRegistered(s.ctx, leveragetypes.Token{
		BaseDenom:   "ibc/CDC4587874B85BEA4FCEC3CEA5A1195139799A1FEE711A07D972537E18FDA39D",
		SymbolDenom: "atom",
		Exponent:    6,
	})
	s.Require().Len(s.app.OracleKeeper.AcceptList(s.ctx), 2)
}
