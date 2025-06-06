package validation

import (
	"testing"
)

// TestValidate_Integers checks the Validate function with integer inputs.
func TestValidate_Integers(t *testing.T) {
	// Test cases
	tests := []struct {
		name         string // Name of the test case
		value        int    // The value to validate
		validOptions []int  // The valid options
		expectError  bool   // true if an error is expected, false otherwise
	}{
		// Valid tests cases
		{
			name:         "value_is_present_in_options",
			value:        1,
			validOptions: []int{1, 2, 3},
			expectError:  false,
		},
		// Invalid tests cases
		{
			name:         "value_is_not_present_in_options",
			value:        4,
			validOptions: []int{1, 2, 3},
			expectError:  true,
		},
		{
			name:         "value_with_empty_options_list",
			value:        1,
			validOptions: []int{},
			expectError:  true,
		},
		{
			name:         "value_with_nil_options_list",
			value:        1,
			validOptions: nil,
			expectError:  true,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Validate(tt.value, tt.validOptions)
			if (got != nil) != tt.expectError {
				if tt.expectError {
					t.Errorf("Validate(%d, %v)\nExpected error, but got: %v", tt.value, tt.validOptions, got)
				} else {
					t.Errorf("Validate(%d, %v)\nExpected no error, but got: %v", tt.value, tt.validOptions, got)
				}
			}
		})
	}
}

// TestValidate_Strings checks the Validate function with string inputs.
func TestValidate_Strings(t *testing.T) {
	// Tests cases
	tests := []struct {
		name         string   // Name of the test case
		value        string   // The value to validate
		validOptions []string // The valid options
		expectError  bool     // true if an error is expected, false otherwise
	}{
		// Valid tests cases
		{
			name:         "value_is_present_in_options",
			value:        "apple",
			validOptions: []string{"apple", "banana", "cherry"},
			expectError:  false,
		},
		{
			name:         "empty_string_value_present_in_options",
			value:        "",
			validOptions: []string{"", "a", "b"},
			expectError:  false,
		},
		// Invalid tests cases
		{
			name:         "value_is_not_present_in_options",
			value:        "grape",
			validOptions: []string{"apple", "banana", "cherry"},
			expectError:  true,
		},
		{
			name:         "empty_string_value_not_present_in_options",
			value:        "",
			validOptions: []string{"a", "b", "c"},
			expectError:  true,
		},
		{
			name:         "value_with_empty_options_list",
			value:        "apple",
			validOptions: []string{},
			expectError:  true,
		},
		{
			name:         "value_with_nil_options_list",
			value:        "apple",
			validOptions: nil,
			expectError:  true,
		},
		{
			name:         "case_sensitive_value_not_present",
			value:        "Apple",
			validOptions: []string{"apple", "banana", "cherry"},
			expectError:  true,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Validate(tt.value, tt.validOptions)
			if (got != nil) != tt.expectError {
				if tt.expectError {
					t.Errorf("Validate(%q, %v)\nExpected error, but got: %v", tt.value, tt.validOptions, got)
				} else {
					t.Errorf("Validate(%q, %v)\nExpected no error, but got: %v", tt.value, tt.validOptions, got)
				}
			}
		})
	}
}
