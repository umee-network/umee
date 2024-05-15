package metoken

// DefaultParams returns default genesis params
func DefaultParams() Params {
	return Params{
		RebalancingFrequency:    60 * 60 * 12,     // 12h
		ClaimingFrequency:       60 * 60 * 24 * 7, // 7d
		RewardsAuctionFeeFactor: 2000,             // 20% of fees goes to rewards auction
	}
}
