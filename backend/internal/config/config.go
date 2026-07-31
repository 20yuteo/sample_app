package config

import "os"

type Config struct {
	DatabaseURL string
	APIAddr     string
}

func Load() Config {
	return Config{
		DatabaseURL: env("DATABASE_URL", "postgres://commerce:commerce@localhost:5432/commerce?sslmode=disable"),
		APIAddr:     env("API_ADDR", ":8080"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
