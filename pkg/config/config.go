package config

import (
	"os"
)

// Load reads environment variables and provides a simple config loader.
func Load(key, defaultValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultValue
}
