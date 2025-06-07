// logging_test.go
package logging

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/Madh93/prxy/internal/config"
)

// newTestDefaultConfig is a helper to create a default valid config.LoggingConfig for tests.
// Assumes config package defines these constants/types (e.g., config.InfoLevel = "info").
func newTestDefaultConfig() *config.LoggingConfig {
	return &config.LoggingConfig{
		Level:  config.LogLevelInfo,
		Format: config.LogFormatText,
		Output: config.LogOutputStdout,
		Path:   "",
	}
}

// TestNew checks the Logging creation.
func TestNew(t *testing.T) {
	t.Run("should_create_logger_with_default_config", func(t *testing.T) {
		cfg := newTestDefaultConfig() // Uses StdoutOutput and TextFormat by default from helper

		logger, err := New(cfg)
		if err != nil {
			t.Fatalf("New() with default config failed: %v", err)
		}
		if logger == nil {
			t.Fatal("New() returned nil logger for default config")
		}
		if logger.slogger == nil {
			t.Fatal("New() logger.slogger is nil")
		}
	})

	t.Run("should_create_logger_for_file_output", func(t *testing.T) {
		tempDir := t.TempDir()
		logFilePath := filepath.Join(tempDir, "test_new.log")

		cfg := newTestDefaultConfig()
		cfg.Output = config.LogOutputFile
		cfg.Path = logFilePath
		cfg.Format = config.LogFormatJSON // Explicitly JSON to test this path

		logger, err := New(cfg)
		if err != nil {
			t.Fatalf("New() with file output config failed: %v", err)
		}
		if logger == nil {
			t.Fatal("New() returned nil logger for file output")
		}
		defer func() {
			if err := logger.Close(); err != nil {
				t.Fatalf("Failed to close log file %q: %v", logFilePath, err)
			}
		}()

		logger.Info("test message for file output")

		content, readErr := os.ReadFile(logFilePath)
		if readErr != nil {
			t.Fatalf("Failed to read log file %q: %v", logFilePath, readErr)
		}
		if !strings.Contains(string(content), "test message for file output") {
			t.Errorf("Log file content does not contain expected message. Got: %s", string(content))
		}
	})

	t.Run("should_fail_if_parseOutput_fails_due_to_empty_path_for_file", func(t *testing.T) {
		cfg := newTestDefaultConfig()
		cfg.Output = config.LogOutputFile
		cfg.Path = "" // Invalid: empty path for file output

		_, err := New(cfg)
		if err == nil {
			t.Fatal("New() succeeded with empty path for file output, but expected error")
		}
		if !strings.Contains(err.Error(), "could not parse log output") || !strings.Contains(err.Error(), "internal error: file output mode requires a non-empty path") {
			t.Errorf("Expected error about parsing output or internal error for empty path, got: %v", err)
		}
	})

	t.Run("should_fail_if_parseOutput_fails_to_open_file", func(t *testing.T) {
		cfg := newTestDefaultConfig()
		cfg.Output = config.LogOutputFile
		cfg.Path = "/this/path/should/not/exist/or/be/writable/test.log" // Invalid path

		_, err := New(cfg)
		if err == nil {
			t.Fatal("New() with invalid file path succeeded, but expected error")
		}
		if !strings.Contains(err.Error(), "could not parse log output") || !strings.Contains(err.Error(), "failed to open log file") {
			t.Errorf("Expected error about failing to open log file, got: %v", err)
		}
	})
}

// TestLogger_Close checks if the Close method correctly closes the file writer.
func TestLogger_Close(t *testing.T) {
	t.Run("should_close_file_writer_successfully", func(t *testing.T) {
		tempDir := t.TempDir()
		logFilePath := filepath.Join(tempDir, "test_close.log")
		cfg := &config.LoggingConfig{Output: config.LogOutputFile, Path: logFilePath, Level: config.LogLevelInfo, Format: config.LogFormatText}

		logger, err := New(cfg)
		if err != nil {
			t.Fatalf("Failed to create logger for TestLogger_Close: %v", err)
		}

		err = logger.Close()
		if err != nil {
			t.Errorf("logger.Close() failed for file writer: %v", err)
		}

		// Try to close again (for *os.File, this should error, testing robustness of app logic around it)
		// However, our Close() method doesn't prevent this internally. The os.File.Close() will error.
		// This tests that the error from the underlying closer is propagated.
		err = logger.Close()
		if err == nil {
			t.Error("logger.Close() on an already closed file writer did not return an error")
		}
	})

	t.Run("should_return_nil_when_closing_nopWriteCloser_for_stdout", func(t *testing.T) {
		cfg := &config.LoggingConfig{Output: "some_default_that_goes_to_stdout", Level: config.LogLevelInfo, Format: config.LogFormatText}
		logger, err := New(cfg) // This will use nopWriteCloser{os.Stdout} due to default in parseOutput
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		err = logger.Close()
		if err != nil {
			t.Errorf("logger.Close() failed for NopWriteCloser (stdout): %v", err)
		}
	})

	t.Run("should_return_nil_when_closer_is_nil_in_logger_struct", func(t *testing.T) {
		// This case is unlikely if New always initializes closer, but tests defensiveness.
		logger := &Logger{closer: nil}
		err := logger.Close()
		if err != nil {
			t.Errorf("logger.Close() with nil closer failed: %v", err)
		}
	})
}

// TestParseOutput checks the Logging Output configuration.
func TestParseOutput(t *testing.T) {
	// Test cases
	tests := []struct {
		name          string                // Name of the test case
		cfg           *config.LoggingConfig // The Logging configuration
		expectError   bool                  // true if an error is expected, false otherwise
		errorContains string                // A substring expected to be in the error message if expectError is true
		expectedOut   io.Writer             // Specific os.File instance for stdout/stderr to compare against
	}{
		{
			name:        "output_stdout_should_use_stdout",
			cfg:         &config.LoggingConfig{Output: config.LogOutputStdout},
			expectError: false,
			expectedOut: os.Stdout,
		},
		{
			name:        "output_stderr_should_use_stderr",
			cfg:         &config.LoggingConfig{Output: config.LogOutputStderr},
			expectError: false,
			expectedOut: os.Stderr,
		},
		{
			name: "output_file_with_valid_path",
			cfg: &config.LoggingConfig{
				Output: config.LogOutputFile,
				Path:   filepath.Join(t.TempDir(), "test_parseoutput.log"),
			},
			expectError: false,
			// expectedOut cannot be directly os.File as it's a new file. Check type and path.
		},
		{
			name:          "output_file_with_empty_path_should_error",
			cfg:           &config.LoggingConfig{Output: config.LogOutputFile, Path: ""},
			expectError:   true,
			errorContains: "internal error: file output mode requires a non-empty path",
		},
		{
			name: "output_file_with_unwritable_path_should_error",
			cfg: &config.LoggingConfig{
				Output: config.LogOutputFile,
				Path:   "/this/path/is/likely/unwritable/test.log",
			},
			expectError:   true,
			errorContains: "failed to open log file",
		},
		{
			name:        "output_default_should_use_stdout",
			cfg:         &config.LoggingConfig{Output: "some_other_value"}, // Default case
			expectError: false,
			expectedOut: os.Stdout,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer, err := parseOutput(tt.cfg)

			hasError := (err != nil)
			if hasError != tt.expectError {
				t.Fatalf("parseOutput() error = %v, expectError %v", err, tt.expectError)
			}

			if tt.expectError {
				if err != nil && tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("parseOutput() error = %q, want error to contain %q", err.Error(), tt.errorContains)
				}
				return // Stop if error was expected and occurred (or not as expected)
			}

			// If no error expected, check the writer type
			if fileCase := strings.Contains(tt.name, "output_file_with_valid_path"); fileCase {
				f, ok := writer.(*os.File)
				if !ok {
					t.Fatalf("Expected *os.File for file output, got %T", writer)
				}
				if f.Name() != tt.cfg.Path {
					t.Errorf("Expected file path %q, got %q", tt.cfg.Path, f.Name())
				}
				// Clean up by closing the file; parseOutput doesn't close, the caller (New->Logger) does.
				// Here, the test itself needs to close what parseOutput returned for this test case.
				closeErr := writer.Close()
				if closeErr != nil {
					t.Errorf("Failed to close file opened by parseOutput: %v", closeErr)
				}
			} else {
				// For stdout/stderr, check if it's a nopWriteCloser wrapping the expected stream
				nwc, ok := writer.(nopWriteCloser)
				if !ok {
					t.Fatalf("Expected nopWriteCloser for stdout/stderr, got %T", writer)
				}
				if nwc.Writer != tt.expectedOut {
					t.Errorf("Expected nopWriteCloser to wrap %v, but wrapped %v", tt.expectedOut, nwc.Writer)
				}
			}
		})
	}
}

// TestParseLevel checks the Logging Level configuration.
func TestParseLevel(t *testing.T) {
	// Tests cases
	tests := []struct {
		name            string          // Name of the subtest
		inputConfigLvl  config.LogLevel // The providedlogging level
		expectedSlogLvl slog.Level      // The expected logging level
	}{
		{"level_debug", config.LogLevel("debug"), slog.LevelDebug},
		{"level_info", config.LogLevel("info"), slog.LevelInfo},
		{"level_warn", config.LogLevel("warn"), slog.LevelWarn},
		{"level_error", config.LogLevel("error"), slog.LevelError},
		{"level_invalid_string", config.LogLevel("verbose"), slog.LevelInfo}, // Default
		{"level_empty_string", config.LogLevel(""), slog.LevelInfo},          // Default
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLevel(tt.inputConfigLvl)
			if got != tt.expectedSlogLvl {
				t.Errorf("parseLevel(%q) = %s, want %s", tt.inputConfigLvl, got, tt.expectedSlogLvl)
			}
		})
	}
}

// TestParseFormat checks the Logging Format configuration.
func TestParseFormat(t *testing.T) {
	var buf bytes.Buffer // Dummy writer for handler creation

	// Tests cases
	tests := []struct {
		name          string                // Name of the subtest
		cfg           *config.LoggingConfig // The Logging configuration
		expectJSON    bool                  // true if JSONHandler is expected
		expectHandler slog.Leveler          // For checking the level
	}{
		{
			name:          "format_text_should_be_text",
			cfg:           &config.LoggingConfig{Format: config.LogFormatText, Level: config.LogLevel("info")},
			expectJSON:    false,
			expectHandler: slog.LevelInfo,
		},
		{
			name:          "format_json_should_be_json",
			cfg:           &config.LoggingConfig{Format: config.LogFormatJSON, Level: config.LogLevel("debug")},
			expectJSON:    true,
			expectHandler: slog.LevelDebug,
		},
		{
			name:          "format_default_should_be_text_handler", // slog.TextHandler
			cfg:           &config.LoggingConfig{Format: "other", Level: config.LogLevel("warn")},
			expectJSON:    false,
			expectHandler: slog.LevelWarn,
		},
	}

	// Run Tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := parseFormat(&buf, tt.cfg)
			if err != nil {
				t.Fatalf("parseFormat() failed: %v", err)
			}

			if tt.expectJSON {
				if _, ok := handler.(*slog.JSONHandler); !ok {
					t.Errorf("Expected *slog.JSONHandler, got %T", handler)
				}
			} else {
				if _, ok := handler.(*slog.TextHandler); !ok {
					t.Errorf("Expected *slog.TextHandler, got %T", handler)
				}
			}

			// Check if the handler has the correct level set
			if !handler.Enabled(context.TODO(), tt.expectHandler.(slog.Level)) {
				t.Errorf("Handler not enabled for expected level %v", tt.expectHandler)
			}
			// Check that it's not enabled for a level below the set level (if not min level)
			if tt.expectHandler.(slog.Level) > slog.LevelDebug { // Arbitrary lowest level for this check
				if handler.Enabled(context.TODO(), tt.expectHandler.(slog.Level)-1) {
					t.Errorf("Handler unexpectedly enabled for level below %v", tt.expectHandler)
				}
			}
		})
	}
}

// TestLogger_LoggingMethods checks individual logger methods (Debug, Info, Warn, Error).
func TestLogger_LoggingMethods(t *testing.T) {
	// Tests cases
	tests := []struct {
		name          string
		logMethod     func(l *Logger, msg string, args ...any)
		handlerLevel  slog.Level       // Level to set the test handler to
		logEntryLevel slog.Level       // Level of the log entry being made
		expectedMsg   string           // The expected message
		args          []any            // Arbitrary args
		format        config.LogFormat // "json" or "text" (which will become slog.TextHandler)
		shouldLog     bool             // Whether the message is expected to be logged based on handlerLevel
	}{
		// JSON Handler
		{"json_info_level_log_info", (*Logger).Info, slog.LevelInfo, slog.LevelInfo, "info test", []any{"key", "val"}, config.LogFormatJSON, true},
		{"json_info_level_log_debug", (*Logger).Debug, slog.LevelInfo, slog.LevelDebug, "debug test", []any{"key", "val"}, config.LogFormatJSON, false}, // Debug < Info
		{"json_debug_level_log_debug", (*Logger).Debug, slog.LevelDebug, slog.LevelDebug, "debug test", []any{"key", "val"}, config.LogFormatJSON, true},
		{"json_warn_level_log_warn", (*Logger).Warn, slog.LevelWarn, slog.LevelWarn, "warn test", []any{"key", "val"}, config.LogFormatJSON, true},
		{"json_error_level_log_error", (*Logger).Error, slog.LevelError, slog.LevelError, "error test", []any{"key", "val"}, config.LogFormatJSON, true},
		// Text Handler
		{"text_info_level_log_info", (*Logger).Info, slog.LevelInfo, slog.LevelInfo, "info test", []any{"key", "val"}, "default_text", true},
		{"text_info_level_log_debug", (*Logger).Debug, slog.LevelInfo, slog.LevelDebug, "debug test", []any{"key", "val"}, "default_text", false},
		{"text_debug_level_log_debug", (*Logger).Debug, slog.LevelDebug, slog.LevelDebug, "debug test", []any{"key", "val"}, "default_text", true},
		{"text_warn_level_log_warn", (*Logger).Warn, slog.LevelWarn, slog.LevelWarn, "warn test", []any{"key", "val"}, "default_text", true},
		{"text_error_level_log_error", (*Logger).Error, slog.LevelError, slog.LevelError, "error test", []any{"key", "val"}, "default_text", true},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			var handler slog.Handler

			opts := &slog.HandlerOptions{Level: tt.handlerLevel}
			if tt.format == config.LogFormatJSON {
				handler = slog.NewJSONHandler(&buf, opts)
			} else { // Default to slog.TextHandler
				handler = slog.NewTextHandler(&buf, opts)
			}

			logger := &Logger{slogger: slog.New(handler), exitFunc: func(int) {}} // Mock exit

			tt.logMethod(logger, tt.expectedMsg, tt.args...)
			output := buf.String()

			if tt.shouldLog {
				if output == "" {
					t.Fatalf("Expected log output for %q with level %s, but got empty string", tt.expectedMsg, tt.handlerLevel)
				}
				if !strings.Contains(output, tt.expectedMsg) {
					t.Errorf("Output %q does not contain expected message %q", output, tt.expectedMsg)
				}
				if len(tt.args) > 0 {
					// Crude check for args, specific checks depend on Text vs JSON output structure
					if tt.format == config.LogFormatJSON {
						if !strings.Contains(output, fmt.Sprintf("%q:%q", tt.args[0], tt.args[1])) && // "key":"val"
							!strings.Contains(output, fmt.Sprintf("%q:%v", tt.args[0], tt.args[1])) { // "key":val (for numbers)
							t.Errorf("JSON output %q does not seem to contain args %v", output, tt.args)
						}
					} else { // TextFormat
						if !strings.Contains(output, fmt.Sprintf("%s=%s", tt.args[0], tt.args[1])) && // key=val (if val is simple string)
							!strings.Contains(output, fmt.Sprintf("%s=%q", tt.args[0], tt.args[1])) { // key="val" (if val has spaces or is quoted)
							t.Errorf("Text output %q does not seem to contain args %v", output, tt.args)
						}
					}
				}

				// Check for level string (case-sensitive for JSON, slog.TextHandler also uses uppercase)
				var expectedLevelStr string
				switch tt.logEntryLevel {
				case slog.LevelDebug:
					expectedLevelStr = "DEBUG"
				case slog.LevelInfo:
					expectedLevelStr = "INFO"
				case slog.LevelWarn:
					expectedLevelStr = "WARN"
				case slog.LevelError:
					expectedLevelStr = "ERROR"
				}
				if !strings.Contains(output, "level="+expectedLevelStr) && !strings.Contains(output, "\"level\":\""+expectedLevelStr+"\"") {
					t.Errorf("Output %q does not contain expected level string for %s", output, expectedLevelStr)
				}

			} else { // Message should not have been logged
				if output != "" {
					t.Errorf("Expected no log output for %q with handler level %s and entry level %s, but got: %s", tt.expectedMsg, tt.handlerLevel, tt.logEntryLevel, output)
				}
			}
		})
	}
}

// TestLogger_Fatal checks the Fatal logger method.
func TestLogger_Fatal(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError})

	var exitCode = -1
	var exited = false
	var exitMutex sync.Mutex

	mockExitFunc := func(code int) {
		exitMutex.Lock()
		defer exitMutex.Unlock()
		exitCode = code
		exited = true
	}

	logger := &Logger{
		slogger:  slog.New(handler),
		exitFunc: mockExitFunc,
		closer:   nopWriteCloser{&buf}, // Dummy closer
	}

	testMsg := "critical failure happened"
	args := []any{"code", 123, "component", "test"}
	logger.Fatal(testMsg, args...)

	output := buf.String()

	// Check that the message was logged at Error level
	if !strings.Contains(output, "level=ERROR") {
		t.Errorf("Fatal log output %q does not contain 'level=ERROR'", output)
	}
	if !strings.Contains(output, testMsg) {
		t.Errorf("Fatal log output %q does not contain message %q", output, testMsg)
	}
	if !strings.Contains(output, "code=123") || !strings.Contains(output, "component=test") {
		t.Errorf("Fatal log output %q does not contain all args", output)
	}

	// Check that the exit function was called
	exitMutex.Lock()
	finalExited := exited
	finalExitCode := exitCode
	exitMutex.Unlock()

	if !finalExited {
		t.Error("logger.Fatal() did not call the exit function")
	}
	if finalExitCode != 1 {
		t.Errorf("logger.Fatal() called exit function with code %d, want 1", finalExitCode)
	}
}
