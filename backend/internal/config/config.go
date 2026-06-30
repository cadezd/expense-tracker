package config

import (
	"fmt"
	"os"
)

type Config struct {
	AppPort         string
	DatabseURL      string
	UploadDirectory string
}

func Load() (Config, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	return Config{
		AppPort:         getEnv("APP_PORT", "8080"),
		DatabseURL:      databaseURL,
		UploadDirectory: getEnv("UPLOAD_DIR", "./data/uploads"),
	}, nil
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}

	return val
}
