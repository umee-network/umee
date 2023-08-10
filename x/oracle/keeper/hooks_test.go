package keeper_test

import (
	appparams "github.com/umee-network/umee/v6/app/params"
	leveragetypes "github.com/umee-network/umee/v6/x/leverage/types"
)

func (s *IntegrationTestSuite) TestHooks_AfterTokenRegistered() {
	h := s.app.OracleKeeper.Hooks()
	s.Require().Len(s.app.OracleKeeper.AcceptList(s.ctx), 1)

	// require that an existing token does not change the accept list
	h.AfterTokenRegistered(s.ctx, leveragetypes.Token{
		BaseDenom:   appparams.BondDenom,
		SymbolDenom: appparams.DisplayDenom,
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

	// require a blacklisted token does not update the accept list
	h.AfterTokenRegistered(s.ctx, leveragetypes.Token{
		BaseDenom:   "unope",
		SymbolDenom: "NOPE",
		Exponent:    6,
		Blacklist:   true,
	})
	s.Require().Len(s.app.OracleKeeper.AcceptList(s.ctx), 2)
}
