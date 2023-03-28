package incentive

import (
	gov "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

var (
	proposalTypeMsgGovSetParams      = MsgGovSetParams{}.Type()
	proposalTypeMsgGovCreatePrograms = MsgGovCreatePrograms{}.Type()
)

var (
	// Ensure we implement Proposal Interface
	_ gov.Content = &MsgGovSetParams{}
	_ gov.Content = &MsgGovCreatePrograms{}
)

func init() {
	gov.RegisterProposalType(proposalTypeMsgGovSetParams)
	gov.RegisterProposalType(proposalTypeMsgGovCreatePrograms)
}

// GetTitle implements gov.Content
func (msg *MsgGovSetParams) GetTitle() string { return msg.Title }

// GetDescription implements gov.Content
func (msg *MsgGovSetParams) GetDescription() string { return msg.Description }

// GetDescription implements gov.Content
func (msg *MsgGovSetParams) ProposalRoute() string { return RouterKey }

// ProposalType implements gov.Content
func (msg *MsgGovSetParams) ProposalType() string { return proposalTypeMsgGovSetParams }

// GetTitle implements gov.Content
func (msg *MsgGovCreatePrograms) GetTitle() string { return msg.Title }

// GetDescription implements gov.Content
func (msg *MsgGovCreatePrograms) GetDescription() string { return msg.Description }

// GetDescription implements gov.Content
func (msg *MsgGovCreatePrograms) ProposalRoute() string { return RouterKey }

// ProposalType implements gov.Content
func (msg *MsgGovCreatePrograms) ProposalType() string { return proposalTypeMsgGovSetParams }
