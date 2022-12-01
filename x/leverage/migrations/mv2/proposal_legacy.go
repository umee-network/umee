package mv2

import (
	gov1b1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// Assert UpdateRegistryProposal implements govtypes.Content at compile-time
var _ gov1b1.Content = &UpdateRegistryProposal{}

// String implements the Stringer interface.
func (p UpdateRegistryProposal) String() string { return "not needed" }

// GetTitle returns the title of the proposal.
func (p *UpdateRegistryProposal) GetTitle() string { return p.Title }

// GetDescription returns the description of the proposal.
func (p *UpdateRegistryProposal) GetDescription() string { return p.Description }

// ProposalRoute returns the x/gov routing key of the proposal.
func (p *UpdateRegistryProposal) ProposalRoute() string { return "not needed" }

// ProposalType returns the x/gov type of the proposal.
func (p *UpdateRegistryProposal) ProposalType() string { return "not needed" }

// ValidateBasic validates the proposal returning an error if invalid.
func (p *UpdateRegistryProposal) ValidateBasic() error { return nil }
