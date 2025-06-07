package config

import (
	"errors"
	"fmt"

	"github.com/Madh93/prxy/internal/validation"
)

// LogLevel defines the severity of a log entry.
type LogLevel string

// LogFormat defines the output format of logs.
type LogFormat string

// LogOutput defines the destination for log entries.
type LogOutput string

// LoggingConfig represents a configuration for logging.
type LoggingConfig struct {
	Level  LogLevel  `koanf:"level"`  // Log level
	Format LogFormat `koanf:"format"` // Log format
	Output LogOutput `koanf:"output"` // Output destination
	Path   string    `koanf:"path"`   // File path for logging output (if output is a file)
}

// Logging configuration values.
const (
	// Log levels.
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"

	// Log formats.
	LogFormatText LogFormat = "text"
	LogFormatJSON LogFormat = "json"

	// Output destinations.
	LogOutputStdout LogOutput = "stdout"
	LogOutputStderr LogOutput = "stderr"
	LogOutputFile   LogOutput = "file"
)

// Define typed slices of allowed values.
var (
	ValidLogLevels  = []LogLevel{LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError, LogLevelFatal}
	ValidLogFormats = []LogFormat{LogFormatText, LogFormatJSON}
	ValidLogOutputs = []LogOutput{LogOutputStdout, LogOutputStderr, LogOutputFile}
)

// Validate checks if the logging configuration is valid.
func (cfg LoggingConfig) Validate() error {
	var errs []error

	// Validate Level
	if err := validation.Validate(cfg.Level, ValidLogLevels); err != nil {
		errs = append(errs, fmt.Errorf("invalid log level: %v", err))
	}

	// Validate Format
	if err := validation.Validate(cfg.Format, ValidLogFormats); err != nil {
		errs = append(errs, fmt.Errorf("invalid log format: %v", err))
	}

	// Validate Output
	if err := validation.Validate(cfg.Output, ValidLogOutputs); err != nil {
		errs = append(errs, fmt.Errorf("invalid log output destination: %v", err))
	}

	// Conditional validation for Path
	if cfg.Output == LogOutputFile && cfg.Path == "" {
		errs = append(errs, errors.New("log path must be specified when output is 'file'"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
