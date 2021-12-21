package ioutils

import "os"

func ReadEnv(key string, defaultValue string) string {
	result, found := os.LookupEnv(key)
	if !found {
		return defaultValue
	}
	return result
}
