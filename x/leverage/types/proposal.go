package types

import (
	gov "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

var (
	proposalTypeMsgGovUpdateRegistry      = MsgGovUpdateRegistry{}.Type()
	proposalTypeMsgGovUpdateSpecialAssets = MsgGovUpdateSpecialAssets{}.Type()
)

func init() {
	gov.RegisterProposalType(proposalTypeMsgGovUpdateRegistry)
	gov.RegisterProposalType(proposalTypeMsgGovUpdateSpecialAssets)
}

// Implements Proposal Interface
var (
	_ gov.Content = &MsgGovUpdateRegistry{}
	_ gov.Content = &MsgGovUpdateSpecialAssets{}
)

func (msg *MsgGovUpdateRegistry) GetTitle() string       { return msg.Title }
func (msg *MsgGovUpdateRegistry) GetDescription() string { return msg.Description }
func (msg *MsgGovUpdateRegistry) ProposalRoute() string  { return ModuleName }
func (msg *MsgGovUpdateRegistry) ProposalType() string {
	return proposalTypeMsgGovUpdateRegistry
}

func (msg *MsgGovUpdateSpecialAssets) GetTitle() string       { return "Special Asset Pairs" }
func (msg *MsgGovUpdateSpecialAssets) GetDescription() string { return "" }
func (msg *MsgGovUpdateSpecialAssets) ProposalRoute() string  { return ModuleName }
func (msg *MsgGovUpdateSpecialAssets) ProposalType() string {
	return proposalTypeMsgGovUpdateSpecialAssets
}
