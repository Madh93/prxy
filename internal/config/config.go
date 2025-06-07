// Package config manages application configuration using the Koanf library.
//
// This package provides structures and functions for handling application
// configuration settings. It uses the Koanf library to facilitate loading
// configuration from various sources.
//
// The main structures include:
//
//   - Config: Represents the overall configuration object, containing nested
//     configurations for Host, Port, Proxy URL and Target URL, and Logging settings.
//
//   - LoggingConfig: Holds logging configuration settings, including the log level,
//     format, output destination, and path for log files. It includes validation
//     to ensure the logging settings are correct and conform to allowed values.
//
// The package also provides a New function to create a new configuration
// instance, initializing it with default values, loading settings from environment
// variables and processing command line flags. It ensures that settings are
// validated before they are used, enhancing the reliability of the application.
package config

import (
	"fmt"

	"github.com/Madh93/prxy/internal/validation"
	"github.com/knadh/koanf/providers/cliflagv3"
	"github.com/knadh/koanf/v2"
	"github.com/urfave/cli/v3"
)

// Config represents a configuration object. This type is
// designed to hold server and other configurations.
type Config struct {
	Target  string        `koanf:"target"` // Target service URL
	Proxy   string        `koanf:"proxy"`  // Outbound Proxy URL
	Host    string        `koanf:"host"`   // Server listening host
	Port    int           `koanf:"port"`   // Server listening port
	Logging LoggingConfig `koanf:"log"`    // Logging configuration
}

// AppName is the name of the application.
const AppName = "prxy"

// Defaults is the default configuration for the app.
var Defaults = Config{
	Host: "localhost",
	Port: 0,
	Logging: LoggingConfig{
		Level:  LogLevelInfo,
		Format: LogFormatText,
		Output: LogOutputStdout,
	},
}

// New loads the application configuration from various sources:
//   - Defaults
//   - Environment Variables
//   - Flags
func New(cmd *cli.Command) (*Config, error) {
	// Setup koanf
	k := koanf.New(".")

	// Load defaults
	cfg := Defaults

	// Load environment variables and flags
	if err := k.Load(cliflagv3.Provider(cmd, "-"), nil); err != nil {
		return nil, fmt.Errorf("failed to load CLI flags: %v", err)
	}

	// Unmarshal the loaded configuration
	if err := k.Unmarshal(AppName, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// Validate the configuration
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	return &cfg, nil
}

// validateConfig checks the validity of the configuration.
func validateConfig(cfg *Config) error {
	// Target URL
	if err := validation.ValidateURL(cfg.Target); err != nil {
		return fmt.Errorf("invalid target URL: %v", err)
	}

	// Proxy URL
	if err := validation.ValidateURL(cfg.Proxy); err != nil {
		return fmt.Errorf("invalid proxy URL: %v", err)
	}

	// Port
	if cfg.Port < 0 || cfg.Port > 65535 {
		return fmt.Errorf("invalid port: %d", cfg.Port)
	}

	// Logging
	if err := cfg.Logging.Validate(); err != nil {
		return err
	}

	return nil
}
