package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/umee-network/umee/price-feeder/config"
	"github.com/umee-network/umee/price-feeder/oracle"
	"github.com/umee-network/umee/price-feeder/oracle/client"
	"github.com/umee-network/umee/price-feeder/router"
	"golang.org/x/sync/errgroup"
)

const (
	logLevelJSON = "json"
	logLevelText = "text"
)

var (
	logLevel  string
	logFormat string
)

var rootCmd = &cobra.Command{
	Use:   "price-feeder [config-file]",
	Args:  cobra.ExactArgs(1),
	Short: "price-feeder is a side-car process for providing Umee's on-chain oracle with price data",
	Long: `A side-car process that Umee validators must run in order to provide
Umee's on-chain price oracle with price information. The price-feeder performs
two primary functions. First, it is responsible for obtaining price information
from various reliable data sources, e.g. exchanges, and exposing this data via
an API. Secondly, the price-feeder consumes this data and periodically submits
vote and prevote messages following the oracle voting procedure.`,
	RunE: priceFeederCmdHandler,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", zerolog.InfoLevel.String(), "logging level")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", logLevelJSON, "logging format; must be either json or text")

	rootCmd.AddCommand(getVersionCmd())
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func priceFeederCmdHandler(cmd *cobra.Command, args []string) error {
	logLvl, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return err
	}

	zerolog.SetGlobalLevel(logLvl)

	switch logFormat {
	case logLevelJSON:
		// JSON is the default logging format

	case logLevelText:
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	default:
		return fmt.Errorf("invalid logging format: %s", logFormat)
	}

	cfg, err := config.ParseConfig(args[0])
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	// Set up chain

	timeout, err := strconv.Atoi(cfg.Rpc.RPCTimeout)

	if err != nil {
		panic("Invalid gRPC Timeout in config file")
	}

	cosmosChain, err := client.NewCosmosChain(
		cfg.ChainID,
		cfg.Keyring.Backend,
		cfg.Keyring.Dir,
		cfg.Keyring.Pass,
		cfg.Rpc.TMRPCEndpoint,
		time.Duration(timeout),
		cfg.Account.From,
		cfg.Account.Validator,
		cfg.GRPCEndpoint,
	)

	if err != nil {
		panic(err)
	}

	oc, err := client.NewOracleClient(cosmosChain)

	if err != nil {
		panic(err)
	}

	oracle := oracle.New(oc)

	g.Go(func() error {
		// start the process that observes and publishes exchange prices
		return startPriceFeeder(ctx, cfg, oracle)
	})
	g.Go(func() error {
		// Starts the process that calculates oracle prices
		// And also votes
		return startPriceOracle(ctx, oracle)
	})

	// listen for and trap any OS signal to gracefully shutdown and exit
	trapSignal(cancel)

	// Block main process until all spawned goroutines have gracefully exited and
	// signal has been captured in the main process or if an error occurs.
	return g.Wait()
}

// trapSignal will listen for any OS signal and invoke Done on the main
// WaitGroup allowing the main process to gracefully exit.
func trapSignal(cancel context.CancelFunc) {
	var sigCh = make(chan os.Signal)

	signal.Notify(sigCh, syscall.SIGTERM)
	signal.Notify(sigCh, syscall.SIGINT)

	go func() {
		sig := <-sigCh
		log.Info().Str("signal", sig.String()).Msg("caught signal; shutting down...")
		cancel()
	}()
}

func startPriceFeeder(ctx context.Context, cfg config.Config, oracle *oracle.Oracle) error {
	rtr := mux.NewRouter()
	rtrWrapper := router.New(cfg, rtr, oracle)
	rtrWrapper.RegisterRoutes()

	writeTimeout, err := time.ParseDuration(cfg.Server.WriteTimeout)
	if err != nil {
		return err
	}
	readTimeout, err := time.ParseDuration(cfg.Server.ReadTimeout)
	if err != nil {
		return err
	}

	srvErrCh := make(chan error, 1)
	srv := &http.Server{
		Handler:      rtr,
		Addr:         cfg.Server.ListenAddr,
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
	}

	go func() {
		log.Info().Str("listen_addr", cfg.Server.ListenAddr).Msg("starting price-feeder server...")
		srvErrCh <- srv.ListenAndServe()
	}()

	for {
		select {
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			log.Info().Str("listen_addr", cfg.Server.ListenAddr).Msg("shutting down price-feeder server...")
			if err := srv.Shutdown(shutdownCtx); err != nil {
				log.Error().Err(err).Msg("failed to gracefully shutdown price-feeder server")
				return err
			}

			return nil

		case err := <-srvErrCh:
			log.Error().Err(err).Msg("failed to start price-feeder server")
			return err
		}
	}
}

func startPriceOracle(ctx context.Context, oracle *oracle.Oracle) error {
	go func() {
		log.Info().Msg("starting price-feeder oracle...")
		oracle.Start(ctx)
	}()

	<-ctx.Done()
	log.Info().Msg("shutting down price-feeder oracle...")
	oracle.Stop()
	return nil
}
