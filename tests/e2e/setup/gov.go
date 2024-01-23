package setup

import (
	"time"

	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// GovVoteAndWait votes for a given proposal with provided list of clients and waits for a proposal to pass.
func (s *E2ETestSuite) GovVoteAndWait(propID uint64) {
	assert := s.Assert()
	for _, ta := range s.Chain.TestAccounts {
		err := ta.client.GovVote(propID, nil)
		assert.NoError(err)
	}

	c := s.AccountClient(0)
	prop, err := c.GovProposal(propID)
	assert.NoError(err)

	now := time.Now()
	sleepDuration := prop.VotingEndTime.Sub(now) + 1*time.Second
	s.T().Log("sleeping ", sleepDuration, " until end of voting period + 1 block\n")
	time.Sleep(sleepDuration)

	prop, err = c.GovProposal(propID)
	assert.NoError(err)

	assert.Equal(govv1.ProposalStatus_PROPOSAL_STATUS_PASSED, prop.Status,
		"proposal %d didn't pass, status: %s", propID, prop.Status.String())
}
