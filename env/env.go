package env

import (
	"os"
	"strconv"
)

func GetAsString(key string, defaultValues ...string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	if len(defaultValues) > 0 {
		return defaultValues[0]
	}
	return ""
}

func GetAsInt(key string, defaultValues ...int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	if len(defaultValues) > 0 {
		return defaultValues[0]
	}
	return 0
}

func GetAsBool(key string, defaultValues ...bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	if len(defaultValues) > 0 {
		return defaultValues[0]
	}
	return false
}
