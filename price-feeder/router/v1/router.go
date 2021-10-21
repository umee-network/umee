package v1

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"

	"github.com/umee-network/umee/price-feeder/oracle"
	"github.com/umee-network/umee/price-feeder/router/httputil"
)

const (
	V1APIPathPrefix = "/api/v1"
)

// RegisterRoutes register v1 API routes on the provided sub-router.
func RegisterRoutes(v1SubRtr *mux.Router, oracle oracle.Oracle, mChain alice.Chain) {
	v1SubRtr.Handle(
		"/healthz",
		mChain.Then(healthzHandler()),
	).Methods(httputil.MethodGET)
}

func healthzHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		httputil.RespondWithJSON(w, http.StatusOK, HealthZResponse{})
	}
}
