package validation

import (
	"errors"
	"fmt"
	"net/url"
	"slices"
)

// ValidateURL checks if the given URL is valid based on valid HTTP/HTTPS schemes
// and if it has a non-empty host.
func ValidateURL(rawURL string) error {
	validSchemes := []string{"http", "https"}

	// Parse the URL using net/url
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("cannot parse URL %q: %v", rawURL, err)
	}

	// Check if the scheme is in the list of valid schemes
	if !slices.Contains(validSchemes, parsedURL.Scheme) {
		return fmt.Errorf("URL scheme %q is invalid; must be one of: %v", parsedURL.Scheme, validSchemes)
	}

	// Check if the host is not empty
	if parsedURL.Host == "" {
		return errors.New("URL must have a non-empty host")
	}

	return nil // Return nil if the URL is valid
}
