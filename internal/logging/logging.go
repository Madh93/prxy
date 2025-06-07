// Package logging manages application logging using the log/slog package.
//
// This package provides a structured logging facility for applications. It
// allows the creation of a Logger instance that can log messages at different
// severity levels such as Debug, Info, Warn, Error, and Fatal. The logging
// configuration is flexible and supports different output destinations (such as
// standard output or files) and formats (such as JSON or text).
//
// The Logger uses the slog package for structured logging and can be configured
// to determine the logging output and format based on user-defined settings.
//
// Use the New function to create a Logger instance with specified logging
// configuration. Various methods are provided to log messages at different
// severity levels with additional context.
package logging

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/Madh93/prxy/internal/config"
)

// nopWriteCloser wraps an io.Writer with a no-op Close method.
// This is useful for standard streams like os.Stdout and os.Stderr
// when an io.WriteCloser interface is required.
type nopWriteCloser struct {
	io.Writer
}

// Close implements the io.Closer interface for nopWriteCloser. It does nothing and returns nil.
func (nwc nopWriteCloser) Close() error { return nil }

// Logger represents an instance of the logging system.
type Logger struct {
	slogger  *slog.Logger // The slogger instance
	exitFunc func(int)    // Function to call on Fatal, defaults to os.Exit
	closer   io.Closer    // The underlying writer that might need to be closed (e.g., a file)
}

// New creates a new Logger instance with the specified logging configuration.
// It returns an error if the configuration is invalid or setup fails.
func New(cfg *config.LoggingConfig) (*Logger, error) {
	// Set the output based on the configuration
	output, err := parseOutput(cfg)
	if err != nil {
		return nil, fmt.Errorf("could not parse log output: %v", err)
	}

	// Setup the handler based on the format
	handler, err := parseFormat(output, cfg)
	if err != nil {
		var errs []error
		errs = append(errs, fmt.Errorf("could not set up log handler: %v", err))

		// Close output
		if closeErr := output.Close(); closeErr != nil {
			errs = append(errs, fmt.Errorf("failed to close output writer: %v", closeErr))
		}

		if len(errs) > 0 {
			return nil, errors.Join(errs...)
		}
	}

	return &Logger{
		slogger:  slog.New(handler),
		exitFunc: os.Exit,
		closer:   output,
	}, nil
}

// Close closes the logger's underlying output writer, if it is closable
// (e.g., a file). It should be called when the logger is no longer needed
// to release resources.
func (l *Logger) Close() error {
	if l.closer != nil {
		return l.closer.Close()
	}
	return nil
}

// Debug logs a message at the debug level.
func (l *Logger) Debug(msg string, args ...any) {
	l.slogger.Debug(msg, args...)
}

// Info logs a message at the info level.
func (l *Logger) Info(msg string, args ...any) {
	l.slogger.Info(msg, args...)
}

// Warn logs a message at the warn level.
func (l *Logger) Warn(msg string, args ...any) {
	l.slogger.Warn(msg, args...)
}

// Error logs a message at the error level.
func (l *Logger) Error(msg string, args ...any) {
	l.slogger.Error(msg, args...)
}

// Fatal logs a message at the error level and then exits the program.
// The exit behavior can be overridden for testing.
func (l *Logger) Fatal(msg string, args ...any) {
	l.slogger.Error(msg, args...)
	if l.exitFunc != nil {
		l.exitFunc(1)
	} else {
		os.Exit(1)
	}
}

// parseOutput determines the io.Writer for logging based on the configuration.
// Note: If a file is opened, the caller is responsible for closing it.
func parseOutput(cfg *config.LoggingConfig) (io.WriteCloser, error) {
	var output io.WriteCloser

	switch cfg.Output {
	case config.LogOutputStderr:
		output = nopWriteCloser{os.Stderr}
	case config.LogOutputFile:
		if cfg.Path == "" {
			return nil, errors.New("internal error: file output mode requires a non-empty path, but path is empty")
		} else {
			file, err := os.OpenFile(cfg.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
			if err != nil {
				return nil, fmt.Errorf("failed to open log file %q: %v", cfg.Path, err)
			}
			output = file
		}
	case config.LogOutputStdout:
		fallthrough
	default:
		output = nopWriteCloser{os.Stdout}
	}

	return output, nil
}

// parseLevel converts a string representation of a log level from config.LogLevel
// to slog.Level. Defaults to slog.LevelInfo on parsing failure.
func parseLevel(logLevel config.LogLevel) slog.Level {
	var level slog.Level
	if err := level.UnmarshalText([]byte(string(logLevel))); err != nil {
		return slog.LevelInfo
	}
	return level
}

// parseFormat creates an slog.Handler based on the configuration.
func parseFormat(output io.Writer, cfg *config.LoggingConfig) (slog.Handler, error) {
	var handler slog.Handler

	// Setup handler options like log level.
	options := slog.HandlerOptions{
		Level: parseLevel(cfg.Level),
	}

	switch cfg.Format {
	case config.LogFormatJSON:
		handler = slog.NewJSONHandler(output, &options)
	case config.LogFormatText:
		fallthrough
	default:
		handler = slog.NewTextHandler(output, &options)
	}

	return handler, nil
}
