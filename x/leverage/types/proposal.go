package types

import (
	gov "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

var (
	proposalTypeMsgGovUpdateRegistry          = MsgGovUpdateRegistry{}.Type()
	proposalTypeMsgGovUpdateSpecialAssetPairs = MsgGovUpdateSpecialAssetPairs{}.Type()
)

func init() {
	gov.RegisterProposalType(proposalTypeMsgGovUpdateRegistry)
	gov.RegisterProposalType(proposalTypeMsgGovUpdateSpecialAssetPairs)
}

// Implements Proposal Interface
var (
	_ gov.Content = &MsgGovUpdateRegistry{}
	_ gov.Content = &MsgGovUpdateSpecialAssetPairs{}
)

func (msg *MsgGovUpdateRegistry) GetTitle() string       { return msg.Title }
func (msg *MsgGovUpdateRegistry) GetDescription() string { return msg.Description }
func (msg *MsgGovUpdateRegistry) ProposalRoute() string  { return ModuleName }
func (msg *MsgGovUpdateRegistry) ProposalType() string {
	return proposalTypeMsgGovUpdateRegistry
}

func (msg *MsgGovUpdateSpecialAssetPairs) GetTitle() string       { return "Special Asset Pairs" }
func (msg *MsgGovUpdateSpecialAssetPairs) GetDescription() string { return "" }
func (msg *MsgGovUpdateSpecialAssetPairs) ProposalRoute() string  { return ModuleName }
func (msg *MsgGovUpdateSpecialAssetPairs) ProposalType() string {
	return proposalTypeMsgGovUpdateSpecialAssetPairs
}
