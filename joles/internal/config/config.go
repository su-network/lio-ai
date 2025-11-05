package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	Server ServerConfig
	Database DatabaseConfig
	App AppConfig
}

// ServerConfig contains server configuration
type ServerConfig struct {
	Host string
	Port string
}

// DatabaseConfig contains database configuration
type DatabaseConfig struct {
	DSN string
}

// AppConfig contains application configuration
type AppConfig struct {
	Name string
	Version string
	Environment string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load environment from a single place: prefer root .env
	// Attempt in this order: ENV_FILE, .env (cwd), ../.env, ../../.env
	if envFile := os.Getenv("ENV_FILE"); envFile != "" {
		_ = godotenv.Overload(envFile)
	} else {
		// Try current directory
		_ = godotenv.Load(".env")
		// Try parent directories (repo root)
		_ = godotenv.Overload(filepath.Clean("../.env"))
		_ = godotenv.Overload(filepath.Clean("../../.env"))
	}

	config := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			// Store DB under repository root data/ directory by default
			DSN: getEnv("DATABASE_URL", "data/lio.db"),
		},
		App: AppConfig{
			Name: getEnv("APP_NAME", "Lio AI API"),
			Version: getEnv("APP_VERSION", "0.1.0"),
			Environment: getEnv("ENVIRONMENT", "development"),
		},
	}

	return config, nil
}

// getEnv retrieves environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetDSN returns the formatted database connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf("file:%s?cache=shared&mode=rwc&_journal_mode=WAL", c.Database.DSN)
}
