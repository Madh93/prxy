package validation

import (
	"strings"
	"testing"
)

// TestValidateURL tests the ValidateURL function with various inputs.
func TestValidateURL(t *testing.T) {
	// Test cases
	tests := []struct {
		name          string // Name of the subtest
		rawURL        string // Input URL string
		expectError   bool   // true if an error is expected, false otherwise
		errorContains string // A substring expected to be in the error message if expectError is true
	}{
		// Valid URL cases
		{
			name:        "valid_http_url",
			rawURL:      "http://example.com",
			expectError: false,
		},
		{
			name:        "valid_https_url",
			rawURL:      "https://example.com",
			expectError: false,
		},
		{
			name:        "valid_http_url_with_port",
			rawURL:      "http://localhost:8080",
			expectError: false,
		},
		{
			name:        "valid_https_url_with_path_query_fragment",
			rawURL:      "https://example.com/path?query=value#fragment",
			expectError: false,
		},
		{
			name:        "valid_http_url_with_ip_address_host",
			rawURL:      "http://127.0.0.1/status",
			expectError: false,
		},
		// Invalid scheme cases
		{
			name:          "invalid_scheme_ftp",
			rawURL:        "ftp://example.com",
			expectError:   true,
			errorContains: "URL scheme \"ftp\" is invalid",
		},
		{
			name:          "invalid_scheme_empty_from_relative_url",
			rawURL:        "//example.com/path",
			expectError:   true,
			errorContains: "URL scheme \"\" is invalid",
		},
		{
			name:          "invalid_scheme_empty_from_path_only_url",
			rawURL:        "/just/a/path",
			expectError:   true,
			errorContains: "URL scheme \"\" is invalid",
		},
		// Missing host cases
		{
			name:          "missing_host_for_http_scheme",
			rawURL:        "http://",
			expectError:   true,
			errorContains: "URL must have a non-empty host",
		},
		{
			name:          "missing_host_for_https_scheme",
			rawURL:        "https://",
			expectError:   true,
			errorContains: "URL must have a non-empty host",
		},
		{
			name:          "missing_host_when_path_looks_like_host",
			rawURL:        "http:example.com",
			expectError:   true,
			errorContains: "URL must have a non-empty host",
		},
		// Malformed URL / Parsing error cases
		{
			name:          "malformed_url_parse_error",
			rawURL:        "://leading.colon.com",
			expectError:   true,
			errorContains: "cannot parse URL",
		},
		{
			name:          "malformed_url_with_space_in_host",
			rawURL:        "http://exa mple.com",
			expectError:   true,
			errorContains: "cannot parse URL",
		},
		{
			name:          "empty_string_as_url",
			rawURL:        "",
			expectError:   true,
			errorContains: "URL scheme \"\" is invalid",
		},
		{
			name:          "url_with_invalid_control_character",
			rawURL:        "http://example.com/path\x00",
			expectError:   true,
			errorContains: "cannot parse URL",
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateURL(tt.rawURL)

			// The expectation about whether an error should occur was wrong.
			if (got != nil) != tt.expectError {
				if tt.expectError {
					t.Errorf("ValidateURL(%q)\nExpected error, but got: %v", tt.rawURL, got)
				} else {
					t.Errorf("ValidateURL(%q)\nExpected no error, but got: %v", tt.rawURL, got)
				}
				return // Stop further checks if error presence is not as expected.
			}

			// If an error was expected and it did occur, check its content.
			if tt.expectError && got != nil {
				if tt.errorContains == "" {
					t.Logf("Subtest %q: an error was expected, but no 'errorContains' string was specified for content checking.", tt.name)
				} else if !strings.Contains(got.Error(), tt.errorContains) {
					t.Errorf("ValidateURL(%q)\nExpected error to contain: %q, but got: %q", tt.rawURL, tt.errorContains, got.Error())
				}
			}
		})
	}
}
