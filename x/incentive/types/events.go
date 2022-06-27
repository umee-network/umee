package types

// Event types and attributes for the incentive module
const (
	EventTypeClaimReward      = "claim_reward"
	EventTypeLockCollateral   = "lock_collateral"
	EventTypeUnlockCollateral = "unlock_collateral"
	EventTypeSponsorProgram   = "sponsor_program"

	EventAttrModule  = ModuleName
	EventAttrAddress = "address"
)
