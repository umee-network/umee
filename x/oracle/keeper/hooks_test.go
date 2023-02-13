package keeper_test

import (
	appparams "github.com/umee-network/umee/v4/app/params"
	leveragetypes "github.com/umee-network/umee/v4/x/leverage/types"
)

func (s *IntegrationTestSuite) TestHooks_AfterTokenRegistered() {
	h := s.app.OracleKeeper.Hooks()
	s.Require().Len(s.app.OracleKeeper.AcceptList(s.ctx), 2)

	// require that an existing token does not change the accept list
	h.AfterTokenRegistered(s.ctx, leveragetypes.Token{
		BaseDenom:   appparams.BondDenom,
		SymbolDenom: appparams.DisplayDenom,
		Exponent:    6,
	})
	s.Require().Len(s.app.OracleKeeper.AcceptList(s.ctx), 2)

	// require a new registered token updates the accept list
	h.AfterTokenRegistered(s.ctx, leveragetypes.Token{
		BaseDenom:   "ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518",
		SymbolDenom: "OSMO",
		Exponent:    6,
	})
	s.Require().Len(s.app.OracleKeeper.AcceptList(s.ctx), 3)

	// require a blacklisted token does not update the accept list
	h.AfterTokenRegistered(s.ctx, leveragetypes.Token{
		BaseDenom:   "unope",
		SymbolDenom: "NOPE",
		Exponent:    6,
		Blacklist:   true,
	})
	s.Require().Len(s.app.OracleKeeper.AcceptList(s.ctx), 3)
}
