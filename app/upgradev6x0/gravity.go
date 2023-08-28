package upgradev6x0

import (
	fmt "fmt"
	"strings"

	"github.com/umee-network/umee/v6/util/checkers"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ProposalTypeUnhaltBridge = "UnhaltBridge"
	ProposalTypeAirdrop      = "Airdrop"
	ProposalTypeIBCMetadata  = "IBCMetadata"
	RouterKey                = "gravity"
)

func (p *UnhaltBridgeProposal) GetTitle() string { return p.Title }

func (p *UnhaltBridgeProposal) GetDescription() string { return p.Description }

func (p *UnhaltBridgeProposal) ProposalRoute() string { return RouterKey }

func (p *UnhaltBridgeProposal) ProposalType() string {
	return ProposalTypeUnhaltBridge
}

func (p *UnhaltBridgeProposal) ValidateBasic() error {
	return nil
}

func (p UnhaltBridgeProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Unhalt Bridge Proposal:
  Title:          %s
  Description:    %s
  target_nonce:   %d
`, p.Title, p.Description, p.TargetNonce))
	return b.String()
}

func (p *AirdropProposal) GetTitle() string { return p.Title }

func (p *AirdropProposal) GetDescription() string { return p.Description }

func (p *AirdropProposal) ProposalRoute() string { return RouterKey }

func (p *AirdropProposal) ProposalType() string {
	return ProposalTypeAirdrop
}

func (p *AirdropProposal) ValidateBasic() error {
	return nil
}

func (p AirdropProposal) String() string {
	var b strings.Builder
	total := uint64(0)
	for _, v := range p.Amounts {
		total += v
	}
	parsedRecipients := make([]sdk.AccAddress, len(p.Recipients)/20)
	for i := 0; i < len(p.Recipients)/20; i++ {
		indexStart := i * 20
		indexEnd := indexStart + 20
		addr := p.Recipients[indexStart:indexEnd]
		parsedRecipients[i] = addr
	}
	recipients := ""
	for i, a := range parsedRecipients {
		recipients += fmt.Sprintf("Account: %s Amount: %d%s", a.String(), p.Amounts[i], p.Denom)
	}

	b.WriteString(fmt.Sprintf(`Airdrop Proposal:
  Title:          %s
  Description:    %s
  Total Amount:   %d%s
  Recipients:     %s
`, p.Title, p.Description, total, p.Denom, recipients))
	return b.String()
}

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
