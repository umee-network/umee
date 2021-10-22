package v1

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"

	"github.com/umee-network/umee/price-feeder/oracle"
	"github.com/umee-network/umee/price-feeder/pkg/httputil"
)

const (
	V1APIPathPrefix = "/api/v1"
)

// RegisterRoutes register v1 API routes on the provided sub-router.
func RegisterRoutes(v1SubRtr *mux.Router, oracle *oracle.Oracle, mChain alice.Chain) {
	v1SubRtr.Handle(
		"/healthz",
		mChain.Then(healthzHandler(oracle)),
	).Methods(httputil.MethodGET)

	v1SubRtr.Handle(
		"/prices",
		mChain.Then(pricesHandler(oracle)),
	).Methods(httputil.MethodGET)
}

func healthzHandler(oracle *oracle.Oracle) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		resp := HealthZResponse{
			Status: StatusAvailable,
		}
		resp.Oracle.LastSync = oracle.GetLastPriceSyncTimestamp().Format(time.RFC3339)

		httputil.RespondWithJSON(w, http.StatusOK, resp)
	}
}

func pricesHandler(oracle *oracle.Oracle) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		resp := PricesResponse{
			Prices: oracle.GetPrices(),
		}

		httputil.RespondWithJSON(w, http.StatusOK, resp)
	}
}
