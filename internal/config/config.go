package config

import (
	"os"
	"strings"
)

type Config struct {
	ServerPort     string
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOUseSSL    bool
}

func Load() Config {
	return Config{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		MinIOEndpoint:  getEnv("MINIO_ENDPOINT", "minio:9000"),
		MinIOAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinIOUseSSL:    strings.EqualFold(getEnv("MINIO_USE_SSL", "false"), "true"),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
