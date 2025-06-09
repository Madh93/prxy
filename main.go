// Package main is the entry point for the prxy application.
//
// It defines the command-line interface (CLI) using the urfave/cli library,
// handles configuration loading from flags and environment variables, sets up
// structured logging, and manages the lifecycle of the proxy server, including
// graceful shutdown.

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Madh93/prxy/internal/config"
	"github.com/Madh93/prxy/internal/logging"
	"github.com/Madh93/prxy/internal/prxy"
	"github.com/Madh93/prxy/internal/version"
	"github.com/urfave/cli/v3"
)

func init() {
	cli.VersionPrinter = func(cmd *cli.Command) {
		//nolint:errcheck
		fmt.Fprintf(cmd.Root().Writer, "%s %s\n", cmd.Root().Name, cmd.Root().Version)
	}
}

// main defines the CLI, initializes the configuration, sets up logging,
// and starts the application.
func main() {
	// More info at: https://cli.urfave.org/v3/getting-started/
	// TODO: flag validations (https://github.com/urfave/cli/issues/2132)
	cmd := &cli.Command{
		Name:                  config.AppName,
		Usage:                 "Forwards HTTP requests to a target via an external proxy",
		Version:               version.Get().String(),
		Suggest:               true,
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "target", Required: true, Usage: "target service URL", Sources: cli.EnvVars("PRXY_TARGET"), Aliases: []string{"t"}},
			&cli.StringFlag{Name: "proxy", Required: true, Usage: "outbound HTTP Proxy URL", Sources: cli.EnvVars("PRXY_PROXY"), Aliases: []string{"x"}},
			&cli.StringFlag{Name: "host", Value: config.Defaults.Host, Usage: "host to listen on", Sources: cli.EnvVars("PRXY_HOST"), Aliases: []string{"H"}},
			&cli.IntFlag{Name: "port", Value: config.Defaults.Port, Usage: "port to listen on", DefaultText: "random", Sources: cli.EnvVars("PRXY_PORT"), Aliases: []string{"P"}},
			&cli.StringFlag{Name: "log-level", Value: string(config.Defaults.Logging.Level), Usage: fmt.Sprintf("set log level. Available options: %s", config.ValidLogLevels), Sources: cli.EnvVars("PRXY_LOG_LEVEL"), Aliases: []string{"l"}},
			&cli.StringFlag{Name: "log-format", Value: string(config.Defaults.Logging.Format), Usage: fmt.Sprintf("set log format. Available options: %s", config.ValidLogFormats), Sources: cli.EnvVars("PRXY_LOG_FORMAT"), Aliases: []string{"f"}},
			&cli.StringFlag{Name: "log-output", Value: string(config.Defaults.Logging.Output), Usage: fmt.Sprintf("set log output. Available options: %s", config.ValidLogOutputs), Sources: cli.EnvVars("PRXY_LOG_OUTPUT"), Aliases: []string{"o"}},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// Load configuration
			cfg, err := config.New(cmd)
			if err != nil {
				return fmt.Errorf("failed to initialize configuration: %v", err)
			}

			// Setup logger
			logger, err := logging.New(&cfg.Logging)
			if err != nil {
				return fmt.Errorf("failed to initialize logger: %v", err)
			}

			logger.Debug("Configuration loaded successfully", "config", cfg)

			// Setup prxy
			logger.Info("Starting prxy...", version.Get().ToLogFields()...)
			prxyServer, err := prxy.New(cfg, logger)
			if err != nil {
				return fmt.Errorf("failed to create proxy server: %v", err)
			}

			// Handling graceful shutdown with signals. Create a context that
			// listens for the interrupt signal.
			// More info at: https://henvic.dev/posts/signal-notify-context/
			signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM) // TODO: https://pkg.go.dev/os/signal#hdr-Windows
			defer stop()

			// Run the server in a separate goroutine so that it doesn't block.
			errChan := make(chan error, 1)
			go func() {
				logger.Info("Server starting to listen...", "address", prxyServer.Addr(), "target", cfg.Target, "proxy", cfg.Proxy)
				errChan <- prxyServer.Run()
			}()

			// Block until we receive a signal or the server exits with an error.
			select {
			case err := <-errChan:
				if err != nil {
					return fmt.Errorf("server stopped with an error: %v", err)
				}
				logger.Info("Server stopped gracefully.")
			case <-signalCtx.Done():
				stop() // Clean up the signal notifier.
				logger.Info("Shutdown signal received. Shutting down gracefully...")
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := prxyServer.Shutdown(shutdownCtx); err != nil {
					return fmt.Errorf("error during graceful shutdown: %v", err)
				}
				logger.Info("All done! prxy has been shut down.")
			}

			// Cleanly close the logger before exiting.
			if cerr := logger.Close(); cerr != nil {
				return fmt.Errorf("failed to close log file (%s): %v", cfg.Logging.Path, cerr)
			}

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
