package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	Redis    RedisConfig
	LogLevel string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type ServerConfig struct {
	Host string
	Port int
}

type RedisConfig struct {
	URL      string
	Host     string
	Port     int
	Password string
	DB       int
	Enabled  bool
}

// LoadWithValidation loads configuration with proper error handling
func LoadWithValidation() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Parse database port with proper error handling
	dbPort, err := parseIntWithDefault("DB_PORT", "5432")
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	// Parse server port with proper error handling
	serverPort, err := parseIntWithDefault("SERVER_PORT", "8080")
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_PORT: %w", err)
	}

	// Parse Redis port with proper error handling
	redisPort, err := parseIntWithDefault("REDIS_PORT", "6379")
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_PORT: %w", err)
	}

	// Parse Redis DB with proper error handling
	redisDB, err := parseIntWithDefault("REDIS_DB", "0")
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_DB: %w", err)
	}

	// Parse Redis enabled flag with proper error handling
	redisEnabled, err := parseBoolWithDefault("REDIS_ENABLED", "true")
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_ENABLED: %w", err)
	}

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "libmngmt"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "localhost"),
			Port: serverPort,
		},
		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", ""),
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     redisPort,
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
			Enabled:  redisEnabled,
		},
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}, nil
}

// Load provides backward compatibility with improved error handling
func Load() *Config {
	config, err := LoadWithValidation()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	return config
}

func parseIntWithDefault(key, defaultValue string) (int, error) {
	valueStr := getEnv(key, defaultValue)
	return strconv.Atoi(valueStr)
}

func parseBoolWithDefault(key, defaultValue string) (bool, error) {
	valueStr := getEnv(key, defaultValue)
	return strconv.ParseBool(valueStr)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
