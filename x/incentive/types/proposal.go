package types

import (
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"gopkg.in/yaml.v3"
)

const (
	// ProposalTypeCreateIncentiveProgramProposal defines the type for a CreateIncentiveProgramProposal
	// proposal type.
	ProposalTypeCreateIncentiveProgramProposal = "CreateIncentiveProgramProposal"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeCreateIncentiveProgramProposal)
	govtypes.RegisterProposalTypeCodec(&CreateIncentiveProgramProposal{}, "umee/CreateIncentiveProgramProposal")
}

// Assert CreateIncentiveProgramProposal implements govtypes.Content at compile-time
var _ govtypes.Content = &CreateIncentiveProgramProposal{}

func NewCreateIncentiveProgramProposal(title, description string, program IncentiveProgram) *CreateIncentiveProgramProposal {
	return &CreateIncentiveProgramProposal{
		Title:       title,
		Description: description,
		Program:     program,
	}
}

// String implements the Stringer interface.
func (p CreateIncentiveProgramProposal) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// GetTitle returns the title of the proposal.
func (p *CreateIncentiveProgramProposal) GetTitle() string { return p.Title }

// GetDescription returns the description of the proposal.
func (p *CreateIncentiveProgramProposal) GetDescription() string { return p.Description }

// ProposalRoute returns the x/gov routing key of the proposal.
func (p *CreateIncentiveProgramProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the x/gov type of the proposal.
func (p *CreateIncentiveProgramProposal) ProposalType() string {
	return ProposalTypeCreateIncentiveProgramProposal
}

// ValidateBasic validates the proposal returning an error if invalid.
func (p *CreateIncentiveProgramProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	// TODO: Validate program
	return nil
}
