package v1

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
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

// errorResponse defines the attributes of a JSON error response.
type errorResponse struct {
	Code  int    `json:"code,omitempty"`
	Error string `json:"error"`
}

// newErrorResponse creates a new errorResponse instance.
func newErrorResponse(code int, err string) errorResponse {
	return errorResponse{Code: code, Error: err}
}

// writeErrorResponse prepares and writes a HTTP error
// given a status code and an error message.
func writeErrorResponse(w http.ResponseWriter, status int, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(legacy.Cdc.MustMarshalJSON(newErrorResponse(0, err)))
}
