package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"gopkg.in/yaml.v3"
)

const (
	// ProposalTypeChange defines the type for a UpdateAssetsProposal
	ProposalTypeUpdateAssetsProposal = "UpdateAssetsProposal"
)

// Assert UpdateAssetsProposal implements govtypes.Content at compile-time
var _ govtypes.Content = &UpdateAssetsProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeUpdateAssetsProposal)
	govtypes.RegisterProposalTypeCodec(&UpdateAssetsProposal{}, "umee/UpdateAssetsProposal")
}

func NewUpdateAssetsProposal(title, description string, assets []Asset) *UpdateAssetsProposal {
	return &UpdateAssetsProposal{
		Title:       title,
		Description: description,
		Assets:      assets,
	}
}

// String implements the Stringer interface.
func (p UpdateAssetsProposal) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// GetTitle returns the title of the proposal.
func (p *UpdateAssetsProposal) GetTitle() string { return p.Title }

// GetDescription returns the description of the proposal.
func (p *UpdateAssetsProposal) GetDescription() string { return p.Description }

// ProposalRoute returns the x/gov routing key of the proposal.
func (p *UpdateAssetsProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the x/gov type of the proposal.
func (p *UpdateAssetsProposal) ProposalType() string { return ProposalTypeUpdateAssetsProposal }

// ValidateBasic validates the proposal returning an error if invalid.
func (p *UpdateAssetsProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	for _, asset := range p.Assets {
		if err := asset.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate performs validation on an Asset type returning an error if the Asset
// is invalid.
func (a Asset) Validate() error {
	if err := sdk.ValidateDenom(a.BaseTokenDenom); err != nil {
		return err
	}

	// TODO: Evaluate if we need additional constraints on the exchange rate.
	if a.ExchangeRate.IsNegative() || a.ExchangeRate.GT(sdk.OneDec()) {
		return fmt.Errorf("invalid exchange rate: %s", a.ExchangeRate)
	}

	// TODO: Evaluate if we need additional constraints on the collateral rate.
	if a.CollateralWeight.IsNegative() || a.CollateralWeight.GT(sdk.OneDec()) {
		return fmt.Errorf("invalid collateral rate: %s", a.CollateralWeight)
	}

	// TODO: Evaluate if we need additional constraints on the base borrow rate.
	if a.BaseBorrowRate.IsNegative() || a.BaseBorrowRate.GT(sdk.OneDec()) {
		return fmt.Errorf("invalid base borrow rate: %s", a.BaseBorrowRate)
	}

	return nil
}
