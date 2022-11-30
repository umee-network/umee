package types

import (
	gov1b1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"gopkg.in/yaml.v3"
)

const (
	// ProposalTypeUpdateRegistryProposal defines the type for a UpdateRegistryProposal
	// proposal type.
	ProposalTypeUpdateRegistryProposal = "UpdateRegistryProposal"
)

func init() {
	gov1b1.RegisterProposalType(ProposalTypeUpdateRegistryProposal)
}

// Assert UpdateRegistryProposal implements govtypes.Content at compile-time
var _ gov1b1.Content = &UpdateRegistryProposal{}

func NewUpdateRegistryProposal(title, description string, tokens []Token) *UpdateRegistryProposal {
	return &UpdateRegistryProposal{
		Title:       title,
		Description: description,
		Registry:    tokens,
	}
}

// String implements the Stringer interface.
func (p UpdateRegistryProposal) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// GetTitle returns the title of the proposal.
func (p *UpdateRegistryProposal) GetTitle() string { return p.Title }

// GetDescription returns the description of the proposal.
func (p *UpdateRegistryProposal) GetDescription() string { return p.Description }

// ProposalRoute returns the x/gov routing key of the proposal.
func (p *UpdateRegistryProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the x/gov type of the proposal.
func (p *UpdateRegistryProposal) ProposalType() string { return ProposalTypeUpdateRegistryProposal }

// ValidateBasic validates the proposal returning an error if invalid.
func (p *UpdateRegistryProposal) ValidateBasic() error {
	err := gov1b1.ValidateAbstract(p)
	if err != nil {
		return err
	}

	for _, token := range p.Registry {
		if err := token.Validate(); err != nil {
			return err
		}
	}

	return nil
}
