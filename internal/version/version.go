// Package version provides information about the version of the application.
//
// It includes the app version and Git commit hash, and offers a way to retrieve
// this information in a structured format.
//
// This package aims to standardize how version information is stored and
// accessed.
package version

import "fmt"

// VersionInfo holds the version information of the application. It contains the
// app version and the corresponding Git commit hash.
type VersionInfo struct {
	AppVersion string
	CommitHash string
}

var (
	appVersion = "unknown" // The default app version, set to "unknown" if not provided by ldflags.
	commitHash = "unknown" // The default Git commit hash, set to "unknown" if not provided by ldflags.
)

// Get returns a pointer to a VersionInfo struct containing the current version
// information of the application.
func Get() *VersionInfo {
	return &VersionInfo{
		AppVersion: appVersion,
		CommitHash: commitHash,
	}
}

// String returns a formatted string representation of the version information,
// including both the app version and the Git commit hash.
func (vi VersionInfo) String() string {
	return fmt.Sprintf("version %s (%s)", vi.AppVersion, vi.CommitHash)
}

// ToLogFields returns a []any containing the version information,
// suitable for use with loggers that accept key-value pairs.
func (vi VersionInfo) ToLogFields() []any {
	return []any{
		"version", vi.AppVersion,
		"commit_hash", vi.CommitHash,
	}
}
