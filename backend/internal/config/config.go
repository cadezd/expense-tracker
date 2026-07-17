package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppPort                  string
	DatabseURL               string
	UploadDirectory          string
	MaxFileUploadSizeInBytes int64
}

func Load() (Config, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	return Config{
		AppPort:                  getEnv("APP_PORT", "8080"),
		DatabseURL:               databaseURL,
		UploadDirectory:          getEnv("UPLOAD_DIR", "./data/"),
		MaxFileUploadSizeInBytes: getEnvInt64("MAX_UPLOAD_SIZE_BYTES", 10*1024*1024),
	}, nil
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}

	return val
}

func getEnvInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}

	return parsed
}
