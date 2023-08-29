package upgradev6x0

import (
	fmt "fmt"
	"strings"

	"github.com/umee-network/umee/v6/util/checkers"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ProposalTypeIBCMetadata = "IBCMetadata"
	RouterKey               = "gravity"
)

func (p *IBCMetadataProposal) GetTitle() string { return p.Title }

func (p *IBCMetadataProposal) GetDescription() string { return p.Description }

func (p *IBCMetadataProposal) ProposalRoute() string { return RouterKey }

func (p *IBCMetadataProposal) ProposalType() string {
	return ProposalTypeIBCMetadata
}

func (p *IBCMetadataProposal) ValidateBasic() error {
	return nil
}

func (p IBCMetadataProposal) String() string {
	decimals := uint32(0)
	for _, denomUnit := range p.Metadata.DenomUnits {
		if denomUnit.Denom == p.Metadata.Display {
			decimals = denomUnit.Exponent
			break
		}
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf(`IBC Metadata setting proposal:
  Title:             %s
  Description:       %s
  Token Name:        %s
  Token Symbol:      %s
  Token Display:     %s
  Token Decimals:    %d
  Token Description: %s
`, p.Title, p.Description, p.Metadata.Name, p.Metadata.Symbol, p.Metadata.Display, decimals, p.Metadata.Description))
	return b.String()
}

// GetTitle returns the title of the proposal.
func (msg *MsgGovUpdateRegistry) GetTitle() string { return msg.Title }

// GetDescription returns the description of the proposal.
func (msg *MsgGovUpdateRegistry) GetDescription() string { return msg.Description }

// Type implements Msg
func (msg MsgGovUpdateRegistry) Type() string { return sdk.MsgTypeURL(&msg) }

// ValidateBasic implements Msg
func (msg MsgGovUpdateRegistry) ValidateBasic() error {
	return nil
}

// GetSignBytes implements Msg
func (msg MsgGovUpdateRegistry) GetSignBytes() []byte {
	return []byte{}
}

// GetSigners implements Msg
func (msg MsgGovUpdateRegistry) GetSigners() []sdk.AccAddress {
	return checkers.Signers(msg.Authority)
}
