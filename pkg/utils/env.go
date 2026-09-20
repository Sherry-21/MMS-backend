package utils

import (
	"os"
	"strconv"
	"strings"
	"time"
)

func GetEnv(key, defaultValue string) string {
	val, exists := os.LookupEnv(key)
	if !exists || strings.TrimSpace(val) == "" {
		return defaultValue
	}
	return strings.TrimSpace(val)
}

func GetEnvAsInt(key string, defaultValue int) int {
	valStr := GetEnv(key, "")
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue
	}
	return val
}

// GetEnvAsBool retrieves a boolean environment variable or returns defaultValue if not set/invalid.
func GetEnvAsBool(key string, defaultValue bool) bool {
	valStr := GetEnv(key, "")
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.ParseBool(valStr)
	if err != nil {
		return defaultValue
	}
	return val
}

// GetEnvAsDuration retrieves a time.Duration environment variable (e.g. "10s", "1m")
// or returns defaultValue if not set or cannot be parsed.
func GetEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valStr := GetEnv(key, "")
	if valStr == "" {
		return defaultValue
	}
	val, err := time.ParseDuration(valStr)
	if err != nil {
		return defaultValue
	}
	return val
}

// GetEnvAsSlice retrieves a comma-separated environment variable and parses it into a slice of strings.
func GetEnvAsSlice(key string, defaultValue []string, sep string) []string {
	valStr := GetEnv(key, "")
	if valStr == "" {
		return defaultValue
	}
	items := strings.Split(valStr, sep)
	var result []string
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return defaultValue
	}
	return result
}
