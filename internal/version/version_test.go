package version

import (
	"testing"
)

// TestGet checks the Get function, simulating different states for the package-level version variables.
func TestGet(t *testing.T) {
	// Store original values of the package-level variables to restore them after the test.
	originalAppVersion := appVersion
	originalCommitHash := commitHash

	// Ensures that the original values are restored after this test function
	// (and all its subtests) have completed, regardless of panics or test failures.
	t.Cleanup(func() {
		appVersion = originalAppVersion
		commitHash = originalCommitHash
	})

	// Test cases
	tests := []struct {
		name               string // Name of the test case
		appVersion         string // Value to set for the global appVersion for this test case
		commitHash         string // Value to set for the global commitHash for this test case
		expectedAppVersion string // Expected App version
		expectedCommitHash string // Expected Commit Hash
	}{
		{
			name:               "should_return_default_values_when_globals_are_unknown",
			appVersion:         "unknown",
			commitHash:         "unknown",
			expectedAppVersion: "unknown",
			expectedCommitHash: "unknown",
		},
		{
			name:               "should_return_custom_values_when_globals_are_set",
			appVersion:         "1.2.3",
			commitHash:         "abcdef1",
			expectedAppVersion: "1.2.3",
			expectedCommitHash: "abcdef1",
		},
		{
			name:               "should_return_custom_app_version_and_default_commit_hash",
			appVersion:         "2.0.0-beta",
			commitHash:         "unknown",
			expectedAppVersion: "2.0.0-beta",
			expectedCommitHash: "unknown",
		},
		{
			name:               "should_return_default_app_version_and_custom_commit_hash",
			appVersion:         "unknown",
			commitHash:         "fedcba98",
			expectedAppVersion: "unknown",
			expectedCommitHash: "fedcba98",
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the package-level variables for this specific subtest scenario.
			appVersion = tt.appVersion
			commitHash = tt.commitHash

			got := Get()
			if got.AppVersion != tt.expectedAppVersion {
				t.Errorf("AppVersion: %+v\nExpected %q, but got: %q", appVersion, tt.expectedAppVersion, got.AppVersion)
			}
			if got.CommitHash != tt.expectedCommitHash {
				t.Errorf("CommitHash: %+v\nExpected %q, but got: %q", commitHash, tt.expectedCommitHash, got.CommitHash)
			}
		})
	}
}

// TestVersionInfo_String checks the String method of the VersionInfo struct.
func TestVersionInfo_String(t *testing.T) {
	// Test cases
	tests := []struct {
		name     string      // Name of the test case
		input    VersionInfo // The Version struct
		expected string      // The expected string output
	}{
		{
			name:     "empty_version_info_struct",
			input:    VersionInfo{},
			expected: "version  ()",
		},
		{
			name:     "standard_version_and_commit_hash",
			input:    VersionInfo{AppVersion: "1.0.0", CommitHash: "abc123xyz"},
			expected: "version 1.0.0 (abc123xyz)",
		},
		{
			name:     "version_with_prerelease_tag",
			input:    VersionInfo{AppVersion: "v2.1.3-rc1", CommitHash: "def456uvw"},
			expected: "version v2.1.3-rc1 (def456uvw)",
		},
		{
			name:     "only_app_version_present",
			input:    VersionInfo{AppVersion: "3.0", CommitHash: ""},
			expected: "version 3.0 ()",
		},
		{
			name:     "only_commit_hash_present",
			input:    VersionInfo{AppVersion: "", CommitHash: "master-branch-latest"},
			expected: "version  (master-branch-latest)",
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.String()
			if got != tt.expected {
				t.Errorf("Input: %+v\nExpected %q, but got: %q", tt.input, tt.expected, got)
			}
		})
	}
}

// TestVersionInfo_ToLogFields checks the ToLogFields method of the VersionInfo struct.
func TestVersionInfo_ToLogFields(t *testing.T) {
	// Test cases
	tests := []struct {
		name     string      // Name of the test case
		input    VersionInfo // The VersionInfo struct
		expected []any       // The expected slice of any
	}{
		{
			name: "empty_version_info_struct",
			input: VersionInfo{
				AppVersion: "",
				CommitHash: "",
			},
			expected: []any{
				"version", "",
				"commit_hash", "",
			},
		},
		{
			name: "standard_version_and_commit_hash",
			input: VersionInfo{
				AppVersion: "1.0.0",
				CommitHash: "abc123xyz",
			},
			expected: []any{
				"version", "1.0.0",
				"commit_hash", "abc123xyz",
			},
		},
		{
			name: "version_with_prerelease_tag",
			input: VersionInfo{
				AppVersion: "v2.1.3-rc1",
				CommitHash: "def456uvw",
			},
			expected: []any{
				"version", "v2.1.3-rc1",
				"commit_hash", "def456uvw",
			},
		},
		{
			name: "only_app_version_present",
			input: VersionInfo{
				AppVersion: "3.0",
				CommitHash: "",
			},
			expected: []any{
				"version", "3.0",
				"commit_hash", "",
			},
		},
		{
			name: "only_commit_hash_present",
			input: VersionInfo{
				AppVersion: "",
				CommitHash: "master-branch-latest",
			},
			expected: []any{
				"version", "",
				"commit_hash", "master-branch-latest",
			},
		},
		{
			name: "version_and_commit_with_special_characters",
			input: VersionInfo{
				AppVersion: "1.2.3-beta+special",
				CommitHash: "a!b@c#1$2%3^",
			},
			expected: []any{
				"version", "1.2.3-beta+special",
				"commit_hash", "a!b@c#1$2%3^",
			},
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.ToLogFields()

			// Check if the lengths are equal
			if len(got) != len(tt.expected) {
				t.Errorf("Input: %+v\nExpected length %d, but got length %d", tt.input, len(tt.expected), len(got))
				return // Important: Exit the test case if lengths differ
			}

			// Iterate and compare element by element
			for i := range got {
				gotStr, gotOK := got[i].(string)
				expectedStr, expectedOK := tt.expected[i].(string)

				if !gotOK || !expectedOK {
					t.Errorf("Input: %+v\nIndex %d: Expected string, but got different type", tt.input, i)
					return
				}

				if gotStr != expectedStr {
					t.Errorf("Input: %+v\nIndex %d: Expected %q, but got %q", tt.input, i, expectedStr, gotStr)
				}
			}
		})
	}
}
