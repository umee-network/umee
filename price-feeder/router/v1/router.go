package v1

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"

	"github.com/umee-network/umee/price-feeder/config"
	"github.com/umee-network/umee/price-feeder/pkg/httputil"
	"github.com/umee-network/umee/price-feeder/router/middleware"
)

const (
	APIPathPrefix = "/api/v1"
)

// Router defines a router wrapper used for registering v1 API routes.
type Router struct {
	logger zerolog.Logger
	cfg    config.Config
	oracle Oracle
}

func New(logger zerolog.Logger, cfg config.Config, oracle Oracle) *Router {
	return &Router{
		logger: logger.With().Str("module", "router").Logger(),
		cfg:    cfg,
		oracle: oracle,
	}
}

// RegisterRoutes register v1 API routes on the provided sub-router.
func (r *Router) RegisterRoutes(rtr *mux.Router, prefix string) {
	v1Router := rtr.PathPrefix(prefix).Subrouter()

	// build middleware chain
	mChain := middleware.Build(r.logger, r.cfg)

	// handle all preflight request
	v1Router.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.WriteHeader(http.StatusOK)
	})

	v1Router.Handle(
		"/healthz",
		mChain.ThenFunc(r.healthzHandler()),
	).Methods(httputil.MethodGET)

	v1Router.Handle(
		"/prices",
		mChain.ThenFunc(r.pricesHandler()),
	).Methods(httputil.MethodGET)
}

func (r *Router) healthzHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		resp := HealthZResponse{
			Status: StatusAvailable,
		}

		resp.Oracle.LastSync = r.oracle.GetLastPriceSyncTimestamp().Format(time.RFC3339)

		httputil.RespondWithJSON(w, http.StatusOK, resp)
	}
}

func (r *Router) pricesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		resp := PricesResponse{
			Prices: r.oracle.GetPrices(),
		}

		httputil.RespondWithJSON(w, http.StatusOK, resp)
	}
}
