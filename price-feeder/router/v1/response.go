package v1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Response constants
const (
	StatusAvailable = "available"
)

type (
	// HealthZResponse defines the response type for the healthy API handler.
	HealthZResponse struct {
		Status string `json:"status" yaml:"status"`
		Oracle struct {
			LastSync string `json:"last_sync"`
		} `json:"oracle"`
	}

	// PricesResponse defines the response type for getting the latest exchange
	// rates from the oracle.
	PricesResponse struct {
		Prices map[string]sdk.Dec `json:"prices"`
	}
)
