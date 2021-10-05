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

	// TODO: Evaluate if we need additional constraints on the exchange rate.
	if t.ExchangeRate.IsNegative() || t.ExchangeRate.GT(sdk.OneDec()) {
		return fmt.Errorf("invalid exchange rate: %s", t.ExchangeRate)
	}

	// TODO: Evaluate if we need additional constraints on the collateral rate.
	if t.CollateralWeight.IsNegative() || t.CollateralWeight.GT(sdk.OneDec()) {
		return fmt.Errorf("invalid collateral rate: %s", t.CollateralWeight)
	}

	// TODO: Evaluate if we need additional constraints on the base borrow rate.
	if t.BaseBorrowRate.IsNegative() || t.BaseBorrowRate.GT(sdk.OneDec()) {
		return fmt.Errorf("invalid base borrow rate: %s", t.BaseBorrowRate)
	}

	return nil
}
