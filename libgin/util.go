package libgin

// Common utilities for the GIN services

import (
	"fmt"
	"net/http"
	"os"
)

// ReadConfDefault returns the value of a configuration env variable.
// If the variable is not set, the default is returned.
func ReadConfDefault(key, defval string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defval
	}
	return value
}

// ReadConf returns the value of a configuration env variable.
// If the variable is not set, an empty string is returned (ignores any errors).
func ReadConf(key string) string {
	value, _ := os.LookupEnv(key)
	return value
}

// GetArchiveSize returns the size of the archive at the given URL.
// If the URL is invalid or unreachable, an error is returned.
func GetArchiveSize(archiveURL string) (uint64, error) {
	resp, err := http.Get(archiveURL)
	if err != nil {
		return 0, fmt.Errorf("Request for archive %q failed: %s\n", archiveURL, err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Request for archive %q failed: %s\n", archiveURL, resp.Status)
	}
	if resp.ContentLength < 0 {
		// returns -1 when size is unknown; let's turn it into an error
		return 0, fmt.Errorf("Unable to determine size of %q", archiveURL)
	}
	return uint64(resp.ContentLength), nil
}
