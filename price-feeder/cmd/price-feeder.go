package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/umee-network/umee/price-feeder/config"
)

const (
	logLevelJSON = "json"
	logLevelText = "text"
)

var (
	logLevel  string
	logFormat string

	wg sync.WaitGroup
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

	// add for the main process
	wg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())

	go startPriceFeeder(ctx, cfg)
	go startPriceOracle(ctx, cfg)

	// listen for and trap any OS signal to gracefully shutdown and exit
	trapSignal(cancel)

	// Block main process until all spawned goroutines have gracefully exited and
	// signal has been captured in the main process.
	wg.Wait()
	return nil
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
		wg.Done()
		cancel()
	}()
}

// TODO: This function started in goroutine is currently only boilerplate. It
// is subject to change in structure and behavior.
func startPriceFeeder(ctx context.Context, cfg config.Config) {
	wg.Add(1)
	defer wg.Done()

	log.Info().Str("listen_addr", cfg.ListenAddr).Msg("starting price-feeder...")

	for {
		select {
		case <-ctx.Done():
			return
		}
	}
}

// TODO: This function started in goroutine is currently only boilerplate. It
// is subject to change in structure and behavior.
func startPriceOracle(ctx context.Context, cfg config.Config) {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		}
	}
}
