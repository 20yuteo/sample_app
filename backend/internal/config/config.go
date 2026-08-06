package config

import (
	"os"
	"strings"
)

type Config struct {
	DatabaseURL        string
	APIAddr            string
	CORSAllowedOrigins []string
}

func Load() Config {
	return Config{
		DatabaseURL:        env("DATABASE_URL", "postgres://commerce:commerce@localhost:5432/commerce?sslmode=disable"),
		APIAddr:            env("API_ADDR", ":8080"),
		CORSAllowedOrigins: csvEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func csvEnv(key, fallback string) []string {
	raw := env(key, fallback)
	values := strings.Split(raw, ",")
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}
