package main

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

func getEnvOrDefault(envKey, defaultValue string) string {
	if value := os.Getenv(envKey); value != "" {
		return value
	}
	return defaultValue
}

func LoadEnv() {
	if err := godotenv.Load(".env.config"); err == nil {
		slog.Info("Loaded configuration from .env")
		return
	}

	if err := godotenv.Load(".env.example.config"); err == nil {
		slog.Info("Loaded configuration from .env.example")
		return
	}

	slog.Warn("No .env.config file found. Using default values.")
}

func resetToDefaults() error {
	if _, err := os.Stat(".env.config"); os.IsNotExist(err) {
		slog.Info("No .env file found, nothing to reset")
		return nil
	}

	if err := os.WriteFile(".env.config", []byte(""), 0600); err != nil {
		return err
	}

	return nil
}
