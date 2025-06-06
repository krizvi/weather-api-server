package utils

import (
	"fmt"
	"os"
	"strconv"
)

// GetEnvAsStrWithDefault retrieves environment variable as string, returns default value if not found
func GetEnvAsStrWithDefault(envName string, defValue string) string {
	envVal := os.Getenv(envName)
	if envVal == "" {
		return defValue
	}
	return envVal
}

// GetEnvAsMustStr retrieves required environment variable as string, returns error if not found
func GetEnvAsMustStr(envName string, errMsg string) (string, error) {
	envVal := os.Getenv(envName)
	if envVal == "" {
		return "", fmt.Errorf(errMsg)
	}
	return envVal, nil
}

// GetEnvAsIntWithDefault retrieves environment variable as integer, returns default value if not found or invalid
func GetEnvAsIntWithDefault(envName string, defValue int) int {
	envVal := os.Getenv(envName)
	if envVal == "" {
		return defValue
	}
	envValAsInt, err := strconv.Atoi(envVal)
	if err != nil {
		return defValue
	}
	return envValAsInt
}
