package router

import (
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/umee-network/umee/price-feeder/config"
	"github.com/umee-network/umee/price-feeder/oracle"
	"github.com/umee-network/umee/price-feeder/router/middleware"
	v1router "github.com/umee-network/umee/price-feeder/router/v1"
)

// Router defines a router wrapper used for registering API routes.
type Router struct {
	logger zerolog.Logger
	cfg    config.Config
	oracle oracle.Oracle
	rtr    *mux.Router
}

func New(cfg config.Config, rtr *mux.Router, oracle oracle.Oracle) Router {
	return Router{
		logger: log.With().Str("module", "router").Logger(),
		cfg:    cfg,
		oracle: oracle,
		rtr:    rtr,
	}
}

// RegisterRoutes registers API routes on all supported versioned API sub-paths.
func (r Router) RegisterRoutes() {
	mChain := middleware.Build(r.logger, r.cfg)

	v1SubRtr := r.rtr.PathPrefix(v1router.V1APIPathPrefix).Subrouter()
	v1router.RegisterRoutes(v1SubRtr, r.oracle, mChain)
}
