package types

import (
	gov "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

var proposalTypeMsgGovUpdateRegistry = MsgGovUpdateRegistry{}.Type()

func init() {
	gov.RegisterProposalType(proposalTypeMsgGovUpdateRegistry)
}

// Implements Proposal Interface
var _ gov.Content = &MsgGovUpdateRegistry{}

// GetTitle returns the title of a community pool spend proposal.
func (msg *MsgGovUpdateRegistry) GetTitle() string { return msg.Title }

// GetDescription returns the description of a community pool spend proposal.
func (msg *MsgGovUpdateRegistry) GetDescription() string { return msg.Description }

// GetDescription returns the routing key of a community pool spend proposal.
func (msg *MsgGovUpdateRegistry) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a community pool spend proposal.
func (msg *MsgGovUpdateRegistry) ProposalType() string { return proposalTypeMsgGovUpdateRegistry }
