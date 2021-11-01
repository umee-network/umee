package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"gopkg.in/yaml.v3"
)

const (
	// UTokenPrefix defines the uToken denomination prefix for all uToken types.
	UTokenPrefix = "u/"

	// ProposalTypeUpdateRegistryProposal defines the type for a UpdateRegistryProposal
	// proposal type.
	ProposalTypeUpdateRegistryProposal = "UpdateRegistryProposal"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeUpdateRegistryProposal)
	govtypes.RegisterProposalTypeCodec(&UpdateRegistryProposal{}, "umee/UpdateRegistryProposal")
}

// UTokenFromTokenDenom returns the uToken denom given a token denom.
func UTokenFromTokenDenom(tokenDenom string) string {
	return UTokenPrefix + tokenDenom
}

// Assert UpdateRegistryProposal implements govtypes.Content at compile-time
var _ govtypes.Content = &UpdateRegistryProposal{}

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
	err := govtypes.ValidateAbstract(p)
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

// Validate performs validation on an Token type returning an error if the Token
// is invalid.
func (t Token) Validate() error {
	if err := sdk.ValidateDenom(t.BaseDenom); err != nil {
		return err
	}

	// Reserve factor and collateral weight range between 0 and 1, inclusive.
	if t.ReserveFactor.IsNegative() || t.ReserveFactor.GT(sdk.OneDec()) {
		return fmt.Errorf("invalid reserve factor: %s", t.ReserveFactor)
	}
	if t.CollateralWeight.IsNegative() || t.CollateralWeight.GT(sdk.OneDec()) {
		return fmt.Errorf("invalid collateral rate: %s", t.CollateralWeight)
	}

	// Kink utilization rate ranges between 0 and 1, exclusive. This prevents multiple interest rates being
	// defined at exactly 0% or 100% utilization (e.g. kink at 0%, 2% base borrow rate, 4% borrow rate at kink.)
	if !t.KinkUtilizationRate.IsPositive() || t.KinkUtilizationRate.GTE(sdk.OneDec()) {
		return fmt.Errorf("invalid kink utilization rate: %s", t.KinkUtilizationRate)
	}

	// Interest rates are non-negative. They do not need to have a maximum value.
	if t.BaseBorrowRate.IsNegative() {
		return fmt.Errorf("invalid base borrow rate: %s", t.BaseBorrowRate)
	}
	if t.KinkBorrowRate.IsNegative() {
		return fmt.Errorf("invalid kink borrow rate: %s", t.KinkBorrowRate)
	}
	if t.MaxBorrowRate.IsNegative() {
		return fmt.Errorf("invalid max borrow rate: %s", t.MaxBorrowRate)
	}

	return nil
}
