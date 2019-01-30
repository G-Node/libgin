package libgin

// Common utilities for the GIN services

import (
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

// ReadConf returns the value of a configuration env variable and exits with an error if it is not set.
func ReadConf(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		os.Exit(-1)
	}
	return value
}
